package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ys-1052/yossid/backend/internal/config"
	"github.com/ys-1052/yossid/backend/internal/mail"
	"github.com/ys-1052/yossid/backend/internal/repository"
	"github.com/ys-1052/yossid/backend/internal/security"
	"github.com/ys-1052/yossid/backend/internal/storage/postgres/db"
)

var (
	ErrInvalidInput      = errors.New("invalid registration input parameters")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	ErrTokenExpired      = errors.New("verification token has expired")
	ErrTokenUsed         = errors.New("verification token has already been used")
	ErrPendingNotFound   = errors.New("pending registration not found")
)

type RegisterInput struct {
	Email                string
	Password             string
	PasswordConfirmation string
	FamilyName           string
	GivenName            string
	FamilyNameKana       string
	GivenNameKana        string
	Gender               string
	BirthdateStr         string // YYYY-MM-DD
	CountryCode          string // 2-letter uppercase
}

type RegistrationService interface {
	RegisterPending(ctx context.Context, input RegisterInput) error
	VerifyEmailToken(ctx context.Context, token string) error
}

type registrationService struct {
	cfg          *config.Config
	userRepo     repository.UserRepository
	registerRepo repository.RegistrationRepository
	mailer       mail.Mailer
}

func NewRegistrationService(cfg *config.Config, userRepo repository.UserRepository, registerRepo repository.RegistrationRepository, mailer mail.Mailer) RegistrationService {
	return &registrationService{
		cfg:          cfg,
		userRepo:     userRepo,
		registerRepo: registerRepo,
		mailer:       mailer,
	}
}

func (s *registrationService) RegisterPending(ctx context.Context, input RegisterInput) error {
	// 1. Validation checks
	if !security.ValidateEmail(input.Email) {
		return fmt.Errorf("%w: invalid email address", ErrInvalidInput)
	}
	if !security.ValidatePassword(input.Password) {
		return fmt.Errorf("%w: password must be at least 8 chars and contain uppercase, lowercase, and a digit", ErrInvalidInput)
	}
	if input.Password != input.PasswordConfirmation {
		return fmt.Errorf("%w: passwords do not match", ErrInvalidInput)
	}
	if input.FamilyName == "" || input.GivenName == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	if !security.ValidateKatakana(input.FamilyNameKana) || !security.ValidateKatakana(input.GivenNameKana) {
		return fmt.Errorf("%w: name kana must contain only full-width katakana", ErrInvalidInput)
	}
	if !security.ValidateGender(input.Gender) {
		return fmt.Errorf("%w: invalid gender value", ErrInvalidInput)
	}
	birthdate, ok := security.ValidateBirthdate(input.BirthdateStr)
	if !ok {
		return fmt.Errorf("%w: invalid birthdate format or value (use YYYY-MM-DD)", ErrInvalidInput)
	}
	if !security.ValidateCountryCode(input.CountryCode) {
		return fmt.Errorf("%w: invalid country code (use 2-letter ISO uppercase)", ErrInvalidInput)
	}

	// 2. Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err == nil && existingUser != nil {
		// User already exists. To avoid email enumeration attacks, we do NOT return an error to the frontend,
		// but we do return an error internally.
		return ErrUserAlreadyExists
	}

	// 3. Create password hash
	passwordHash, err := security.HashPassword(input.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 4. Generate verification token
	token, err := security.GenerateRandomToken()
	if err != nil {
		return fmt.Errorf("failed to generate verification token: %w", err)
	}
	tokenHash := security.HashWithPepper(token, s.cfg.TokenPepper)

	// 5. Store pending registration
	pendingID := uuid.New()
	pending := &db.PendingUserRegistration{
		ID:                    pendingID,
		Email:                 input.Email,
		PasswordHash:          passwordHash,
		FamilyName:            input.FamilyName,
		GivenName:             input.GivenName,
		FamilyNameKana:        input.FamilyNameKana,
		GivenNameKana:         input.GivenNameKana,
		Gender:                input.Gender,
		Birthdate:             birthdate,
		CountryCode:           input.CountryCode,
		VerificationTokenHash: tokenHash,
		ExpiresAt:             time.Now().Add(30 * time.Minute), // Expires in 30 minutes
	}

	err = s.registerRepo.CreatePending(ctx, pending)
	if err != nil {
		return fmt.Errorf("failed to save pending registration: %w", err)
	}

	// 6. Send verification email via SES
	err = s.mailer.SendVerificationEmail(ctx, input.Email, token)
	if err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

func (s *registrationService) VerifyEmailToken(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("%w: empty token", ErrInvalidInput)
	}

	// 1. Hash incoming token with pepper
	tokenHash := security.HashWithPepper(token, s.cfg.TokenPepper)

	// 2. Lookup pending registration
	pending, err := s.registerRepo.GetPendingByTokenHash(ctx, tokenHash)
	if err != nil {
		// DB error or not found
		return ErrPendingNotFound
	}

	// 3. Verify status
	if pending.UsedAt.Valid {
		return ErrTokenUsed
	}

	if time.Now().After(pending.ExpiresAt) {
		return ErrTokenExpired
	}

	// 4. Prepare user, profile, and credentials
	userID := uuid.New()
	sub := fmt.Sprintf("usr_%s", uuid.New().String()) // generate OIDC subject identifier

	user := &db.User{
		ID:              userID,
		Sub:             sub,
		Email:           pending.Email,
		EmailVerifiedAt: time.Now(),
		Status:          "active",
	}

	profile := &db.UserProfile{
		UserID:         userID,
		FamilyName:     sql.NullString{String: pending.FamilyName, Valid: pending.FamilyName != ""},
		GivenName:      sql.NullString{String: pending.GivenName, Valid: pending.GivenName != ""},
		FamilyNameKana: sql.NullString{String: pending.FamilyNameKana, Valid: pending.FamilyNameKana != ""},
		GivenNameKana:  sql.NullString{String: pending.GivenNameKana, Valid: pending.GivenNameKana != ""},
		Gender:         sql.NullString{String: pending.Gender, Valid: pending.Gender != ""},
		Birthdate:      sql.NullTime{Time: pending.Birthdate, Valid: true},
		CountryCode:    sql.NullString{String: pending.CountryCode, Valid: pending.CountryCode != ""},
	}

	creds := &db.UserPasswordCredential{
		UserID:       userID,
		PasswordHash: pending.PasswordHash,
	}

	// 5. Commit registration in database transaction
	err = s.registerRepo.ConfirmRegistration(ctx, pending.ID, user, profile, creds)
	if err != nil {
		return fmt.Errorf("failed to confirm user registration: %w", err)
	}

	return nil
}
