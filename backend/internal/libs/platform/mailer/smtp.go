package mailer

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type TLSMode string

const (
	TLSModeNone     TLSMode = "none"
	TLSModeStartTLS TLSMode = "starttls"
	TLSModeTLS      TLSMode = "tls" // implicit TLS, обычно 465
)

type SMTPMailer struct {
	Host               string
	Port               int
	Username           string
	Password           string
	From               string
	TLSMode            TLSMode
	InsecureSkipVerify bool
	HelloHost          string
	Timeout            time.Duration
}

func NewSMTPMailer(
	host string,
	port int,
	username, password, from string,
	tlsMode TLSMode,
	insecureSkipVerify bool,
) *SMTPMailer {
	return &SMTPMailer{
		Host:               host,
		Port:               port,
		Username:           username,
		Password:           password,
		From:               from,
		TLSMode:            tlsMode,
		InsecureSkipVerify: insecureSkipVerify,
		HelloHost:          "localhost",
		Timeout:            10 * time.Second,
	}
}

func (m *SMTPMailer) Send(to, subject, body string) error {
	if m.Host == "" {
		return errors.New("smtp host is required")
	}
	if m.Port <= 0 {
		return errors.New("smtp port must be > 0")
	}
	if m.From == "" {
		return errors.New("smtp from is required")
	}
	if to == "" {
		return errors.New("recipient is required")
	}

	msg := buildMessage(m.From, to, subject, body)
	addr := fmt.Sprintf("%s:%d", m.Host, m.Port)

	switch m.TLSMode {
	case TLSModeTLS:
		return m.sendImplicitTLS(addr, to, msg)
	case TLSModeStartTLS:
		return m.sendStartTLS(addr, to, msg)
	case TLSModeNone:
		return m.sendPlain(addr, to, msg)
	default:
		return fmt.Errorf("unsupported tls mode: %q", m.TLSMode)
	}
}

func (m *SMTPMailer) sendImplicitTLS(addr, to string, msg []byte) error {
	tlsConfig := m.tlsConfig()

	dialer := &net.Dialer{Timeout: m.Timeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("smtp tls dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.Host)
	if err != nil {
		return fmt.Errorf("smtp new client: %w", err)
	}
	defer client.Close()

	if err := m.hello(client); err != nil {
		return err
	}
	if err := m.auth(client); err != nil {
		return err
	}
	if err := m.sendData(client, to, msg); err != nil {
		return err
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("smtp quit: %w", err)
	}
	return nil
}

func (m *SMTPMailer) sendStartTLS(addr, to string, msg []byte) error {
	dialer := &net.Dialer{Timeout: m.Timeout}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.Host)
	if err != nil {
		return fmt.Errorf("smtp new client: %w", err)
	}
	defer client.Close()

	if err := m.hello(client); err != nil {
		return err
	}

	ok, _ := client.Extension("STARTTLS")
	if !ok {
		return errors.New("smtp server does not support STARTTLS")
	}

	if err := client.StartTLS(m.tlsConfig()); err != nil {
		return fmt.Errorf("smtp starttls: %w", err)
	}

	if err := m.auth(client); err != nil {
		return err
	}
	if err := m.sendData(client, to, msg); err != nil {
		return err
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("smtp quit: %w", err)
	}
	return nil
}

func (m *SMTPMailer) sendPlain(addr, to string, msg []byte) error {
	dialer := &net.Dialer{Timeout: m.Timeout}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.Host)
	if err != nil {
		return fmt.Errorf("smtp new client: %w", err)
	}
	defer client.Close()

	if err := m.hello(client); err != nil {
		return err
	}
	if err := m.auth(client); err != nil {
		return err
	}
	if err := m.sendData(client, to, msg); err != nil {
		return err
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("smtp quit: %w", err)
	}
	return nil
}

func (m *SMTPMailer) hello(client *smtp.Client) error {
	if err := client.Hello(m.HelloHost); err != nil {
		return fmt.Errorf("smtp hello: %w", err)
	}
	return nil
}

func (m *SMTPMailer) auth(client *smtp.Client) error {
	if m.Username == "" {
		return nil
	}

	ok, _ := client.Extension("AUTH")
	if !ok {
		return errors.New("smtp server does not support AUTH")
	}

	auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	return nil
}

func (m *SMTPMailer) sendData(client *smtp.Client, to string, msg []byte) error {
	if err := client.Mail(m.From); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	if _, err := w.Write(msg); err != nil {
		_ = w.Close()
		return fmt.Errorf("smtp write message: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close writer: %w", err)
	}

	return nil
}

func (m *SMTPMailer) tlsConfig() *tls.Config {
	return &tls.Config{
		ServerName:         m.Host,
		InsecureSkipVerify: m.InsecureSkipVerify,
		MinVersion:         tls.VersionTLS12,
	}
}

func buildMessage(from, to, subject, body string) []byte {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("From: %s\r\n", from))
	b.WriteString(fmt.Sprintf("To: %s\r\n", to))
	b.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	b.WriteString("\r\n")
	return []byte(b.String())
}
