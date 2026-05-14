package usecase

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/tolik-tachyon/AP2_Final/services/user-service/internal/domain"
	"github.com/tolik-tachyon/AP2_Final/services/user-service/internal/repository"
)

type UserUsecase struct {
	users  repository.UserRepository
	events EventPublisher
}

func NewUserUsecase(users repository.UserRepository, events EventPublisher) *UserUsecase {
	return &UserUsecase{users: users, events: events}
}

func (u *UserUsecase) GetUser(ctx context.Context, id string) (*domain.User, error) {
	return u.users.GetByID(ctx, id)
}

func (u *UserUsecase) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return u.users.GetByEmail(ctx, email)
}

func (u *UserUsecase) UpdateProfile(ctx context.Context, id, username, avatarURL, bio string) (*domain.User, error) {
	user, err := u.users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.Username = username
	user.AvatarURL = avatarURL
	user.Bio = bio
	user.UpdatedAt = time.Now().UTC()

	if err := u.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserUsecase) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := u.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(passwordHash)
	user.UpdatedAt = time.Now().UTC()
	return u.users.Update(ctx, user)
}

func (u *UserUsecase) DeleteUser(ctx context.Context, id string) error {
	user, err := u.users.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := u.users.Delete(ctx, id); err != nil {
		return err
	}

	return u.events.Publish(ctx, "user.deleted", map[string]any{
		"user_id":    user.ID,
		"email":      user.Email,
		"deleted_at": time.Now().UTC(),
	})
}
