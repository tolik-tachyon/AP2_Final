package domain

import "time"

type User struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
	AvatarURL    string
	Bio          string
	Role         string
	IsVerified   bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
