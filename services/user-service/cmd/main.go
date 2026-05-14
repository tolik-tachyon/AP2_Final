package main

import (
	"context"
	"log"
	"net"
	"time"

	userpb "github.com/tolik-tachyon/AP2_Final/gen/go/user"
	"github.com/tolik-tachyon/AP2_Final/services/user-service/internal/config"
	"github.com/tolik-tachyon/AP2_Final/services/user-service/internal/infrastructure/db"
	infranats "github.com/tolik-tachyon/AP2_Final/services/user-service/internal/infrastructure/nats"
	infraredis "github.com/tolik-tachyon/AP2_Final/services/user-service/internal/infrastructure/redis"
	infrasmtp "github.com/tolik-tachyon/AP2_Final/services/user-service/internal/infrastructure/smtp"
	"github.com/tolik-tachyon/AP2_Final/services/user-service/internal/repository"
	transportgrpc "github.com/tolik-tachyon/AP2_Final/services/user-service/internal/transport/grpc"
	"github.com/tolik-tachyon/AP2_Final/services/user-service/internal/usecase"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if cfg.SMTPHost == "" || cfg.SMTPUser == "" || cfg.SMTPPass == "" {
		log.Fatal("SMTP_HOST, SMTP_USER and SMTP_PASS are required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	database, err := db.ConnectPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer database.Close()

	redisClient := infraredis.New(cfg.RedisAddr)
	if err := redisClient.Ping(ctx); err != nil {
		log.Fatalf("connect redis: %v", err)
	}
	defer redisClient.Close()

	natsPublisher, err := infranats.NewPublisher(cfg.NATSURL)
	if err != nil {
		log.Fatalf("connect nats: %v", err)
	}
	defer natsPublisher.Close()

	emailSender := infrasmtp.NewSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass)

	userRepo := repository.NewPostgresUserRepository(database)
	sessionRepo := repository.NewPostgresSessionRepository(database)
	passwordResetRepo := repository.NewPostgresPasswordResetRepository(database)

	authUsecase := usecase.NewAuthUsecase(userRepo, sessionRepo, passwordResetRepo, redisClient, natsPublisher, emailSender)
	userUsecase := usecase.NewUserUsecase(userRepo, natsPublisher)
	userHandler := transportgrpc.NewUserHandler(authUsecase, userUsecase)

	server := grpc.NewServer()
	userpb.RegisterUserServiceServer(server, userHandler)

	listener, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	log.Printf("user-service gRPC server listening on %s", cfg.GRPCAddr)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
