package bootstrap

import (
	"barter-port/internal/libs/platform/mailer"
	"fmt"
)

func InitMailerFromConfig(cfg Config) (*mailer.SMTPMailer, error) {
	if cfg.Mailer.Host == "" {
		return nil, fmt.Errorf("MAILER_HOST is required")
	}

	return mailer.NewSMTPMailer(
		cfg.Mailer.Host,
		cfg.Mailer.Port,
		cfg.Mailer.User,
		cfg.Mailer.Password,
		cfg.Mailer.From,
	), nil
}
