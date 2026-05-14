package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/tolik-tachyon/AP2_Final/services/user-service/internal/repository"
)

type fakeVerificationStore struct {
	byToken map[string]string
}

func newFakeVerificationStore() *fakeVerificationStore {
	return &fakeVerificationStore{byToken: make(map[string]string)}
}

func (s *fakeVerificationStore) SaveEmailVerification(_ context.Context, token, userID string, _ time.Duration) error {
	s.byToken[token] = userID
	return nil
}

func (s *fakeVerificationStore) TakeEmailVerification(_ context.Context, token string) (string, error) {
	userID, ok := s.byToken[token]
	if !ok {
		return "", repository.ErrNotFound
	}
	delete(s.byToken, token)
	return userID, nil
}

type publishedEvent struct {
	subject string
	payload any
}

type fakeEventPublisher struct {
	events []publishedEvent
}

func (p *fakeEventPublisher) Publish(_ context.Context, subject string, payload any) error {
	p.events = append(p.events, publishedEvent{subject: subject, payload: payload})
	return nil
}

func (p *fakeEventPublisher) hasSubject(subject string) bool {
	for _, event := range p.events {
		if event.subject == subject {
			return true
		}
	}
	return false
}

type fakeEmailSender struct {
	verificationTo    string
	verificationToken string
	resetTo           string
	resetToken        string
}

func (s *fakeEmailSender) SendVerificationEmail(to string, token string) error {
	s.verificationTo = to
	s.verificationToken = token
	return nil
}

func (s *fakeEmailSender) SendPasswordResetEmail(to string, token string) error {
	s.resetTo = to
	s.resetToken = token
	return nil
}

type authTestDeps struct {
	users         *repository.MemoryUserRepository
	sessions      *repository.MemorySessionRepository
	passwordReset *repository.MemoryPasswordResetRepository
	verification  *fakeVerificationStore
	events        *fakeEventPublisher
	emails        *fakeEmailSender
	auth          *AuthUsecase
	user          *UserUsecase
}

func newAuthTestDeps() authTestDeps {
	users := repository.NewMemoryUserRepository()
	sessions := repository.NewMemorySessionRepository()
	passwordReset := repository.NewMemoryPasswordResetRepository()
	verification := newFakeVerificationStore()
	events := &fakeEventPublisher{}
	emails := &fakeEmailSender{}

	return authTestDeps{
		users:         users,
		sessions:      sessions,
		passwordReset: passwordReset,
		verification:  verification,
		events:        events,
		emails:        emails,
		auth:          NewAuthUsecase(users, sessions, passwordReset, verification, events, emails),
		user:          NewUserUsecase(users, events),
	}
}

func TestRegisterCreatesUserAndSendsVerification(t *testing.T) {
	deps := newAuthTestDeps()
	ctx := context.Background()

	user, err := deps.auth.Register(ctx, "tachyon", "tachyon@example.com", "strong-pass")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if user.ID == "" {
		t.Fatal("expected generated user id")
	}
	if user.Email != "tachyon@example.com" {
		t.Fatalf("unexpected email: %s", user.Email)
	}
	if user.IsVerified {
		t.Fatal("new user should not be verified")
	}
	if user.PasswordHash == "strong-pass" {
		t.Fatal("password should be hashed")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("strong-pass")); err != nil {
		t.Fatalf("password hash does not match original password: %v", err)
	}
	if deps.emails.verificationTo != user.Email {
		t.Fatalf("verification email sent to %q, want %q", deps.emails.verificationTo, user.Email)
	}
	if deps.emails.verificationToken == "" {
		t.Fatal("expected verification token to be emailed")
	}
	if !deps.events.hasSubject("user.registered") {
		t.Fatal("expected user.registered event")
	}
}

func TestVerifyEmailMarksUserVerified(t *testing.T) {
	deps := newAuthTestDeps()
	ctx := context.Background()

	user, err := deps.auth.Register(ctx, "tachyon", "tachyon@example.com", "strong-pass")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if err := deps.auth.VerifyEmail(ctx, deps.emails.verificationToken); err != nil {
		t.Fatalf("VerifyEmail returned error: %v", err)
	}

	updated, err := deps.users.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if !updated.IsVerified {
		t.Fatal("expected user to be verified")
	}
	if _, err := deps.verification.TakeEmailVerification(ctx, deps.emails.verificationToken); err == nil {
		t.Fatal("verification token should be consumed")
	}
	if !deps.events.hasSubject("user.email_verified") {
		t.Fatal("expected user.email_verified event")
	}
}

