package email

import (
	"context"
	"fmt"
	"net/smtp"
	"readmeow/internal/config"

	"github.com/jordan-wright/email"
)

type EmailSender interface {
	SendMessage(ctx context.Context, sub string, content []byte, to []string, attachFiles []string) error
}

type emailSender struct {
	Auth        smtp.Auth
	EmailConfig config.EmailConfig
}

func NewEmailSender(a smtp.Auth, cfg config.EmailConfig) EmailSender {
	return &emailSender{
		Auth:        a,
		EmailConfig: cfg,
	}
}

func (es *emailSender) SendMessage(ctx context.Context, sub string, content []byte, to []string, attachFiles []string) error {
	op := "emailSender.SendMessage"
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", es.EmailConfig.Name, es.EmailConfig.Address)
	e.Subject = sub
	e.HTML = content
	e.To = to

	for _, f := range attachFiles {
		if _, err := e.AttachFile(f); err != nil {
			return fmt.Errorf("%s : %w", op, err)
		}
	}

	done := make(chan error, 1)
	go func() {
		defer close(done)
		done <- e.Send(es.EmailConfig.SmtpServerAddress, es.Auth)
	}()
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("%s : %w", op, err)
		}
	case <-ctx.Done():
		return fmt.Errorf("%s : %w", op, ctx.Err())
	}
	return nil
}
