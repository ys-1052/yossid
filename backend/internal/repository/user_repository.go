package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/ys-1052/yossid/backend/internal/storage/postgres"
	"github.com/ys-1052/yossid/backend/internal/storage/postgres/db"
)

type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*db.User, error)
	GetBySub(ctx context.Context, sub string) (*db.User, error)
	GetByEmail(ctx context.Context, email string) (*db.User, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*db.UserProfile, error)
	GetCredential(ctx context.Context, userID uuid.UUID) (*db.UserPasswordCredential, error)
	UpdateProfile(ctx context.Context, profile *db.UserProfile) error
	Withdraw(ctx context.Context, userID uuid.UUID, reason string, ipAddress, userAgent string) error
}

type userRepository struct {
	pgDB *postgres.DB
}

func NewUserRepository(pgDB *postgres.DB) UserRepository {
	return &userRepository{pgDB: pgDB}
}

func (u *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*db.User, error) {
	user, err := u.pgDB.Queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *userRepository) GetBySub(ctx context.Context, sub string) (*db.User, error) {
	user, err := u.pgDB.Queries.GetUserBySub(ctx, sub)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *userRepository) GetByEmail(ctx context.Context, email string) (*db.User, error) {
	user, err := u.pgDB.Queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *userRepository) GetProfile(ctx context.Context, userID uuid.UUID) (*db.UserProfile, error) {
	profile, err := u.pgDB.Queries.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (u *userRepository) GetCredential(ctx context.Context, userID uuid.UUID) (*db.UserPasswordCredential, error) {
	cred, err := u.pgDB.Queries.GetPasswordCredential(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

func (u *userRepository) UpdateProfile(ctx context.Context, profile *db.UserProfile) error {
	_, err := u.pgDB.Queries.UpdateUserProfile(ctx, db.UpdateUserProfileParams{
		UserID:         profile.UserID,
		FamilyName:     profile.FamilyName,
		GivenName:      profile.GivenName,
		FamilyNameKana: profile.FamilyNameKana,
		GivenNameKana:  profile.GivenNameKana,
		Gender:         profile.Gender,
		Birthdate:      profile.Birthdate,
		CountryCode:    profile.CountryCode,
	})
	return err
}

func (u *userRepository) Withdraw(ctx context.Context, userID uuid.UUID, reason string, ipAddress, userAgent string) error {
	// Execute in transaction
	return u.pgDB.ExecuteTx(ctx, func(q *db.Queries) error {
		// 1. Update user status to withdrawn
		err := q.WithdrawUser(ctx, userID)
		if err != nil {
			return err
		}

		// 2. Create user withdrawal record
		_, err = q.CreateUserWithdrawal(ctx, db.CreateUserWithdrawalParams{
			ID:        uuid.New(),
			UserID:    userID,
			Reason:    sql.NullString{String: reason, Valid: reason != ""},
			IpAddress: sql.NullString{String: ipAddress, Valid: ipAddress != ""},
			UserAgent: sql.NullString{String: userAgent, Valid: userAgent != ""},
		})
		if err != nil {
			return err
		}

		// 3. Revoke all login sessions for this user
		_, err = u.pgDB.DB.ExecContext(ctx, "UPDATE login_sessions SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL", userID)
		if err != nil {
			return err
		}

		// 4. Revoke all refresh tokens for this user
		_, err = u.pgDB.DB.ExecContext(ctx, "UPDATE refresh_tokens SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL", userID)
		if err != nil {
			return err
		}

		// 5. Create audit log
		_, err = q.CreateAuditLog(ctx, db.CreateAuditLogParams{
			ID:        uuid.New(),
			EventType: "user_withdrawn",
			UserID:    uuid.NullUUID{UUID: userID, Valid: true},
			Result:    "success",
			IpAddress: sql.NullString{String: ipAddress, Valid: ipAddress != ""},
			UserAgent: sql.NullString{String: userAgent, Valid: userAgent != ""},
		})
		if err != nil {
			return err
		}

		return nil
	})
}
