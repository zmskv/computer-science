package email

import (
	"crypto/tls"
	"fmt"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/DelayedNotifier/internal/domain/interfaces"
	"gopkg.in/gomail.v2"
)

type EmailSender struct {
	dialer *gomail.Dialer
	from   string
}

func NewEmailSender(host, email, password string, port int) (interfaces.SMTPClient, error) {
	dialer := gomail.NewDialer(host, port, email, password)

	dialer.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	return &EmailSender{
		dialer: dialer,
		from:   email,
	}, nil
}

func (s *EmailSender) SendEmail(email string, subject string, body string) error {
	return s.send(email, subject, body)
}

func (s *EmailSender) send(address, subject, payload string) error {
	message := gomail.NewMessage()

	message.SetHeader("From", s.from)
	message.SetHeader("To", address)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", payload)

	if err := s.dialer.DialAndSend(message); err != nil {
		return fmt.Errorf("sending email: %v", err)
	}

	return nil
}
