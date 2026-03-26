package bootstrap

import (
	"barter-port/pkg/mailer"
	"fmt"
	"strings"
)

func InitMailerFromConfig(cfg Config) *mailer.SMTPMailer {
	mode := mailer.TLSMode(strings.ToLower(cfg.Mailer.TLSMode))
	if mode == "" {
		mode = mailer.TLSModeStartTLS
	}

	return mailer.NewSMTPMailer(
		cfg.Mailer.Host,
		cfg.Mailer.Port,
		cfg.Mailer.User,
		cfg.Mailer.Password,
		cfg.Mailer.From,
		mode,
		cfg.Mailer.InsecureSkipVerify,
	)
}

func ValidateMailConfig(cfg Config) error {
	if cfg.Mailer.Host == "" {
		return fmt.Errorf("mail.host is required")
	}
	if cfg.Mailer.Port == 0 {
		return fmt.Errorf("mail.port is required")
	}
	if cfg.Mailer.From == "" {
		return fmt.Errorf("mail.from is required")
	}
	if cfg.Mailer.User != "" && cfg.Mailer.Password == "" {
		return fmt.Errorf("SMTP_PASSWORD is required when mail.username is set")
	}
	return nil
}
