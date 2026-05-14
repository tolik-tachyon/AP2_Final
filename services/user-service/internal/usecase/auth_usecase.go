package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/tolik-tachyon/AP2_Final/services/user-service/internal/domain"
	"github.com/tolik-tachyon/AP2_Final/services/user-service/internal/repository"
)

const emailVerificationTTL = 24 * time.Hour

type VerificationTokenStore interface {
	SaveEmailVerification(ctx context.Context, token, userID string, ttl time.Duration) error
	TakeEmailVerification(ctx context.Context, token string) (string, error)
}

type EventPublisher interface {
	Publish(ctx context.Context, subject string, payload any) error
}

type EmailSender interface {
	SendVerificationEmail(to string, token string) error
	SendPasswordResetEmail(to string, token string) error
}

type AuthUsecase struct {
	users              repository.UserRepository
	sessions           repository.SessionRepository
	passwordResets     repository.PasswordResetRepository
	verificationTokens VerificationTokenStore
	events             EventPublisher
	emails             EmailSender
}

func NewAuthUsecase(
	users repository.UserRepository,
	sessions repository.SessionRepository,
	passwordResets repository.PasswordResetRepository,
	verificationTokens VerificationTokenStore,
	events EventPublisher,
	emails EmailSender,
) *AuthUsecase {
	return &AuthUsecase{
		users:              users,
		sessions:           sessions,
		passwordResets:     passwordResets,
		verificationTokens: verificationTokens,
		events:             events,
		emails:             emails,
	}
}

func (u *AuthUsecase) Register(ctx context.Context, username, email, password string) (*domain.User, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:           uuid.NewString(),
		Username:     username,
		Email:        email,
		PasswordHash: string(passwordHash),
		Role:         "user",
		IsVerified:   false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := u.users.Create(ctx, user); err != nil {
		return nil, err
	}

	verificationToken := uuid.NewString()
	if err := u.verificationTokens.SaveEmailVerification(ctx, verificationToken, user.ID, emailVerificationTTL); err != nil {
		return nil, err
	}

	if err := u.emails.SendVerificationEmail(user.Email, verificationToken); err != nil {
		return nil, err
	}

	if err := u.events.Publish(ctx, "user.registered", map[string]any{
		"user_id":            user.ID,
		"username":           user.Username,
		"email":              user.Email,
		"verification_token": verificationToken,
		"created_at":         user.CreatedAt,
	}); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *AuthUsecase) Login(ctx context.Context, email, password, userAgent, ipAddress string) (*domain.User, string, string, error) {
	user, err := u.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", "", errors.New("invalid credentials")
	}

	accessToken := uuid.NewString()
	refreshToken := uuid.NewString()

	session := &domain.LoginSession{
		ID:               uuid.NewString(),
		UserID:           user.ID,
		RefreshTokenHash: hashToken(refreshToken),
		UserAgent:        userAgent,
		IPAddress:        ipAddress,
		ExpiresAt:        time.Now().UTC().Add(30 * 24 * time.Hour),
		CreatedAt:        time.Now().UTC(),
	}

	if err := u.sessions.Create(ctx, session); err != nil {
		return nil, "", "", err
	}

	if err := u.events.Publish(ctx, "user.logged_in", map[string]any{
		"user_id":    user.ID,
		"user_agent": userAgent,
		"ip_address": ipAddress,
		"created_at": session.CreatedAt,
	}); err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (u *AuthUsecase) Logout(ctx context.Context, refreshToken string) error {
	return u.sessions.DeleteByRefreshTokenHash(ctx, hashToken(refreshToken))
}

func (u *AuthUsecase) RefreshSession(ctx context.Context, refreshToken string) (string, string, error) {
	session, err := u.sessions.GetByRefreshTokenHash(ctx, hashToken(refreshToken))
	if err != nil {
		return "", "", err
	}
	if time.Now().UTC().After(session.ExpiresAt) {
		return "", "", errors.New("session expired")
	}

	newAccessToken := uuid.NewString()
	newRefreshToken := uuid.NewString()

	if err := u.sessions.DeleteByRefreshTokenHash(ctx, hashToken(refreshToken)); err != nil {
		return "", "", err
	}

	session.ID = uuid.NewString()
	session.RefreshTokenHash = hashToken(newRefreshToken)
	session.CreatedAt = time.Now().UTC()
	session.ExpiresAt = time.Now().UTC().Add(30 * 24 * time.Hour)

	if err := u.sessions.Create(ctx, session); err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (u *AuthUsecase) VerifyEmail(ctx context.Context, token string) error {
	userID, err := u.verificationTokens.TakeEmailVerification(ctx, token)
	if err != nil {
		return err
	}

	user, err := u.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.IsVerified = true
	user.UpdatedAt = time.Now().UTC()
	if err := u.users.Update(ctx, user); err != nil {
		return err
	}

	return u.events.Publish(ctx, "user.email_verified", map[string]any{
		"user_id":     user.ID,
		"email":       user.Email,
		"verified_at": user.UpdatedAt,
	})
}

func (u *AuthUsecase) RequestPasswordReset(ctx context.Context, email string) (string, error) {
	user, err := u.users.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	plainToken := uuid.NewString()
	now := time.Now().UTC()
	resetToken := &domain.PasswordResetToken{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		TokenHash: hashToken(plainToken),
		ExpiresAt: now.Add(30 * time.Minute),
		CreatedAt: now,
	}

	if err := u.passwordResets.Create(ctx, resetToken); err != nil {
		return "", err
	}

	if err := u.emails.SendPasswordResetEmail(user.Email, plainToken); err != nil {
		return "", err
	}

	if err := u.events.Publish(ctx, "user.password_reset_requested", map[string]any{
		"user_id":     user.ID,
		"email":       user.Email,
		"reset_token": plainToken,
		"expires_at":  resetToken.ExpiresAt,
	}); err != nil {
		return "", err
	}

	return plainToken, nil
}

func (u *AuthUsecase) ResetPassword(ctx context.Context, token, newPassword string) error {
	resetToken, err := u.passwordResets.GetByTokenHash(ctx, hashToken(token))
	if err != nil {
		return err
	}
	if resetToken.UsedAt != nil {
		return errors.New("password reset token already used")
	}
	if time.Now().UTC().After(resetToken.ExpiresAt) {
		return errors.New("password reset token expired")
	}

	user, err := u.users.GetByID(ctx, resetToken.UserID)
	if err != nil {
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(passwordHash)
	user.UpdatedAt = time.Now().UTC()
	if err := u.users.Update(ctx, user); err != nil {
		return err
	}

	if err := u.passwordResets.MarkUsed(ctx, resetToken.ID); err != nil {
		return err
	}

	return u.events.Publish(ctx, "user.password_reset_completed", map[string]any{
		"user_id":  user.ID,
		"reset_at": user.UpdatedAt,
	})
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
