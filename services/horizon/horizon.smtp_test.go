package horizon

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// go test -v ./services/horizon/otp_test.go

func TestHorizonSMTP_Run_Stop(t *testing.T) {
	env := NewEnvironmentService("../../.env")

	host := env.GetString("SMTP_HOST", "")
	port := env.GetInt("SMTP_PORT", 0)
	username := env.GetString("SMTP_USERNAME", "")
	password := env.GetString("SMTP_PASSWORD", "")
	from := env.GetString("SMTP_FROM", "")

	smtp := NewSMTP(host, port, username, password, from)
	ctx := context.Background()

	require.NoError(t, smtp.Run(ctx))
	require.NoError(t, smtp.Stop(ctx))
}

func TestHorizonSMTP_Format_WithTemplateString(t *testing.T) {
	env := NewEnvironmentService("../../.env")

	host := env.GetString("SMTP_HOST", "")
	port := env.GetInt("SMTP_PORT", 0)
	username := env.GetString("SMTP_USERNAME", "")
	password := env.GetString("SMTP_PASSWORD", "")
	from := env.GetString("SMTP_FROM", "")
	reciever := env.GetString("SMTP_TEST_RECIEVER", "")

	smtp := NewSMTP(host, port, username, password, from)
	ctx := context.Background()

	req := SMTPRequest{
		To:      reciever,
		Subject: "Test Subject",
		Body:    "Hello {{.Name}}, welcome!",
		Vars:    map[string]string{"Name": "Alice"},
	}

	formatted, err := smtp.Format(ctx, req)
	require.NoError(t, err)
	assert.Contains(t, formatted.Body, "Hello Alice")
}

func TestHorizonSMTP_Format_WithTemplateFile(t *testing.T) {
	env := NewEnvironmentService("../../.env")

	host := env.GetString("SMTP_HOST", "")
	port := env.GetInt("SMTP_PORT", 0)
	username := env.GetString("SMTP_USERNAME", "")
	password := env.GetString("SMTP_PASSWORD", "")
	from := env.GetString("SMTP_FROM", "")
	reciever := env.GetString("SMTP_TEST_RECIEVER", "")

	file := "test_template.txt"
	content := "Hello {{.Name}}, this is from file."
	err := os.WriteFile(file, []byte(content), 0600)
	require.NoError(t, err)
	defer func() { _ = os.Remove(file) }()
	smtp := NewSMTP(host, port, username, password, from)
	ctx := context.Background()

	req := SMTPRequest{
		To:      reciever,
		Subject: "Test File",
		Body:    file,
		Vars:    map[string]string{"Name": "Bob"},
	}

	formatted, err := smtp.Format(ctx, req)
	require.NoError(t, err)
	assert.Contains(t, formatted.Body, "Hello Bob")
}

func TestHorizonSMTP_Send_InvalidEmail(t *testing.T) {
	env := NewEnvironmentService("../../.env")

	host := env.GetString("SMTP_HOST", "")
	port := env.GetInt("SMTP_PORT", 0)
	username := env.GetString("SMTP_USERNAME", "")
	password := env.GetString("SMTP_PASSWORD", "")
	from := env.GetString("SMTP_FROM", "")

	smtp := NewSMTP(host, port, username, password, from)
	ctx := context.Background()
	_ = smtp.Run(ctx)

	req := SMTPRequest{
		To:      "also-invalid",
		Subject: "Test",
		Body:    "Hello {{.Name}}",
		Vars:    map[string]string{"Name": "Test"},
	}

	err := smtp.Send(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "format is invalid")
}
