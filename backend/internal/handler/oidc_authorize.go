package handler

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/token/jwt"
	"github.com/ys-1052/yossid/backend/internal/config"
	"github.com/ys-1052/yossid/backend/internal/security"
	"github.com/ys-1052/yossid/backend/internal/service"
	"github.com/ys-1052/yossid/backend/internal/storage/postgres"
	"github.com/ys-1052/yossid/backend/internal/storage/postgres/db"
)

type OIDCHandler struct {
	cfg          *config.Config
	oauth2       fosite.OAuth2Provider
	pgDB         *postgres.DB
	loginService service.LoginService
	userRepo     service.RegistrationService // we can use userRepository directly
}

func NewOIDCHandler(cfg *config.Config, oauth2 fosite.OAuth2Provider, pgDB *postgres.DB, loginService service.LoginService) *OIDCHandler {
	return &OIDCHandler{
		cfg:          cfg,
		oauth2:       oauth2,
		pgDB:         pgDB,
		loginService: loginService,
	}
}

func (h *OIDCHandler) GetAuthorize(c echo.Context) error {
	ctx := c.Request().Context()

	// 1. Check if we are resuming via request_id
	requestID := c.QueryParam("request_id")
	if requestID != "" {
		return h.handleResumedAuthorize(c, requestID)
	}

	// 2. Parse authorization request with Ory Fosite
	requester, err := h.oauth2.NewAuthorizeRequest(ctx, c.Request())
	if err != nil {
		h.oauth2.WriteAuthorizeError(ctx, c.Response(), requester, err)
		return nil
	}

	// 3. Check user session via op_session cookie
	cookie, err := c.Cookie("op_session")
	var sessionRecord *db.LoginSession
	if err == nil && cookie.Value != "" {
		sessionRecord, _ = h.loginService.GetSession(ctx, cookie.Value)
	}

	// 4. If not authenticated, save state and redirect to frontend login
	if sessionRecord == nil {
		return h.handleUnauthenticatedAuthorize(c, requester)
	}

	// 5. If authenticated, process authorization (consent skipped for MVP)
	return h.handleAuthenticatedAuthorize(c, requester, sessionRecord)
}

func (h *OIDCHandler) handleUnauthenticatedAuthorize(c echo.Context, requester fosite.AuthorizeRequester) error {
	ctx := c.Request().Context()

	// Look up client DB UUID
	clientRecord, err := h.pgDB.Queries.GetClientByClientID(ctx, requester.GetClient().GetID())
	if err != nil {
		h.oauth2.WriteAuthorizeError(ctx, c.Response(), requester, err)
		return nil
	}

	// Generate request ID
	reqID, err := security.GenerateRandomToken()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate request token")
	}
	reqIDHash := security.HashWithPepper(reqID, h.cfg.TokenPepper)

	// Save authorization request state
	scope := strings.Join(requester.GetRequestedScopes(), " ")
	redirectURI := requester.GetRequestForm().Get("redirect_uri")
	nonce := requester.GetRequestForm().Get("nonce")
	codeChallenge := requester.GetRequestForm().Get("code_challenge")
	codeChallengeMethod := requester.GetRequestForm().Get("code_challenge_method")

	_, err = h.pgDB.Queries.CreateAuthorizationRequest(ctx, db.CreateAuthorizationRequestParams{
		ID:                  uuid.New(),
		RequestIDHash:       reqIDHash,
		ClientID:            clientRecord.ID,
		RedirectUri:         redirectURI,
		Scope:               scope,
		State:               requester.GetState(),
		Nonce:               nonce,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		ExpiresAt:           time.Now().Add(10 * time.Minute), // 10 minutes expiry
	})
	if err != nil {
		h.oauth2.WriteAuthorizeError(ctx, c.Response(), requester, err)
		return nil
	}

	// Redirect to frontend login screen
	loginURL := "/login?request_id=" + reqID
	return c.Redirect(http.StatusFound, loginURL)
}

