package mail

import (
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

const defaultFrom = "noreply@forgeboard.local"

// Message is a plain-text email to send.
type Message struct {
	To      string
	Subject string
	Body    string
}

// Sender sends outbound email messages.
type Sender interface {
	Send(msg Message) error
}

// SMTPMailer delivers mail via SMTP (Mailpit in development).
type SMTPMailer struct {
	addr string
	from string
}

// NewSMTP constructs a mailer for SMTP_HOST:SMTP_PORT.
// No authentication is used — suitable for Mailpit and similar local relays.
func NewSMTP(host, port string) (*SMTPMailer, error) {
	host = strings.TrimSpace(host)
	port = strings.TrimSpace(port)
	if host == "" || port == "" {
		return nil, fmt.Errorf("mail: SMTP_HOST and SMTP_PORT are required")
	}
	return &SMTPMailer{
		addr: net.JoinHostPort(host, port),
		from: defaultFrom,
	}, nil
}

// Send delivers a plain-text message over SMTP.
func (m *SMTPMailer) Send(msg Message) error {
	to := strings.TrimSpace(msg.To)
	if to == "" {
		return fmt.Errorf("mail: recipient is required")
	}
	subject := strings.TrimSpace(msg.Subject)
	if subject == "" {
		return fmt.Errorf("mail: subject is required")
	}

	payload := buildMessage(m.from, to, subject, msg.Body)
	if err := smtp.SendMail(m.addr, nil, m.from, []string{to}, payload); err != nil {
		return fmt.Errorf("mail: send: %w", err)
	}
	return nil
}

func buildMessage(from, to, subject, body string) []byte {
	var b strings.Builder
	b.WriteString("From: ")
	b.WriteString(from)
	b.WriteString("\r\n")
	b.WriteString("To: ")
	b.WriteString(to)
	b.WriteString("\r\n")
	b.WriteString("Subject: ")
	b.WriteString(subject)
	b.WriteString("\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	if !strings.HasSuffix(body, "\n") {
		b.WriteString("\r\n")
	}
	return []byte(b.String())
}

// NopMailer discards messages (used in tests when SMTP is unset).
type NopMailer struct{}

// Send implements Sender.
func (NopMailer) Send(Message) error { return nil }
