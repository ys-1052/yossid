package security

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// SetSecureCookie sets a cookie with secure flags: HttpOnly, Secure, SameSite=Lax, Path=/
func SetSecureCookie(c echo.Context, name, value string, expires time.Time) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  expires,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Always set Secure=true since CloudFront distribution or localhost TLS is used
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)
}

// ClearCookie deletes a cookie by setting MaxAge to -1.
func ClearCookie(c echo.Context, name string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Expires:  time.Unix(0, 0),
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)
}
