package bootstrap

import (
	"barter-port/internal/libs/platform/mailer"
	"log"
)

func InitMailer() *mailer.SMTPMailer {
	smtpHost := GetEnv("SMTP_HOST", "")
	smtpPort := mustInt(GetEnv("SMTP_PORT", ""))
	smtpUser := GetEnv("SMTP_USER", "")
	smtpPass := GetEnv("SMTP_PASS", "")
	smtpFrom := GetEnv("SMTP_FROM", smtpUser)

	if smtpHost == "" {
		log.Fatal("SMTP_HOST is required")
	}

	return mailer.NewSMTPMailer(smtpHost, smtpPort, smtpUser, smtpPass, smtpFrom)
}
