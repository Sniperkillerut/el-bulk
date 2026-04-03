package mailer

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/jordan-wright/email"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

func SendEmail(to string, subject string, bodyHTML string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")

	if host == "" || user == "" || pass == "" {
		logger.Warn("SMTP settings not configured, skipping email to %s", to)
		return nil
	}

	e := email.NewEmail()
	e.From = from
	e.To = []string{to}
	e.Subject = subject
	e.HTML = []byte(bodyHTML)

	addr := fmt.Sprintf("%s:%s", host, port)
	auth := smtp.PlainAuth("", user, pass, host)

	var err error
	if port == "465" {
		// Implicit TLS for port 465
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         host,
		}
		err = e.SendWithTLS(addr, auth, tlsConfig)
	} else {
		// Standard submission (e.g. 587) with STARTTLS if supported by library or Send
		// The jordan-wright/email library's Send method uses net/smtp.SendMail which handles STARTTLS.
		err = e.Send(addr, auth)
	}

	if err != nil {
		logger.Error("Failed to send email to %s: %v", to, err)
		return err
	}

	return nil
}

func BroadcastNotice(db *sqlx.DB, notice models.Notice) {
	logger.Info("Starting newsletter broadcast for notice: %s", notice.Title)

	var subscribers []string
	err := db.Select(&subscribers, "SELECT email FROM newsletter_subscriber")
	if err != nil {
		logger.Error("Failed to fetch newsletter subscribers: %v", err)
		return
	}

	siteURL := os.Getenv("SITE_URL")
	if siteURL == "" {
		siteURL = "http://localhost:3000"
	}

	noticeURL := fmt.Sprintf("%s/notices/%s", siteURL, notice.Slug)

	// Basic HTML Template
	template := `
	<!DOCTYPE html>
	<html>
	<head>
		<style>
			body { font-family: sans-serif; background-color: #f4efea; color: #2d2a26; padding: 20px; }
			.container { max-width: 600px; margin: 0 auto; background: white; border: 1px solid #d4c4b5; padding: 40px; }
			.header { border-bottom: 4px solid #8b0000; padding-bottom: 20px; margin-bottom: 30px; }
			.title { font-size: 24px; font-weight: bold; text-transform: uppercase; margin: 0; color: #1a1a1a; }
			.content { line-height: 1.6; font-size: 16px; margin-bottom: 30px; }
			.btn { display: inline-block; background-color: #d4af37; color: white; padding: 12px 24px; text-decoration: none; font-weight: bold; text-transform: uppercase; font-size: 14px; }
			.footer { margin-top: 40px; padding-top: 20px; border-top: 1px solid #eee; font-size: 12px; color: #888; text-align: center; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h1 class="title">NEWS FROM EL BULK SHOP</h1>
			</div>
			<div class="content">
				<h2>%s</h2>
				<p>We've just published a new update you might be interested in!</p>
				<a href="%s" class="btn">READ THE FULL STORY</a>
			</div>
			<div class="footer">
				<p>&copy; %d El Bulk Shop. All rights reserved.</p>
				<p>You are receiving this because you subscribed to our newsletter.</p>
				<p><a href="%s/unsubscribe">Unsubscribe</a></p>
			</div>
		</div>
	</body>
	</html>
	`

	year := time.Now().Year()
	body := fmt.Sprintf(template, notice.Title, noticeURL, year, siteURL)

	successCount := 0
	for _, email := range subscribers {
		if strings.Contains(email, "example.com") {
			// Skip seed/example emails to avoid spamming fake addresses
			continue
		}
		
		err = SendEmail(email, "New Update: "+notice.Title, body)
		if err == nil {
			successCount++
		}
	}

	logger.Info("Newsletter broadcast complete. Sent to %d subscribers.", successCount)
}
