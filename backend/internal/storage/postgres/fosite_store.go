package postgres

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/handler/pkce"
	"github.com/ys-1052/yossid/backend/internal/config"
	"github.com/ys-1052/yossid/backend/internal/security"
	"github.com/ys-1052/yossid/backend/internal/storage/postgres/db"
)

type FositeStore interface {
	fosite.Storage
	oauth2.AuthorizeCodeStorage
	oauth2.TokenRevocationStorage
	pkce.PKCERequestStorage
	openid.OpenIDConnectRequestStorage
}

type fositeStore struct {
	pgDB *DB
	cfg  *config.Config
}

func NewFositeStore(pgDB *DB, cfg *config.Config) FositeStore {
	return &fositeStore{
		pgDB: pgDB,
		cfg:  cfg,
	}
}

// Client Storage / fosite.ClientManager methods
func (s *fositeStore) GetClient(ctx context.Context, id string) (fosite.Client, error) {
	client, err := s.pgDB.Queries.GetClientByClientID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fosite.ErrNotFound
		}
		return nil, err
	}

	redirectURIs, err := s.pgDB.Queries.GetClientRedirectURIs(ctx, client.ID)
	if err != nil {
		return nil, err
	}

	uris := make([]string, len(redirectURIs))
	for i, r := range redirectURIs {
		uris[i] = r.RedirectUri
	}

	scopes, err := s.pgDB.Queries.GetClientAllowedScopes(ctx, client.ID)
	if err != nil {
		return nil, err
	}

	allowedScopes := make([]string, len(scopes))
	for i, sc := range scopes {
		allowedScopes[i] = sc
	}

	return &fosite.DefaultClient{
		ID:            client.ClientID,
		Secret:        []byte(client.ClientSecretHash.String),
		RedirectURIs:  uris,
		GrantTypes:    []string{"authorization_code", "refresh_token"},
		ResponseTypes: []string{"code", "id_token"},
		Scopes:        allowedScopes,
		Public:        client.ClientType == "public",
	}, nil
}

func (s *fositeStore) ClientAssertionJWTValid(ctx context.Context, jti string) error {
	// No-op for MVP (client application uses client_secret_basic / client_secret_post)
	return nil
}

func (s *fositeStore) SetClientAssertionJWT(ctx context.Context, jti string, exp time.Time) error {
	// No-op for MVP
	return nil
}

// Authorize Code Storage methods
func (s *fositeStore) CreateAuthorizeCodeSession(ctx context.Context, code string, requester fosite.Requester) error {
	codeHash := security.HashWithPepper(code, s.cfg.TokenPepper)

	client, err := s.pgDB.Queries.GetClientByClientID(ctx, requester.GetClient().GetID())
	if err != nil {
		return err
	}

	sub := requester.GetSession().GetSubject()
	var user db.User
	if sub != "" {
		user, err = s.pgDB.Queries.GetUserBySub(ctx, sub)
		if err != nil {
			return err
		}
	} else {
		return errors.New("fosite session subject is empty")
	}

	redirectURI := requester.GetRequestForm().Get("redirect_uri")
	scope := strings.Join(requester.GetGrantedScopes(), " ")
	nonce := requester.GetRequestForm().Get("nonce")
	codeChallenge := requester.GetRequestForm().Get("code_challenge")
	codeChallengeMethod := requester.GetRequestForm().Get("code_challenge_method")

	authTime := time.Now()
	if oidcSession, ok := requester.GetSession().(*openid.DefaultSession); ok && oidcSession.Claims != nil {
		authTime = oidcSession.Claims.AuthTime
	}

	expiresAt := requester.GetSession().GetExpiresAt(fosite.AuthorizeCode)

	_, err = s.pgDB.Queries.CreateAuthorizationCode(ctx, db.CreateAuthorizationCodeParams{
		ID:                  uuid.New(),
		CodeHash:            codeHash,
		ClientID:            client.ID,
		UserID:              user.ID,
		RedirectUri:         redirectURI,
		Scope:               scope,
		Nonce:               nonce,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		AuthTime:            authTime,
		ExpiresAt:           expiresAt,
	})
	return err
}

