package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
	"github.com/ys-1052/yossid/backend/internal/config"
	"github.com/ys-1052/yossid/backend/internal/mail"
	"github.com/ys-1052/yossid/backend/internal/repository"
	"github.com/ys-1052/yossid/backend/internal/security"
	"github.com/ys-1052/yossid/backend/internal/storage/postgres/db"
)

var (
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrAccountInactive      = errors.New("account is inactive")
	ErrOTPExpired           = errors.New("verification code has expired")
	ErrOTPMaxAttempts       = errors.New("maximum verification attempts exceeded")
	ErrOTPInvalid           = errors.New("invalid verification code")
	ErrOTPChallengeNotFound = errors.New("verification challenge not found")
)

type LoginInput struct {
	Email    string
	Password string
}

type LoginResult struct {
	ChallengeID string
}

type VerifyMFAInput struct {
	ChallengeID string
	OTP         string
	IpAddress   string
	UserAgent   string
}

type VerifyMFAResult struct {
	SessionID string
	UserID    string
}

type LoginService interface {
	Login(ctx context.Context, input LoginInput) (*LoginResult, error)
	VerifyMFA(ctx context.Context, input VerifyMFAInput) (*VerifyMFAResult, error)
	GetSession(ctx context.Context, sessionID string) (*db.LoginSession, error)
	RevokeSession(ctx context.Context, sessionID string) error
}

type loginService struct {
	cfg         *config.Config
	userRepo    repository.UserRepository
	otpRepo     repository.OTPRepository
	sessionRepo repository.SessionRepository
	auditRepo   repository.AuditRepository
	mailer      mail.Mailer
}

func NewLoginService(cfg *config.Config, userRepo repository.UserRepository, otpRepo repository.OTPRepository, sessionRepo repository.SessionRepository, auditRepo repository.AuditRepository, mailer mail.Mailer) LoginService {
	return &loginService{
		cfg:         cfg,
		userRepo:    userRepo,
		otpRepo:     otpRepo,
		sessionRepo: sessionRepo,
		auditRepo:   auditRepo,
		mailer:      mailer,
	}
}

func (s *loginService) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	// 1. Get user by email
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		s.logAudit(ctx, "login_password_failure", uuid.NullUUID{}, "user_not_found", "", "", nil)
		return nil, ErrInvalidCredentials
	}

	// 2. Check if user is active
	if user.Status != "active" {
		s.logAudit(ctx, "login_password_failure", uuid.NullUUID{UUID: user.ID, Valid: true}, "user_inactive", "", "", nil)
		return nil, ErrAccountInactive
	}

	// 3. Get password credential
	cred, err := s.userRepo.GetCredential(ctx, user.ID)
	if err != nil {
		s.logAudit(ctx, "login_password_failure", uuid.NullUUID{UUID: user.ID, Valid: true}, "credentials_missing", "", "", nil)
		return nil, ErrInvalidCredentials
	}

	// 4. Verify password
	ok, err := security.VerifyPassword(input.Password, cred.PasswordHash)
	if err != nil || !ok {
		s.logAudit(ctx, "login_password_failure", uuid.NullUUID{UUID: user.ID, Valid: true}, "password_incorrect", "", "", nil)
		return nil, ErrInvalidCredentials
	}

	// 5. Generate OTP Challenge
	challengeID, err := security.GenerateRandomToken()
	if err != nil {
		return nil, err
	}
	challengeIDHash := security.HashWithPepper(challengeID, s.cfg.TokenPepper)

	otp, err := security.GenerateRandomOTP()
	if err != nil {
		return nil, err
	}
	otpHash := security.HashWithPepper(otp, s.cfg.OTPPepper)

	challenge := &db.EmailOtpChallenge{
		ID:              uuid.New(),
		ChallengeIDHash: challengeIDHash,
		UserID:          user.ID,
		OtpHash:         otpHash,
		ExpiresAt:       time.Now().Add(5 * time.Minute), // 5 minutes validity
	}

	err = s.otpRepo.CreateChallenge(ctx, challenge)
	if err != nil {
		return nil, fmt.Errorf("failed to save OTP challenge: %w", err)
	}

	// 6. Send OTP email
	err = s.mailer.SendOTPEmail(ctx, user.Email, otp)
	if err != nil {
		return nil, fmt.Errorf("failed to send OTP email: %w", err)
	}

	// Log success
	s.logAudit(ctx, "login_password_success", uuid.NullUUID{UUID: user.ID, Valid: true}, "success", "", "", nil)
	s.logAudit(ctx, "email_otp_sent", uuid.NullUUID{UUID: user.ID, Valid: true}, "success", "", "", nil)

	return &LoginResult{ChallengeID: challengeID}, nil
}

