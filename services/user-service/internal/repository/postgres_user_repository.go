package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/tolik-tachyon/AP2_Final/services/user-service/internal/domain"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	const query = `
		INSERT INTO users (
			id,
			username,
			email,
			password_hash,
			avatar_url,
			bio,
			role,
			is_verified,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.AvatarURL,
		user.Bio,
		user.Role,
		user.IsVerified,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return err
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	const query = `
		SELECT
			id,
			username,
			email,
			password_hash,
			avatar_url,
			bio,
			role,
			is_verified,
			created_at,
			updated_at
		FROM users
		WHERE id = $1
	`

	return r.scanUser(r.db.QueryRowContext(ctx, query, id))
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `
		SELECT
			id,
			username,
			email,
			password_hash,
			avatar_url,
			bio,
			role,
			is_verified,
			created_at,
			updated_at
		FROM users
		WHERE email = $1
	`

	return r.scanUser(r.db.QueryRowContext(ctx, query, email))
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *domain.User) error {
	const query = `
		UPDATE users
		SET
			username = $2,
			email = $3,
			password_hash = $4,
			avatar_url = $5,
			bio = $6,
			role = $7,
			is_verified = $8,
			updated_at = $9
		WHERE id = $1
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.AvatarURL,
		user.Bio,
		user.Role,
		user.IsVerified,
		user.UpdatedAt,
	)
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

func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
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

func (r *PostgresUserRepository) scanUser(row *sql.Row) (*domain.User, error) {
	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.Bio,
		&user.Role,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}
