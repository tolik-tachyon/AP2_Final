package smtp

import (
	"fmt"
	"net/smtp"
	"strings"
)

type Sender struct {
	Host string
	Port string
	User string
	Pass string
}

func NewSender(host, port, user, pass string) *Sender {
	return &Sender{Host: host, Port: port, User: user, Pass: pass}
}

func (s *Sender) SendVerificationEmail(to string, token string) error {
	subject := "Verify your Fandom Forum account"
	body := fmt.Sprintf("Your email verification token is: %s\n\nIf you did not create an account, ignore this email.", token)
	return s.send(to, subject, body)
}

func (s *Sender) SendPasswordResetEmail(to string, token string) error {
	subject := "Reset your Fandom Forum password"
	body := fmt.Sprintf("Your password reset token is: %s\n\nThis token expires in 30 minutes. If you did not request it, ignore this email.", token)
	return s.send(to, subject, body)
}

func (s *Sender) send(to string, subject string, body string) error {
	if s.Host == "" || s.Port == "" || s.User == "" || s.Pass == "" {
		return fmt.Errorf("smtp config is incomplete")
	}

	addr := fmt.Sprintf("%s:%s", s.Host, s.Port)
	auth := smtp.PlainAuth("", s.User, s.Pass, s.Host)
	message := strings.Join([]string{
		fmt.Sprintf("From: %s", s.User),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	return smtp.SendMail(addr, auth, s.User, []string{to}, []byte(message))
}
