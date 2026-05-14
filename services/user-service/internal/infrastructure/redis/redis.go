package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	redisgo "github.com/redis/go-redis/v9"
)

const emailVerificationPrefix = "email_verification"

var ErrTokenNotFound = errors.New("token not found")

type Client struct {
	db *redisgo.Client
}

func New(addr string) *Client {
	return &Client{
		db: redisgo.NewClient(&redisgo.Options{Addr: addr}),
	}
}

func (c *Client) Ping(ctx context.Context) error {
	return c.db.Ping(ctx).Err()
}

func (c *Client) Close() error {
	return c.db.Close()
}

func (c *Client) SaveEmailVerification(ctx context.Context, token, userID string, ttl time.Duration) error {
	return c.db.Set(ctx, emailVerificationKey(token), userID, ttl).Err()
}

func (c *Client) TakeEmailVerification(ctx context.Context, token string) (string, error) {
	userID, err := c.db.GetDel(ctx, emailVerificationKey(token)).Result()
	if errors.Is(err, redisgo.Nil) {
		return "", ErrTokenNotFound
	}
	if err != nil {
		return "", err
	}
	return userID, nil
}

func emailVerificationKey(token string) string {
	return fmt.Sprintf("%s:%s", emailVerificationPrefix, token)
}
