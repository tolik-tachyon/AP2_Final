package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/tolik-tachyon/AP2_Final/services/user-service/internal/domain"
)

type PostgresSessionRepository struct {
	db *sql.DB
}

func NewPostgresSessionRepository(db *sql.DB) *PostgresSessionRepository {
	return &PostgresSessionRepository{db: db}
}

func (r *PostgresSessionRepository) Create(ctx context.Context, session *domain.LoginSession) error {
	const query = `
		INSERT INTO login_sessions (
			id,
			user_id,
			refresh_token_hash,
			user_agent,
			ip_address,
			expires_at,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		session.ID,
		session.UserID,
		session.RefreshTokenHash,
		session.UserAgent,
		session.IPAddress,
		session.ExpiresAt,
		session.CreatedAt,
	)
	return err
}

func (r *PostgresSessionRepository) GetByRefreshTokenHash(ctx context.Context, hash string) (*domain.LoginSession, error) {
	const query = `
		SELECT
			id,
			user_id,
			refresh_token_hash,
			user_agent,
			ip_address,
			expires_at,
			created_at
		FROM login_sessions
		WHERE refresh_token_hash = $1
	`

	var session domain.LoginSession
	err := r.db.QueryRowContext(ctx, query, hash).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&session.UserAgent,
		&session.IPAddress,
		&session.ExpiresAt,
		&session.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *PostgresSessionRepository) DeleteByRefreshTokenHash(ctx context.Context, hash string) error {
	const query = `DELETE FROM login_sessions WHERE refresh_token_hash = $1`

	result, err := r.db.ExecContext(ctx, query, hash)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
