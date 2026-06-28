package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ys-1052/yossid/backend/internal/storage/postgres"
	"github.com/ys-1052/yossid/backend/internal/storage/postgres/db"
)

type SessionRepository interface {
	CreateSession(ctx context.Context, session *db.LoginSession) error
	GetSessionByHash(ctx context.Context, hash string) (*db.LoginSession, error)
	RevokeSession(ctx context.Context, id uuid.UUID) error
}

type sessionRepository struct {
	pgDB *postgres.DB
}

func NewSessionRepository(pgDB *postgres.DB) SessionRepository {
	return &sessionRepository{pgDB: pgDB}
}

func (s *sessionRepository) CreateSession(ctx context.Context, session *db.LoginSession) error {
	_, err := s.pgDB.Queries.CreateLoginSession(ctx, db.CreateLoginSessionParams{
		ID:          session.ID,
		SessionHash: session.SessionHash,
		UserID:      session.UserID,
		AuthTime:    session.AuthTime,
		Amr:         session.Amr,
		IpAddress:   session.IpAddress,
		UserAgent:   session.UserAgent,
		ExpiresAt:   session.ExpiresAt,
	})
	return err
}

func (s *sessionRepository) GetSessionByHash(ctx context.Context, hash string) (*db.LoginSession, error) {
	session, err := s.pgDB.Queries.GetLoginSession(ctx, hash)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *sessionRepository) RevokeSession(ctx context.Context, id uuid.UUID) error {
	return s.pgDB.Queries.RevokeLoginSession(ctx, id)
}