func (s *fositeStore) GetAuthorizeCodeSession(ctx context.Context, code string, session fosite.Session) (fosite.Requester, error) {
	codeHash := security.HashWithPepper(code, s.cfg.TokenPepper)

	authCode, err := s.pgDB.Queries.GetAuthorizationCode(ctx, codeHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fosite.ErrNotFound
		}
		return nil, err
	}

	if authCode.UsedAt.Valid {
		return nil, fosite.ErrInvalidGrant.WithDescription("Authorization code already used")
	}

	if time.Now().After(authCode.ExpiresAt) {
		return nil, fosite.ErrTokenExpired.WithDescription("Authorization code expired")
	}

	client, err := s.pgDB.Queries.GetClientByID(ctx, authCode.ClientID)
	if err != nil {
		return nil, err
	}

	user, err := s.pgDB.Queries.GetUserByID(ctx, authCode.UserID)
	if err != nil {
		return nil, err
	}

	fositeClient, err := s.GetClient(ctx, client.ClientID)
	if err != nil {
		return nil, err
	}

	req := fosite.NewRequest()
	req.Client = fositeClient
	req.GrantedScope = fosite.Arguments(strings.Split(authCode.Scope, " "))
	req.RequestedScope = fosite.Arguments(strings.Split(authCode.Scope, " "))

	form := url.Values{}
	form.Set("redirect_uri", authCode.RedirectUri)
	form.Set("nonce", authCode.Nonce)
	form.Set("code_challenge", authCode.CodeChallenge)
	form.Set("code_challenge_method", authCode.CodeChallengeMethod)
	req.Form = form

	session.SetExpiresAt(fosite.AuthorizeCode, authCode.ExpiresAt)

	if oidcSession, ok := session.(*openid.DefaultSession); ok {
		oidcSession.Subject = user.Sub
		if oidcSession.Claims != nil {
			oidcSession.Claims.Subject = user.Sub
			oidcSession.Claims.AuthTime = authCode.AuthTime
		}
	}

	req.Session = session

	return req, nil
}

func (s *fositeStore) InvalidateAuthorizeCodeSession(ctx context.Context, code string) error {
	codeHash := security.HashWithPepper(code, s.cfg.TokenPepper)
	_, err := s.pgDB.Queries.UseAuthorizationCode(ctx, codeHash)
	return err
}

// Refresh Token Storage methods
func (s *fositeStore) CreateRefreshTokenSession(ctx context.Context, signature string, accessSignature string, requester fosite.Requester) error {
	tokenHash := security.HashWithPepper(signature, s.cfg.TokenPepper)

	client, err := s.pgDB.Queries.GetClientByClientID(ctx, requester.GetClient().GetID())
	if err != nil {
		return err
	}

	sub := requester.GetSession().GetSubject()
	user, err := s.pgDB.Queries.GetUserBySub(ctx, sub)
	if err != nil {
		return err
	}

	scope := strings.Join(requester.GetGrantedScopes(), " ")
	expiresAt := requester.GetSession().GetExpiresAt(fosite.RefreshToken)

	// Fetch token family ID and parent ID from OIDC session claims
	var tokenFamilyID uuid.UUID
	var parentID uuid.NullUUID

	if oidcSession, ok := requester.GetSession().(*openid.DefaultSession); ok && oidcSession.Claims != nil && oidcSession.Claims.Extra != nil {
		if fid, ok := oidcSession.Claims.Extra["token_family_id"].(string); ok && fid != "" {
			tokenFamilyID = uuid.MustParse(fid)
		}
		if pid, ok := oidcSession.Claims.Extra["parent_id"].(string); ok && pid != "" {
			parentID = uuid.NullUUID{UUID: uuid.MustParse(pid), Valid: true}
		}
	}

	if tokenFamilyID == uuid.Nil {
		tokenFamilyID = uuid.New()
	}

	_, err = s.pgDB.Queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		ID:            uuid.New(),
		TokenHash:     tokenHash,
		TokenFamilyID: tokenFamilyID,
		ClientID:      client.ID,
		UserID:        user.ID,
		Scope:         scope,
		ParentID:      parentID,
		ExpiresAt:     expiresAt,
	})
	return err
}

