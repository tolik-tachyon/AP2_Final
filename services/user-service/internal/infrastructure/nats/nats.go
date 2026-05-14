package nats

import (
	"context"
	"encoding/json"

	natsgo "github.com/nats-io/nats.go"
)

type Publisher struct {
	conn *natsgo.Conn
}

func NewPublisher(url string) (*Publisher, error) {
	conn, err := natsgo.Connect(url)
	if err != nil {
		return nil, err
	}
	return &Publisher{conn: conn}, nil
}

func (p *Publisher) Publish(_ context.Context, subject string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return p.conn.Publish(subject, data)
}

func (p *Publisher) Close() {
	p.conn.Drain()
	p.conn.Close()
}
