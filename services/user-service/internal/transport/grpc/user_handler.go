package grpc

import (
	"context"

	userpb "github.com/tolik-tachyon/AP2_Final/gen/go/user"
	"github.com/tolik-tachyon/AP2_Final/services/user-service/internal/domain"
	"github.com/tolik-tachyon/AP2_Final/services/user-service/internal/usecase"
)

type UserHandler struct {
	userpb.UnimplementedUserServiceServer

	authUsecase *usecase.AuthUsecase
	userUsecase *usecase.UserUsecase
}

func NewUserHandler(authUsecase *usecase.AuthUsecase, userUsecase *usecase.UserUsecase) *UserHandler {
	return &UserHandler{
		authUsecase: authUsecase,
		userUsecase: userUsecase,
	}
}

func (h *UserHandler) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {
	user, err := h.authUsecase.Register(ctx, req.GetUsername(), req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, err
	}
	return &userpb.RegisterResponse{User: toProtoUser(user)}, nil
}

func (h *UserHandler) VerifyEmail(ctx context.Context, req *userpb.VerifyEmailRequest) (*userpb.VerifyEmailResponse, error) {
	if err := h.authUsecase.VerifyEmail(ctx, req.GetToken()); err != nil {
		return nil, err
	}
	return &userpb.VerifyEmailResponse{Success: true}, nil
}

func (h *UserHandler) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	user, accessToken, refreshToken, err := h.authUsecase.Login(
		ctx,
		req.GetEmail(),
		req.GetPassword(),
		req.GetUserAgent(),
		req.GetIpAddress(),
	)
	if err != nil {
		return nil, err
	}

	return &userpb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         toProtoUser(user),
	}, nil
}

func (h *UserHandler) Logout(ctx context.Context, req *userpb.LogoutRequest) (*userpb.LogoutResponse, error) {
	if err := h.authUsecase.Logout(ctx, req.GetRefreshToken()); err != nil {
		return nil, err
	}
	return &userpb.LogoutResponse{Success: true}, nil
}

func (h *UserHandler) RefreshSession(ctx context.Context, req *userpb.RefreshSessionRequest) (*userpb.RefreshSessionResponse, error) {
	accessToken, refreshToken, err := h.authUsecase.RefreshSession(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, err
	}
	return &userpb.RefreshSessionResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	user, err := h.userUsecase.GetUser(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &userpb.GetUserResponse{User: toProtoUser(user)}, nil
}

func (h *UserHandler) GetUserByEmail(ctx context.Context, req *userpb.GetUserByEmailRequest) (*userpb.GetUserByEmailResponse, error) {
	user, err := h.userUsecase.GetUserByEmail(ctx, req.GetEmail())
	if err != nil {
		return nil, err
	}
	return &userpb.GetUserByEmailResponse{User: toProtoUser(user)}, nil
}

func (h *UserHandler) UpdateProfile(ctx context.Context, req *userpb.UpdateProfileRequest) (*userpb.UpdateProfileResponse, error) {
	user, err := h.userUsecase.UpdateProfile(
		ctx,
		req.GetId(),
		req.GetUsername(),
		req.GetAvatarUrl(),
		req.GetBio(),
	)
	if err != nil {
		return nil, err
	}
	return &userpb.UpdateProfileResponse{User: toProtoUser(user)}, nil
}

func (h *UserHandler) ChangePassword(ctx context.Context, req *userpb.ChangePasswordRequest) (*userpb.ChangePasswordResponse, error) {
	if err := h.userUsecase.ChangePassword(ctx, req.GetUserId(), req.GetOldPassword(), req.GetNewPassword()); err != nil {
		return nil, err
	}
	return &userpb.ChangePasswordResponse{Success: true}, nil
}

func (h *UserHandler) RequestPasswordReset(ctx context.Context, req *userpb.RequestPasswordResetRequest) (*userpb.RequestPasswordResetResponse, error) {
	resetToken, err := h.authUsecase.RequestPasswordReset(ctx, req.GetEmail())
	if err != nil {
		return nil, err
	}
	return &userpb.RequestPasswordResetResponse{Success: true, ResetToken: resetToken}, nil
}

func (h *UserHandler) ResetPassword(ctx context.Context, req *userpb.ResetPasswordRequest) (*userpb.ResetPasswordResponse, error) {
	if err := h.authUsecase.ResetPassword(ctx, req.GetToken(), req.GetNewPassword()); err != nil {
		return nil, err
	}
	return &userpb.ResetPasswordResponse{Success: true}, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.DeleteUserResponse, error) {
	if err := h.userUsecase.DeleteUser(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &userpb.DeleteUserResponse{Success: true}, nil
}

func toProtoUser(user *domain.User) *userpb.User {
	if user == nil {
		return nil
	}

	return &userpb.User{
		Id:         user.ID,
		Username:   user.Username,
		Email:      user.Email,
		AvatarUrl:  user.AvatarURL,
		Bio:        user.Bio,
		Role:       user.Role,
		IsVerified: user.IsVerified,
		CreatedAt:  user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
