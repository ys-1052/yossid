package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ys-1052/yossid/backend/internal/storage/postgres"
	"github.com/ys-1052/yossid/backend/internal/storage/postgres/db"
)

type OTPRepository interface {
	CreateChallenge(ctx context.Context, challenge *db.EmailOtpChallenge) error
	GetChallengeByHash(ctx context.Context, hash string) (*db.EmailOtpChallenge, error)
	IncrementAttempts(ctx context.Context, id uuid.UUID) (*db.EmailOtpChallenge, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
}

type otpRepository struct {
	pgDB *postgres.DB
}

func NewOTPRepository(pgDB *postgres.DB) OTPRepository {
	return &otpRepository{pgDB: pgDB}
}

func (o *otpRepository) CreateChallenge(ctx context.Context, challenge *db.EmailOtpChallenge) error {
	_, err := o.pgDB.Queries.CreateEmailOTPChallenge(ctx, db.CreateEmailOTPChallengeParams{
		ID:              challenge.ID,
		ChallengeIDHash: challenge.ChallengeIDHash,
		UserID:          challenge.UserID,
		OtpHash:         challenge.OtpHash,
		ExpiresAt:       challenge.ExpiresAt,
	})
	return err
}

func (o *otpRepository) GetChallengeByHash(ctx context.Context, hash string) (*db.EmailOtpChallenge, error) {
	challenge, err := o.pgDB.Queries.GetEmailOTPChallenge(ctx, hash)
	if err != nil {
		return nil, err
	}
	return &challenge, nil
}

func (o *otpRepository) IncrementAttempts(ctx context.Context, id uuid.UUID) (*db.EmailOtpChallenge, error) {
	challenge, err := o.pgDB.Queries.IncrementOTPAttempts(ctx, id)
	if err != nil {
		return nil, err
	}
	return &challenge, nil
}

func (o *otpRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	return o.pgDB.Queries.MarkOTPUsed(ctx, id)
}
