package handler

import (
	"net/http"

	"github.com/go-jose/go-jose/v3"
	"github.com/labstack/echo/v4"
)

func (h *OIDCHandler) GetJWKS(c echo.Context) error {
	if h.cfg.JWTPrivateKey == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "JWK Private Key not configured")
	}

	// Active key in JWK format
	jwk := jose.JSONWebKey{
		Key:       &h.cfg.JWTPrivateKey.PublicKey,
		KeyID:     "1", // Standard fallback key ID
		Algorithm: "RS256",
		Use:       "sig",
	}

	jwks := jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{jwk},
	}

	return c.JSON(http.StatusOK, jwks)
}
