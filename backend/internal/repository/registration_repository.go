package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ys-1052/yossid/backend/internal/storage/postgres"
	"github.com/ys-1052/yossid/backend/internal/storage/postgres/db"
)

type RegistrationRepository interface {
	CreatePending(ctx context.Context, pending *db.PendingUserRegistration) error
	GetPendingByTokenHash(ctx context.Context, tokenHash string) (*db.PendingUserRegistration, error)
	ConfirmRegistration(ctx context.Context, pendingID uuid.UUID, user *db.User, profile *db.UserProfile, creds *db.UserPasswordCredential) error
}

type registrationRepository struct {
	pgDB *postgres.DB
}

func NewRegistrationRepository(pgDB *postgres.DB) RegistrationRepository {
	return &registrationRepository{pgDB: pgDB}
}

func (r *registrationRepository) CreatePending(ctx context.Context, pending *db.PendingUserRegistration) error {
	_, err := r.pgDB.Queries.CreatePendingRegistration(ctx, db.CreatePendingRegistrationParams{
		ID:                    pending.ID,
		Email:                 pending.Email,
		PasswordHash:          pending.PasswordHash,
		FamilyName:            pending.FamilyName,
		GivenName:             pending.GivenName,
		FamilyNameKana:        pending.FamilyNameKana,
		GivenNameKana:         pending.GivenNameKana,
		Gender:                pending.Gender,
		Birthdate:             pending.Birthdate,
		CountryCode:           pending.CountryCode,
		VerificationTokenHash: pending.VerificationTokenHash,
		ExpiresAt:             pending.ExpiresAt,
	})
	return err
}

func (r *registrationRepository) GetPendingByTokenHash(ctx context.Context, tokenHash string) (*db.PendingUserRegistration, error) {
	pending, err := r.pgDB.Queries.GetPendingRegistrationByToken(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	return &pending, nil
}

func (r *registrationRepository) ConfirmRegistration(ctx context.Context, pendingID uuid.UUID, user *db.User, profile *db.UserProfile, creds *db.UserPasswordCredential) error {
	// Must execute in database transaction!
	return r.pgDB.ExecuteTx(ctx, func(q *db.Queries) error {
		// 1. Check if user already exists (to avoid duplicate emails)
		var existingCount int64
		err := r.pgDB.DB.QueryRowContext(ctx, "SELECT COUNT(1) FROM users WHERE email = $1", user.Email).Scan(&existingCount)
		if err != nil {
			return fmt.Errorf("failed to check existing user: %w", err)
		}
		if existingCount > 0 {
			return fmt.Errorf("email %s is already registered", user.Email)
		}

		// 2. Create user
		_, err = q.CreateUser(ctx, db.CreateUserParams{
			ID:              user.ID,
			Sub:             user.Sub,
			Email:           user.Email,
			EmailVerifiedAt: user.EmailVerifiedAt,
			Status:          user.Status,
		})
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		// 3. Create user profile
		_, err = q.CreateUserProfile(ctx, db.CreateUserProfileParams{
			UserID:         profile.UserID,
			FamilyName:     profile.FamilyName,
			GivenName:      profile.GivenName,
			FamilyNameKana: profile.FamilyNameKana,
			GivenNameKana:  profile.GivenNameKana,
			Gender:         profile.Gender,
			Birthdate:      profile.Birthdate,
			CountryCode:    profile.CountryCode,
		})
		if err != nil {
			return fmt.Errorf("failed to create user profile: %w", err)
		}

		// 4. Create user credentials
		_, err = q.CreatePasswordCredential(ctx, db.CreatePasswordCredentialParams{
			UserID:       creds.UserID,
			PasswordHash: creds.PasswordHash,
		})
		if err != nil {
			return fmt.Errorf("failed to create user password credential: %w", err)
		}

		// 5. Mark pending registration as used
		err = q.MarkPendingRegistrationUsed(ctx, pendingID)
		if err != nil {
			return fmt.Errorf("failed to mark pending registration used: %w", err)
		}

		// 6. Record audit log
		_, err = q.CreateAuditLog(ctx, db.CreateAuditLogParams{
			ID:        uuid.New(),
			EventType: "user_registered",
			UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
			Result:    "success",
		})
		if err != nil {
			return fmt.Errorf("failed to create registration audit log: %w", err)
		}

		// 7. Record audit log for email verification
		_, err = q.CreateAuditLog(ctx, db.CreateAuditLogParams{
			ID:        uuid.New(),
			EventType: "email_verified",
			UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
			Result:    "success",
		})
		if err != nil {
			return fmt.Errorf("failed to create verification audit log: %w", err)
		}

		return nil
	})
}
