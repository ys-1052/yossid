package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/token/jwt"
)

func (h *OIDCHandler) PostToken(c echo.Context) error {
	ctx := c.Request().Context()

	// 1. Create empty OIDC session
	session := &openid.DefaultSession{
		Claims:  &jwt.IDTokenClaims{},
		Headers: &jwt.Headers{},
	}

	// 2. Parse token exchange request
	requester, err := h.oauth2.NewAccessRequest(ctx, c.Request(), session)
	if err != nil {
		h.oauth2.WriteAccessError(ctx, c.Response(), requester, err)
		return nil
	}

	// 3. Generate token response containing Access Token, ID Token and Refresh Token
	response, err := h.oauth2.NewAccessResponse(ctx, requester)
	if err != nil {
		h.oauth2.WriteAccessError(ctx, c.Response(), requester, err)
		return nil
	}

	// 4. Return tokens in standard JSON format
	h.oauth2.WriteAccessResponse(ctx, c.Response(), requester, response)
	return nil
}
