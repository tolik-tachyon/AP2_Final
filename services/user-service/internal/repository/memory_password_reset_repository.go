package repository

import (
	"context"
	"sync"
	"time"

	"github.com/tolik-tachyon/AP2_Final/services/user-service/internal/domain"
)

type MemoryPasswordResetRepository struct {
	mu     sync.RWMutex
	byHash map[string]*domain.PasswordResetToken
}

func NewMemoryPasswordResetRepository() *MemoryPasswordResetRepository {
	return &MemoryPasswordResetRepository{byHash: make(map[string]*domain.PasswordResetToken)}
}

func (r *MemoryPasswordResetRepository) Create(_ context.Context, token *domain.PasswordResetToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copy := *token
	r.byHash[token.TokenHash] = &copy
	return nil
}

func (r *MemoryPasswordResetRepository) GetByTokenHash(_ context.Context, hash string) (*domain.PasswordResetToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	token, ok := r.byHash[hash]
	if !ok {
		return nil, ErrNotFound
	}
	copy := *token
	return &copy, nil
}

func (r *MemoryPasswordResetRepository) MarkUsed(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	for _, token := range r.byHash {
		if token.ID == id {
			token.UsedAt = &now
			return nil
		}
	}
	return ErrNotFound
}