func TestLoginCreatesSession(t *testing.T) {
	deps := newAuthTestDeps()
	ctx := context.Background()

	user, err := deps.auth.Register(ctx, "tachyon", "tachyon@example.com", "strong-pass")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	loggedIn, accessToken, refreshToken, err := deps.auth.Login(ctx, user.Email, "strong-pass", "test-agent", "127.0.0.1")
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if loggedIn.ID != user.ID {
		t.Fatalf("logged in user id = %q, want %q", loggedIn.ID, user.ID)
	}
	if accessToken == "" || refreshToken == "" {
		t.Fatal("expected access and refresh tokens")
	}

	if _, err := deps.sessions.GetByRefreshTokenHash(ctx, hashForTest(refreshToken)); err != nil {
		t.Fatalf("expected session stored by refresh token hash: %v", err)
	}
	if !deps.events.hasSubject("user.logged_in") {
		t.Fatal("expected user.logged_in event")
	}
}

func TestLoginRejectsInvalidPassword(t *testing.T) {
	deps := newAuthTestDeps()
	ctx := context.Background()

	_, err := deps.auth.Register(ctx, "tachyon", "tachyon@example.com", "strong-pass")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	_, _, _, err = deps.auth.Login(ctx, "tachyon@example.com", "wrong-pass", "test-agent", "127.0.0.1")
	if err == nil {
		t.Fatal("expected invalid password error")
	}
}

func TestPasswordResetChangesPassword(t *testing.T) {
	deps := newAuthTestDeps()
	ctx := context.Background()

	user, err := deps.auth.Register(ctx, "tachyon", "tachyon@example.com", "strong-pass")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	resetToken, err := deps.auth.RequestPasswordReset(ctx, user.Email)
	if err != nil {
		t.Fatalf("RequestPasswordReset returned error: %v", err)
	}
	if resetToken == "" || deps.emails.resetToken != resetToken {
		t.Fatal("expected reset token to be returned and emailed")
	}

	if err := deps.auth.ResetPassword(ctx, resetToken, "new-strong-pass"); err != nil {
		t.Fatalf("ResetPassword returned error: %v", err)
	}

	updated, err := deps.users.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(updated.PasswordHash), []byte("new-strong-pass")); err != nil {
		t.Fatalf("new password hash does not match: %v", err)
	}
	if !deps.events.hasSubject("user.password_reset_requested") {
		t.Fatal("expected user.password_reset_requested event")
	}
	if !deps.events.hasSubject("user.password_reset_completed") {
		t.Fatal("expected user.password_reset_completed event")
	}
}

func TestDeleteUserPublishesEvent(t *testing.T) {
	deps := newAuthTestDeps()
	ctx := context.Background()

	user, err := deps.auth.Register(ctx, "tachyon", "tachyon@example.com", "strong-pass")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if err := deps.user.DeleteUser(ctx, user.ID); err != nil {
		t.Fatalf("DeleteUser returned error: %v", err)
	}
	if _, err := deps.users.GetByID(ctx, user.ID); err == nil {
		t.Fatal("expected user to be deleted")
	}
	if !deps.events.hasSubject("user.deleted") {
		t.Fatal("expected user.deleted event")
	}
}

func TestChangePasswordUpdatesPassword(t *testing.T) {
	deps := newAuthTestDeps()
	ctx := context.Background()

	user, err := deps.auth.Register(ctx, "tachyon", "tachyon@example.com", "strong-pass")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if err := deps.user.ChangePassword(ctx, user.ID, "strong-pass", "new-strong-pass"); err != nil {
		t.Fatalf("ChangePassword returned error: %v", err)
	}

	updated, err := deps.users.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(updated.PasswordHash), []byte("new-strong-pass")); err != nil {
		t.Fatalf("new password hash does not match: %v", err)
	}
}

func TestUpdateProfileChangesUserFields(t *testing.T) {
	deps := newAuthTestDeps()
	ctx := context.Background()

	user, err := deps.auth.Register(ctx, "tachyon", "tachyon@example.com", "strong-pass")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	updated, err := deps.user.UpdateProfile(ctx, user.ID, "new-name", "https://example.com/avatar.png", "new bio")
	if err != nil {
		t.Fatalf("UpdateProfile returned error: %v", err)
	}
	if updated.Username != "new-name" {
		t.Fatalf("username = %q, want new-name", updated.Username)
	}
	if updated.AvatarURL != "https://example.com/avatar.png" {
		t.Fatalf("avatar url = %q", updated.AvatarURL)
	}
	if updated.Bio != "new bio" {
		t.Fatalf("bio = %q, want new bio", updated.Bio)
	}
}

func hashForTest(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
