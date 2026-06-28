package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *OIDCHandler) GetDiscovery(c echo.Context) error {
	issuer := h.cfg.Issuer

	metadata := map[string]interface{}{
		"issuer":                                issuer,
		"authorization_endpoint":                issuer + "/authorize",
		"token_endpoint":                        issuer + "/token",
		"userinfo_endpoint":                     issuer + "/userinfo",
		"jwks_uri":                              issuer + "/jwks.json",
		"scopes_supported":                      []string{"openid", "profile", "email"},
		"response_types_supported":              []string{"code", "token", "id_token"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic"},
		"claims_supported":                      []string{"iss", "sub", "aud", "exp", "iat", "auth_time", "nonce", "email", "email_verified", "family_name", "given_name", "family_name#ja-Kana-JP", "given_name#ja-Kana-JP", "gender", "birthdate"},
	}

	return c.JSON(http.StatusOK, metadata)
}