func (s *fositeStore) GetRefreshTokenSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	tokenHash := security.HashWithPepper(signature, s.cfg.TokenPepper)

	token, err := s.pgDB.Queries.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fosite.ErrNotFound
		}
		return nil, err
	}

	// Token Reuse Detection
	if token.RevokedAt.Valid || token.ReuseDetectedAt.Valid {
		// Revoke entire family!
		_ = s.pgDB.Queries.RevokeRefreshTokenFamily(ctx, token.TokenFamilyID)
		_ = s.pgDB.Queries.MarkRefreshTokenReuse(ctx, token.ID)

		// Create Audit Log
		_, _ = s.pgDB.Queries.CreateAuditLog(ctx, db.CreateAuditLogParams{
			ID:        uuid.New(),
			EventType: "refresh_token_reuse_detected",
			UserID:    uuid.NullUUID{UUID: token.UserID, Valid: true},
			Result:    "failure",
		})

		return nil, fosite.ErrInactiveToken.WithDescription("Refresh token reuse detected. Revoked all tokens in the family.")
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, fosite.ErrTokenExpired.WithDescription("Refresh token expired")
	}

	client, err := s.pgDB.Queries.GetClientByID(ctx, token.ClientID)
	if err != nil {
		return nil, err
	}

	user, err := s.pgDB.Queries.GetUserByID(ctx, token.UserID)
	if err != nil {
		return nil, err
	}

	fositeClient, err := s.GetClient(ctx, client.ClientID)
	if err != nil {
		return nil, err
	}

	req := fosite.NewRequest()
	req.Client = fositeClient
	req.GrantedScope = fosite.Arguments(strings.Split(token.Scope, " "))
	req.RequestedScope = fosite.Arguments(strings.Split(token.Scope, " "))

	session.SetExpiresAt(fosite.RefreshToken, token.ExpiresAt)

	// Propagate token family ID and current token ID (as parent_id) in extra claims
	if oidcSession, ok := session.(*openid.DefaultSession); ok {
		oidcSession.Subject = user.Sub
		if oidcSession.Claims != nil {
			oidcSession.Claims.Subject = user.Sub
			if oidcSession.Claims.Extra == nil {
				oidcSession.Claims.Extra = make(map[string]interface{})
			}
			oidcSession.Claims.Extra["token_family_id"] = token.TokenFamilyID.String()
			oidcSession.Claims.Extra["parent_id"] = token.ID.String()
		}
	}

	req.Session = session

	return req, nil
}

func (s *fositeStore) DeleteRefreshTokenSession(ctx context.Context, signature string) error {
	tokenHash := security.HashWithPepper(signature, s.cfg.TokenPepper)
	token, err := s.pgDB.Queries.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		return err
	}

	// We mark it as revoked (rotated)
	err = s.pgDB.Queries.RotateRefreshToken(ctx, db.RotateRefreshTokenParams{
		ID:          token.ID,
		RotatedToID: uuid.NullUUID{Valid: false},
	})
	return err
}

func (s *fositeStore) RotateRefreshToken(ctx context.Context, requestID string, refreshTokenSignature string) error {
	tokenHash := security.HashWithPepper(refreshTokenSignature, s.cfg.TokenPepper)
	token, err := s.pgDB.Queries.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		return err
	}

	return s.pgDB.Queries.RotateRefreshToken(ctx, db.RotateRefreshTokenParams{
		ID:          token.ID,
		RotatedToID: uuid.NullUUID{Valid: false},
	})
}

// PKCE Storage methods
func (s *fositeStore) CreatePKCERequestSession(ctx context.Context, signature string, requester fosite.Requester) error {
	return s.CreateAuthorizeCodeSession(ctx, signature, requester)
}

func (s *fositeStore) GetPKCERequestSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	return s.GetAuthorizeCodeSession(ctx, signature, session)
}

func (s *fositeStore) DeletePKCERequestSession(ctx context.Context, signature string) error {
	return s.InvalidateAuthorizeCodeSession(ctx, signature)
}

// OpenIDConnectRequestStorage methods
func (s *fositeStore) CreateOpenIDConnectSession(ctx context.Context, authorizeCode string, requester fosite.Requester) error {
	return s.CreateAuthorizeCodeSession(ctx, authorizeCode, requester)
}

func (s *fositeStore) GetOpenIDConnectSession(ctx context.Context, authorizeCode string, requester fosite.Requester) (fosite.Requester, error) {
	return s.GetAuthorizeCodeSession(ctx, authorizeCode, requester.GetSession())
}

func (s *fositeStore) DeleteOpenIDConnectSession(ctx context.Context, authorizeCode string) error {
	return s.InvalidateAuthorizeCodeSession(ctx, authorizeCode)
}

// AccessTokenStorage methods (no-op for stateless JWT access tokens)
func (s *fositeStore) CreateAccessTokenSession(ctx context.Context, signature string, requester fosite.Requester) error {
	return nil
}

func (s *fositeStore) GetAccessTokenSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	return nil, fosite.ErrNotFound
}

func (s *fositeStore) DeleteAccessTokenSession(ctx context.Context, signature string) error {
	return nil
}

// TokenRevocationStorage methods
func (s *fositeStore) RevokeRefreshToken(ctx context.Context, requestID string) error {
	// No-op for MVP (client application doesn't use RFC 7009 token revocation)
	return nil
}

func (s *fositeStore) RevokeAccessToken(ctx context.Context, requestID string) error {
	// No-op for MVP
	return nil
}
