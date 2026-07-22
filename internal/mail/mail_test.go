package mail

import (
	"strings"
	"testing"
)

func TestNewSMTPRequiresHostAndPort(t *testing.T) {
	t.Parallel()

	if _, err := NewSMTP("", "1025"); err == nil {
		t.Fatal("want error for empty host")
	}
	if _, err := NewSMTP("mailpit", ""); err == nil {
		t.Fatal("want error for empty port")
	}
}

func TestNewSMTPOK(t *testing.T) {
	t.Parallel()

	m, err := NewSMTP("mailpit", "1025")
	if err != nil {
		t.Fatalf("NewSMTP: %v", err)
	}
	if m.addr != "mailpit:1025" {
		t.Fatalf("addr = %q, want mailpit:1025", m.addr)
	}
	if m.from != defaultFrom {
		t.Fatalf("from = %q, want %q", m.from, defaultFrom)
	}
}

func TestBuildMessage(t *testing.T) {
	t.Parallel()

	raw := string(buildMessage("from@example.com", "to@example.com", "Hello", "Body line"))
	for _, want := range []string{
		"From: from@example.com\r\n",
		"To: to@example.com\r\n",
		"Subject: Hello\r\n",
		"Content-Type: text/plain; charset=UTF-8\r\n",
		"\r\nBody line",
	} {
		if !strings.Contains(raw, want) {
			t.Fatalf("message missing %q; got %q", want, raw)
		}
	}
}

func TestSMTPMailerSendValidation(t *testing.T) {
	t.Parallel()

	m := &SMTPMailer{addr: "localhost:1025", from: defaultFrom}
	if err := m.Send(Message{Subject: "x", Body: "y"}); err == nil {
		t.Fatal("want error for empty recipient")
	}
	if err := m.Send(Message{To: "a@b.com", Body: "y"}); err == nil {
		t.Fatal("want error for empty subject")
	}
}

func TestNopMailerSend(t *testing.T) {
	t.Parallel()

	if err := (NopMailer{}).Send(Message{To: "a@b.com", Subject: "s", Body: "b"}); err != nil {
		t.Fatalf("NopMailer.Send: %v", err)
	}
}
