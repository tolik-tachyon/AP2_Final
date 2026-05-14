package repository

import (
	"context"
	"sync"

	"github.com/tolik-tachyon/AP2_Final/services/user-service/internal/domain"
)

type MemorySessionRepository struct {
	mu     sync.RWMutex
	byHash map[string]*domain.LoginSession
}

func NewMemorySessionRepository() *MemorySessionRepository {
	return &MemorySessionRepository{byHash: make(map[string]*domain.LoginSession)}
}

func (r *MemorySessionRepository) Create(_ context.Context, session *domain.LoginSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copy := *session
	r.byHash[session.RefreshTokenHash] = &copy
	return nil
}

func (r *MemorySessionRepository) GetByRefreshTokenHash(_ context.Context, hash string) (*domain.LoginSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, ok := r.byHash[hash]
	if !ok {
		return nil, ErrNotFound
	}
	copy := *session
	return &copy, nil
}

func (r *MemorySessionRepository) DeleteByRefreshTokenHash(_ context.Context, hash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.byHash, hash)
	return nil
}