func (h *OIDCHandler) handleAuthenticatedAuthorize(c echo.Context, requester fosite.AuthorizeRequester, sessionRecord *db.LoginSession) error {
	ctx := c.Request().Context()

	// Load user details to populate subject and claims
	userRecord, err := h.pgDB.Queries.GetUserByID(ctx, sessionRecord.UserID)
	if err != nil {
		h.oauth2.WriteAuthorizeError(ctx, c.Response(), requester, err)
		return nil
	}

	// Build OIDC Session
	session := &openid.DefaultSession{
		Claims: &jwt.IDTokenClaims{
			Subject:     userRecord.Sub,
			Issuer:      h.cfg.Issuer,
			Audience:    []string{requester.GetClient().GetID()},
			ExpiresAt:   time.Now().Add(15 * time.Minute),
			IssuedAt:    time.Now(),
			AuthTime:    sessionRecord.AuthTime,
			RequestedAt: time.Now(),
		},
		Headers: &jwt.Headers{},
	}

	// Generate authorize response containing the authorization code
	response, err := h.oauth2.NewAuthorizeResponse(ctx, requester, session)
	if err != nil {
		h.oauth2.WriteAuthorizeError(ctx, c.Response(), requester, err)
		return nil
	}

	// Redirect user back to the client redirect URI
	h.oauth2.WriteAuthorizeResponse(ctx, c.Response(), requester, response)
	return nil
}

func (h *OIDCHandler) handleResumedAuthorize(c echo.Context, requestID string) error {
	ctx := c.Request().Context()

	// Hash requestID and look up authorization request
	reqIDHash := security.HashWithPepper(requestID, h.cfg.TokenPepper)
	authReq, err := h.pgDB.Queries.GetAuthorizationRequest(ctx, reqIDHash)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Authorization request not found")
	}

	if !authReq.UserID.Valid {
		// Not yet completed login
		return c.Redirect(http.StatusFound, "/login?request_id="+requestID)
	}

	if time.Now().After(authReq.ExpiresAt) {
		return echo.NewHTTPError(http.StatusGone, "Authorization request expired")
	}

	// Retrieve client
	clientRecord, err := h.pgDB.Queries.GetClientByID(ctx, authReq.ClientID)
	if err != nil {
		return err
	}

	fositeClient, err := h.oauth2.(fosite.ClientManager).GetClient(ctx, clientRecord.ClientID)
	if err != nil {
		return err
	}

	// Reconstruct fosite requester
	requester := fosite.NewAuthorizeRequest()
	requester.Client = fositeClient
	requester.RequestedScope = fosite.Arguments(strings.Split(authReq.Scope, " "))
	requester.GrantedScope = fosite.Arguments(strings.Split(authReq.Scope, " "))
	requester.State = authReq.State

	form := c.Request().Form
	if form == nil {
		form = make(url.Values)
	}
	form.Set("redirect_uri", authReq.RedirectUri)
	form.Set("nonce", authReq.Nonce)
	form.Set("code_challenge", authReq.CodeChallenge)
	form.Set("code_challenge_method", authReq.CodeChallengeMethod)
	requester.Form = form

	// Load user details
	userRecord, err := h.pgDB.Queries.GetUserByID(ctx, authReq.UserID.UUID)
	if err != nil {
		return err
	}

	// Fetch login session details
	var authTime time.Time
	var latestSession db.LoginSession
	err = h.pgDB.DB.QueryRowContext(ctx, "SELECT auth_time FROM login_sessions WHERE user_id = $1 ORDER BY auth_time DESC LIMIT 1", userRecord.ID).Scan(&authTime)
	if err != nil {
		latestSession.AuthTime = time.Now()
	} else {
		latestSession.AuthTime = authTime
	}

	// Build OIDC Session
	session := &openid.DefaultSession{
		Claims: &jwt.IDTokenClaims{
			Subject:     userRecord.Sub,
			Issuer:      h.cfg.Issuer,
			Audience:    []string{clientRecord.ClientID},
			ExpiresAt:   time.Now().Add(15 * time.Minute),
			IssuedAt:    time.Now(),
			AuthTime:    latestSession.AuthTime,
			RequestedAt: time.Now(),
		},
		Headers: &jwt.Headers{},
	}

	response, err := h.oauth2.NewAuthorizeResponse(ctx, requester, session)
	if err != nil {
		h.oauth2.WriteAuthorizeError(ctx, c.Response(), requester, err)
		return nil
	}

	h.oauth2.WriteAuthorizeResponse(ctx, c.Response(), requester, response)
	return nil
}
