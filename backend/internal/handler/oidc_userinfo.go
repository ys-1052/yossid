package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
)

func (h *OIDCHandler) GetUserInfo(c echo.Context) error {
	ctx := c.Request().Context()

	// 1. Extract access token from Authorization header (Bearer token)
	authHeader := c.Request().Header.Get("Authorization")
	tokenStr := ""
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		tokenStr = authHeader[7:]
	} else {
		// Try query param
		tokenStr = c.QueryParam("access_token")
	}

	if tokenStr == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "Bearer token required")
	}

	// 2. Introspect and validate access token
	session := &openid.DefaultSession{}
	_, accessRequester, err := h.oauth2.IntrospectToken(ctx, tokenStr, fosite.AccessToken, session, "")
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired access token: "+err.Error())
	}

	sub := accessRequester.GetSession().GetSubject()
	if sub == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "Subject claim missing in token session")
	}

	// 3. Retrieve user profile
	user, err := h.pgDB.Queries.GetUserBySub(ctx, sub)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "User not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	profile, err := h.pgDB.Queries.GetUserProfile(ctx, user.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// 4. Map user info claims based on granted scopes
	claims := map[string]interface{}{
		"sub": sub,
	}

	grantedScopes := accessRequester.GetGrantedScopes()
	hasScope := func(name string) bool {
		for _, s := range grantedScopes {
			if s == name {
				return true
			}
		}
		return false
	}

	if hasScope("email") {
		claims["email"] = user.Email
		claims["email_verified"] = !user.EmailVerifiedAt.IsZero()
	}

	if hasScope("profile") && err == nil {
		if profile.FamilyName.Valid {
			claims["family_name"] = profile.FamilyName.String
		}
		if profile.GivenName.Valid {
			claims["given_name"] = profile.GivenName.String
		}
		if profile.FamilyNameKana.Valid {
			claims["family_name#ja-Kana-JP"] = profile.FamilyNameKana.String
		}
		if profile.GivenNameKana.Valid {
			claims["given_name#ja-Kana-JP"] = profile.GivenNameKana.String
		}
		if profile.Gender.Valid {
			claims["gender"] = profile.Gender.String
		}
		if profile.Birthdate.Valid {
			claims["birthdate"] = profile.Birthdate.Time.Format("2006-01-02")
		}
	}

	return c.JSON(http.StatusOK, claims)
}

func (h *OIDCHandler) PostUserInfo(c echo.Context) error {
	return h.GetUserInfo(c)
}
