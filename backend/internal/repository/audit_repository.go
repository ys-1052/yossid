package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
	"github.com/ys-1052/yossid/backend/internal/storage/postgres"
	"github.com/ys-1052/yossid/backend/internal/storage/postgres/db"
)

type AuditRepository interface {
	LogEvent(ctx context.Context, eventType string, userID *uuid.UUID, clientID *uuid.UUID, result string, ipAddress, userAgent string, metadata map[string]interface{}) error
	Create(ctx context.Context, logParams db.CreateAuditLogParams) error
}

type auditRepository struct {
	pgDB *postgres.DB
}

func NewAuditRepository(pgDB *postgres.DB) AuditRepository {
	return &auditRepository{pgDB: pgDB}
}

func (a *auditRepository) LogEvent(ctx context.Context, eventType string, userID *uuid.UUID, clientID *uuid.UUID, result string, ipAddress, userAgent string, metadata map[string]interface{}) error {
	var userNullUUID uuid.NullUUID
	if userID != nil {
		userNullUUID = uuid.NullUUID{UUID: *userID, Valid: true}
	}

	var clientNullUUID uuid.NullUUID
	if clientID != nil {
		clientNullUUID = uuid.NullUUID{UUID: *clientID, Valid: true}
	}

	var metaBytes []byte
	if metadata != nil {
		var err error
		metaBytes, err = json.Marshal(metadata)
		if err != nil {
			return err
		}
	}

	var pqMeta pqtype.NullRawMessage
	if len(metaBytes) > 0 {
		pqMeta = pqtype.NullRawMessage{
			RawMessage: metaBytes,
			Valid:      true,
		}
	}

	params := db.CreateAuditLogParams{
		ID:        uuid.New(),
		EventType: eventType,
		UserID:    userNullUUID,
		ClientID:  clientNullUUID,
		Result:    result,
		IpAddress: sql.NullString{String: ipAddress, Valid: ipAddress != ""},
		UserAgent: sql.NullString{String: userAgent, Valid: userAgent != ""},
		Metadata:  pqMeta,
	}

	return a.Create(ctx, params)
}

func (a *auditRepository) Create(ctx context.Context, logParams db.CreateAuditLogParams) error {
	_, err := a.pgDB.Queries.CreateAuditLog(ctx, logParams)
	return err
}