func (s *loginService) VerifyMFA(ctx context.Context, input VerifyMFAInput) (*VerifyMFAResult, error) {
	// 1. Hash challenge ID and fetch challenge
	challengeHash := security.HashWithPepper(input.ChallengeID, s.cfg.TokenPepper)
	challenge, err := s.otpRepo.GetChallengeByHash(ctx, challengeHash)
	if err != nil {
		return nil, ErrOTPChallengeNotFound
	}

	userNullUUID := uuid.NullUUID{UUID: challenge.UserID, Valid: true}

	// 2. Verify challenge status
	if challenge.UsedAt.Valid {
		s.logAudit(ctx, "email_otp_failure", userNullUUID, "otp_already_used", input.IpAddress, input.UserAgent, nil)
		return nil, ErrOTPExpired
	}

	if time.Now().After(challenge.ExpiresAt) {
		s.logAudit(ctx, "email_otp_failure", userNullUUID, "otp_expired", input.IpAddress, input.UserAgent, nil)
		return nil, ErrOTPExpired
	}

	if challenge.Attempts >= challenge.MaxAttempts {
		s.logAudit(ctx, "email_otp_failure", userNullUUID, "max_attempts_exceeded", input.IpAddress, input.UserAgent, nil)
		return nil, ErrOTPMaxAttempts
	}

	// 3. Increment attempts in DB
	challenge, err = s.otpRepo.IncrementAttempts(ctx, challenge.ID)
	if err != nil {
		return nil, err
	}

	// 4. Verify OTP
	inputOTPHash := security.HashWithPepper(input.OTP, s.cfg.OTPPepper)
	if inputOTPHash != challenge.OtpHash {
		s.logAudit(ctx, "email_otp_failure", userNullUUID, "otp_incorrect", input.IpAddress, input.UserAgent, nil)
		return nil, ErrOTPInvalid
	}

	// 5. Mark OTP challenge as used
	err = s.otpRepo.MarkUsed(ctx, challenge.ID)
	if err != nil {
		return nil, err
	}

	// 6. Create login session
	sessionID, err := security.GenerateRandomToken()
	if err != nil {
		return nil, err
	}
	sessionHash := security.HashWithPepper(sessionID, s.cfg.TokenPepper)

	session := &db.LoginSession{
		ID:          uuid.New(),
		SessionHash: sessionHash,
		UserID:      challenge.UserID,
		AuthTime:    time.Now(),
		Amr:         "pwd email",
		IpAddress:   sql.NullString{String: input.IpAddress, Valid: input.IpAddress != ""},
		UserAgent:   sql.NullString{String: input.UserAgent, Valid: input.UserAgent != ""},
		ExpiresAt:   time.Now().Add(12 * time.Hour),
	}

	err = s.sessionRepo.CreateSession(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to save login session: %w", err)
	}

	// Log success
	s.logAudit(ctx, "email_otp_success", userNullUUID, "success", input.IpAddress, input.UserAgent, nil)
	s.logAudit(ctx, "op_session_created", userNullUUID, "success", input.IpAddress, input.UserAgent, nil)

	return &VerifyMFAResult{
		SessionID: sessionID,
		UserID:    challenge.UserID.String(),
	}, nil
}

func (s *loginService) GetSession(ctx context.Context, sessionID string) (*db.LoginSession, error) {
	sessionHash := security.HashWithPepper(sessionID, s.cfg.TokenPepper)
	session, err := s.sessionRepo.GetSessionByHash(ctx, sessionHash)
	if err != nil {
		return nil, err
	}

	if session.RevokedAt.Valid {
		return nil, errors.New("session is revoked")
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, errors.New("session has expired")
	}

	return session, nil
}

func (s *loginService) RevokeSession(ctx context.Context, sessionID string) error {
	sessionHash := security.HashWithPepper(sessionID, s.cfg.TokenPepper)
	session, err := s.sessionRepo.GetSessionByHash(ctx, sessionHash)
	if err != nil {
		return err
	}

	err = s.sessionRepo.RevokeSession(ctx, session.ID)
	if err != nil {
		return err
	}

	s.logAudit(ctx, "logout", uuid.NullUUID{UUID: session.UserID, Valid: true}, "success", "", "", nil)
	return nil
}

// Helper to log audit events
func (s *loginService) logAudit(ctx context.Context, eventType string, userID uuid.NullUUID, result string, ip, ua string, metadata []byte) {
	var pqMeta pqtype.NullRawMessage
	if len(metadata) > 0 {
		pqMeta = pqtype.NullRawMessage{
			RawMessage: metadata,
			Valid:      true,
		}
	}

	err := s.auditRepo.Create(ctx, db.CreateAuditLogParams{
		ID:        uuid.New(),
		EventType: eventType,
		UserID:    userID,
		Result:    result,
		IpAddress: sql.NullString{String: ip, Valid: ip != ""},
		UserAgent: sql.NullString{String: ua, Valid: ua != ""},
		Metadata:  pqMeta,
	})
	if err != nil {
		log.Printf("ERROR: Failed to save audit log: %v", err)
	}
}
