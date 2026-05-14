package repository

import (
	"context"

	"github.com/tolik-tachyon/AP2_Final/services/user-service/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *domain.LoginSession) error
	GetByRefreshTokenHash(ctx context.Context, hash string) (*domain.LoginSession, error)
	DeleteByRefreshTokenHash(ctx context.Context, hash string) error
}

type PasswordResetRepository interface {
	Create(ctx context.Context, token *domain.PasswordResetToken) error
	GetByTokenHash(ctx context.Context, hash string) (*domain.PasswordResetToken, error)
	MarkUsed(ctx context.Context, id string) error
}
