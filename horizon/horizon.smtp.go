/*
	err := smtp.Send(&SMTPRequest{
		To:      "recipient@example.com",
		Subject: "Test Email",
		Body:    `<h1>Hello {{ index . "username" }}!</h1><p>This is a {{ index . "test_value" }} test email.</p>`,
		Vars: &map[string]string{
			"username":   "John Doe",
			"test_value": "TestValue",
		},
	})
*/
package horizon

import (
	"bytes"
	"fmt"
	"html/template"
	"net/mail"
	"net/smtp"

	"go.uber.org/zap"
)

type SMTPRequest struct {
	To      string             `json:"to"`
	Subject string             `json:"subject"`
	Body    string             `json:"body"`
	Vars    *map[string]string `json:"vars,omitempty"`
}

type HorizonSMTP struct {
	config   *HorizonConfig
	log      *HorizonLog
	security *HorizonSecurity
	cache    *HorizonCache
}

func NewHorizonSMTP(
	config *HorizonConfig,
	log *HorizonLog,
	security *HorizonSecurity,
	cache *HorizonCache,
) (*HorizonSMTP, error) {
	return &HorizonSMTP{
		config:   config,
		log:      log,
		security: security,
		cache:    cache,
	}, nil
}

func (hs *HorizonSMTP) Send(req *SMTPRequest) error {
	// Validate the 'To' email address
	if _, err := mail.ParseAddress(req.To); err != nil {
		hs.log.Log(LogEntry{
			Category: CategorySMTP,
			Level:    LevelError,
			Message:  fmt.Sprintf("Invalid email address: %v", err),
			Fields: []zap.Field{
				zap.String("to", req.To),
			},
		})
		return fmt.Errorf("invalid email address: %s", req.To)
	}

	// Process template variables in body
	bodyWithVars, err := hs.injectVarsIntoBody(req.Body, req.Vars)
	if err != nil {
		hs.log.Log(LogEntry{
			Category: CategorySMTP,
			Level:    LevelError,
			Message:  fmt.Sprintf("Failed to inject vars into body: %v", err),
			Fields:   []zap.Field{zap.String("body", req.Body)},
		})
		return fmt.Errorf("failed to inject vars into body: %v", err)
	}

	// Sanitize the final body
	cleaned := hs.security.SanitizeHTML(bodyWithVars)

	// Prepare SMTP authentication
	var auth smtp.Auth
	if hs.config.SMTPUsername != "" && hs.config.SMTPPassword != "" {
		auth = smtp.PlainAuth("", hs.config.SMTPUsername, hs.config.SMTPPassword, hs.config.SMTPHost)
	}

	// Ensure 'From' field is set correctly in headers
	from := hs.config.SMTPFrom
	if from == "" {
		// If no From is provided in the config, you might want to set a default or log an error
		hs.log.Log(LogEntry{
			Category: CategorySMTP,
			Level:    LevelError,
			Message:  "No 'From' address configured in SMTP config",
		})
		return fmt.Errorf("no 'From' address configured in SMTP config")
	}

	// Format the email message with From, To, Subject, and Body
	addr := fmt.Sprintf("%s:%d", hs.config.SMTPHost, hs.config.SMTPPort)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", from, req.To, req.Subject, cleaned)

	// Send the email
	err = smtp.SendMail(addr, auth, from, []string{req.To}, []byte(msg))
	if err != nil {
		hs.log.Log(LogEntry{
			Category: CategorySMTP,
			Level:    LevelError,
			Message:  fmt.Sprintf("SMTP send failed: %v", err),
			Fields: []zap.Field{
				zap.String("to", req.To),
				zap.String("from", from),
				zap.String("subject", req.Subject),
				zap.String("body", cleaned),
			},
		})
		return err
	}

	// Log the successful email send
	hs.log.Log(LogEntry{
		Category: CategorySMTP,
		Level:    LevelInfo,
		Message:  "SMTP email sent successfully",
		Fields: []zap.Field{
			zap.String("to", req.To),
			zap.String("from", from),
			zap.String("subject", req.Subject),
			zap.String("body", req.Body),
		},
	})

	return nil
}

func (hs *HorizonSMTP) injectVarsIntoBody(body string, vars *map[string]string) (string, error) {
	if vars == nil {
		return body, nil
	}

	tmpl, err := template.New("email").Parse(body)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, *vars)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %v", err)
	}

	return buf.String(), nil
}
