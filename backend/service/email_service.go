package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jordan-wright/email"
	"net/smtp"
	"os"
	"strconv"
	"strings"
)

type EmailService struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

func NewEmailService() *EmailService {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	return &EmailService{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     port,
		User:     os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASS"),
		From:     os.Getenv("SMTP_FROM"),
	}
}

func (s *EmailService) SendWithAttachment(ctx context.Context, to string, subject string, body string, fileName string, attachment []byte) error {
	if s.Host == "" || s.User == "" {
		logger.WarnCtx(ctx, "Email service not configured, skipping email to %s", to)
		return nil
	}

	e := email.NewEmail()
	e.From = s.From
	e.To = []string{to}
	e.Subject = subject
	e.HTML = []byte(body)

	if len(attachment) > 0 {
		_, err := e.Attach(strings.NewReader(string(attachment)), fileName, "application/pdf")
		if err != nil {
			return fmt.Errorf("failed to attach file: %v", err)
		}
	}

	auth := smtp.PlainAuth("", s.User, s.Password, s.Host)
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)

	var err error
	if s.Port == 465 {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         s.Host,
		}
		err = e.SendWithTLS(addr, auth, tlsConfig)
	} else {
		err = e.Send(addr, auth)
	}

	if err != nil {
		logger.ErrorCtx(ctx, "Failed to send email to %s: %v", to, err)
		return err
	}

	logger.InfoCtx(ctx, "Email sent successfully to %s", to)
	return nil
}
