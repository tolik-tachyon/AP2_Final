package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GRPCAddr    string
	DatabaseURL string
	RedisAddr   string
	NATSURL     string
	SMTPHost    string
	SMTPPort    string
	SMTPUser    string
	SMTPPass    string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		GRPCAddr:    getEnv("USER_SERVICE_GRPC_ADDR", ":50051"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		NATSURL:     getEnv("NATS_URL", "nats://localhost:4222"),
		SMTPHost:    getEnv("SMTP_HOST", ""),
		SMTPPort:    getEnv("SMTP_PORT", "587"),
		SMTPUser:    getEnv("SMTP_USER", ""),
		SMTPPass:    getEnv("SMTP_PASS", ""),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
