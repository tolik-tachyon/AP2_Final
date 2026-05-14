package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/tolik-tachyon/AP2_Final/services/user-service/internal/domain"
)

type PostgresPasswordResetRepository struct {
	db *sql.DB
}

func NewPostgresPasswordResetRepository(db *sql.DB) *PostgresPasswordResetRepository {
	return &PostgresPasswordResetRepository{db: db}
}

func (r *PostgresPasswordResetRepository) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	const query = `
		INSERT INTO password_reset_tokens (
			id,
			user_id,
			token_hash,
			expires_at,
			used_at,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.UsedAt,
		token.CreatedAt,
	)
	return err
}

func (r *PostgresPasswordResetRepository) GetByTokenHash(ctx context.Context, hash string) (*domain.PasswordResetToken, error) {
	const query = `
		SELECT
			id,
			user_id,
			token_hash,
			expires_at,
			used_at,
			created_at
		FROM password_reset_tokens
		WHERE token_hash = $1
	`

	var token domain.PasswordResetToken
	err := r.db.QueryRowContext(ctx, query, hash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.UsedAt,
		&token.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *PostgresPasswordResetRepository) MarkUsed(ctx context.Context, id string) error {
	const query = `UPDATE password_reset_tokens SET used_at = $2 WHERE id = $1 AND used_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id, time.Now().UTC())
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
