package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/ys-1052/yossid/backend/internal/security"
	"github.com/ys-1052/yossid/backend/internal/service"
)

type LoginHandler struct {
	loginService service.LoginService
}

func NewLoginHandler(loginService service.LoginService) *LoginHandler {
	return &LoginHandler{loginService: loginService}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *LoginHandler) PostLogin(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	input := service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}

	res, err := h.loginService.Login(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid email or password")
		}
		if errors.Is(err, service.ErrAccountInactive) {
			return echo.NewHTTPError(http.StatusForbidden, "Account is disabled or inactive")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"challenge_id": res.ChallengeID,
		"message":      "Verification code sent to your email.",
	})
}

type verifyMFARequest struct {
	ChallengeID string `json:"challenge_id"`
	OTP         string `json:"otp"`
}

func (h *LoginHandler) PostVerifyMFA(c echo.Context) error {
	var req verifyMFARequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	input := service.VerifyMFAInput{
		ChallengeID: req.ChallengeID,
		OTP:         req.OTP,
		IpAddress:   c.RealIP(),
		UserAgent:   c.Request().UserAgent(),
	}

	res, err := h.loginService.VerifyMFA(c.Request().Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrOTPChallengeNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "Challenge not found")
		case errors.Is(err, service.ErrOTPExpired):
			return echo.NewHTTPError(http.StatusGone, "Verification code expired")
		case errors.Is(err, service.ErrOTPMaxAttempts):
			return echo.NewHTTPError(http.StatusForbidden, "Maximum verification attempts exceeded")
		case errors.Is(err, service.ErrOTPInvalid):
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid verification code")
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	// Session is valid for 12 hours
	expires := time.Now().Add(12 * time.Hour)
	security.SetSecureCookie(c, "op_session", res.SessionID, expires)

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Authenticated successfully.",
	})
}

func (h *LoginHandler) PostLogout(c echo.Context) error {
	cookie, err := c.Cookie("op_session")
	if err != nil || cookie.Value == "" {
		return c.JSON(http.StatusOK, map[string]string{"message": "Already logged out."})
	}

	// Revoke session in database
	err = h.loginService.RevokeSession(c.Request().Context(), cookie.Value)
	if err != nil {
		// Log and continue to delete cookie anyway
		c.Logger().Warnf("Failed to revoke session in DB: %v", err)
	}

	// Clear the cookie
	security.ClearCookie(c, "op_session")

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Logged out successfully.",
	})
}
