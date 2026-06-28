package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ys-1052/yossid/backend/internal/service"
)

type RegisterHandler struct {
	regService service.RegistrationService
}

func NewRegisterHandler(regService service.RegistrationService) *RegisterHandler {
	return &RegisterHandler{regService: regService}
}

type registerRequest struct {
	Email                string `json:"email"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
	FamilyName           string `json:"family_name"`
	GivenName            string `json:"given_name"`
	FamilyNameKana       string `json:"family_name_kana"`
	GivenNameKana        string `json:"given_name_kana"`
	Gender               string `json:"gender"`
	Birthdate            string `json:"birthdate"`
	CountryCode          string `json:"country_code"`
}

func (h *RegisterHandler) PostRegister(c echo.Context) error {
	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	input := service.RegisterInput{
		Email:                req.Email,
		Password:             req.Password,
		PasswordConfirmation: req.PasswordConfirmation,
		FamilyName:           req.FamilyName,
		GivenName:            req.GivenName,
		FamilyNameKana:       req.FamilyNameKana,
		GivenNameKana:        req.GivenNameKana,
		Gender:               req.Gender,
		BirthdateStr:         req.Birthdate,
		CountryCode:          req.CountryCode,
	}

	err := h.regService.RegisterPending(c.Request().Context(), input)
	if err != nil {
		// Silent success if user already exists to prevent email enumeration
		if errors.Is(err, service.ErrUserAlreadyExists) {
			return c.JSON(http.StatusOK, map[string]string{
				"message": "Verification email sent if the address is not already registered.",
			})
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Verification email sent successfully.",
	})
}

func (h *RegisterHandler) VerifyEmail(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		// Redirect to frontend error page
		return c.Redirect(http.StatusFound, "/email/verify?status=error&message=token_required")
	}

	err := h.regService.VerifyEmailToken(c.Request().Context(), token)
	if err != nil {
		var redirectMsg string
		switch {
		case errors.Is(err, service.ErrTokenExpired):
			redirectMsg = "token_expired"
		case errors.Is(err, service.ErrTokenUsed):
			redirectMsg = "token_used"
		default:
			redirectMsg = "invalid_token"
		}
		return c.Redirect(http.StatusFound, "/email/verify?status=error&message="+redirectMsg)
	}

	// Success redirect
	return c.Redirect(http.StatusFound, "/email/verify?status=success")
}
