/*
	err := smtp.Send(&SMTPRequest{
		To:      "recipient@example.com",
		Subject: "Test Email",
		Body:    `<h1>Hello {{ .username }}!</h1><p>This is a {{ .test_value }} test email.</p>`,
		Vars: &map[string]string{
			"username":   "John Doe",
			"test_value": "Rate Limited",
		},
	})
*/
package horizon

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/mail"
	"net/smtp"
	"sync"

	"github.com/rotisserie/eris"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
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

	limiter     *rate.Limiter
	limiterOnce sync.Once
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewHorizonSMTP(
	config *HorizonConfig,
	log *HorizonLog,
	security *HorizonSecurity,
	cache *HorizonCache,
) (*HorizonSMTP, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &HorizonSMTP{
		config:   config,
		log:      log,
		security: security,
		cache:    cache,

		limiter: rate.NewLimiter(1, 3),
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}
func (hs *HorizonSMTP) run() {
	hs.limiterOnce.Do(func() {
		hs.limiter = rate.NewLimiter(1, 3)
	})
}

func (hs *HorizonSMTP) stop() {
	hs.log.Log(LogEntry{
		Category: CategorySMTP,
		Level:    LevelInfo,
		Message:  "Stopping HorizonSMTP service gracefully...",
	})
	hs.cancel()
}
func (hs *HorizonSMTP) Send(req *SMTPRequest) error {
	// Validate 'To' email address
	if _, err := mail.ParseAddress(req.To); err != nil {
		hs.log.Log(LogEntry{
			Category: CategorySMTP,
			Level:    LevelError,
			Message:  fmt.Sprintf("Invalid email address: %v", err),
			Fields:   []zap.Field{zap.String("to", req.To)},
		})
		return eris.Wrap(err, "invalid email address")
	}

	// Inject variables into the email body
	bodyWithVars, err := hs.injectVarsIntoBody(req.Body, req.Vars)
	if err != nil {
		hs.log.Log(LogEntry{
			Category: CategorySMTP,
			Level:    LevelError,
			Message:  fmt.Sprintf("Failed to inject variables: %v", err),
			Fields:   []zap.Field{zap.String("body", req.Body)},
		})
		return eris.Wrap(err, "failed to inject vars into body")
	}

	// Sanitize the HTML body
	cleaned := hs.security.SanitizeHTML(bodyWithVars)

	// Check configuration for 'From' address
	from := hs.config.SMTPFrom
	if from == "" {
		err := eris.New("no 'From' address configured in SMTP config")
		hs.log.Log(LogEntry{
			Category: CategorySMTP,
			Level:    LevelWarn, // Log as warning instead of error
			Message:  err.Error(),
		})
		// Optionally provide a fallback email
		from = "landshorizon@gmail.com"
	}

	// Set up authentication if credentials are available
	var auth smtp.Auth
	if hs.config.SMTPUsername != "" && hs.config.SMTPPassword != "" {
		auth = smtp.PlainAuth("", hs.config.SMTPUsername, hs.config.SMTPPassword, hs.config.SMTPHost)
	}

	// Construct the email message
	addr := fmt.Sprintf("%s:%d", hs.config.SMTPHost, hs.config.SMTPPort)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", from, req.To, req.Subject, cleaned)

	// Send the email using SMTP
	err = smtp.SendMail(addr, auth, from, []string{req.To}, []byte(msg))
	if err != nil {
		hs.log.Log(LogEntry{
			Category: CategorySMTP,
			Level:    LevelError,
			Message:  fmt.Sprintf("Failed to send email: %v", err),
			Fields: []zap.Field{
				zap.String("to", req.To),
				zap.String("from", from),
				zap.String("subject", req.Subject),
				zap.String("body", cleaned),
			},
		})
		return eris.Wrap(err, "smtp send failed")
	}

	// Log successful email sending
	hs.log.Log(LogEntry{
		Category: CategorySMTP,
		Level:    LevelInfo,
		Message:  "SMTP email sent successfully",
		Fields: []zap.Field{
			zap.String("to", req.To),
			zap.String("from", from),
			zap.String("subject", req.Subject),
			zap.String("body", cleaned),
		},
	})

	return nil
}

func (hs *HorizonSMTP) injectVarsIntoBody(body string, vars *map[string]string) (string, error) {
	if vars == nil || len(*vars) == 0 {
		hs.log.Log(LogEntry{
			Category: CategorySMTP,
			Level:    LevelWarn,
			Message:  "No variables provided for body templating",
		})
		return body, nil
	}

	tmpl, err := template.New("email").Parse(body)
	if err != nil {
		return "", eris.Wrap(err, "failed to parse template")
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, *vars)
	if err != nil {
		return "", eris.Wrap(err, "failed to execute template")
	}

	return buf.String(), nil
}
