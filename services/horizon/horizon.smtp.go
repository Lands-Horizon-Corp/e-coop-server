package horizon

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"strings"
	"sync"

	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/microcosm-cc/bluemonday"
	"github.com/rotisserie/eris"
	"golang.org/x/time/rate"
)

// sanitizeHeader removes CR and LF to prevent header injection
func sanitizeHeader(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r", ""), "\n", "")
}

// SMTPRequest represents a templated SMTP request with dynamic variables as map[string]string
type SMTPRequest struct {
	To      string            // Recipient SMTP address
	Subject string            // SMTP subject line
	Body    string            // Template body with placeholders
	Vars    map[string]string // Dynamic variables for template interpolation
}

type SMTPService interface {
	// Run initializes internal resources like rate limiter
	Run(ctx context.Context) error

	// Stop cleans up resources
	Stop(ctx context.Context) error

	// Format processes template and injects variables
	Format(ctx context.Context, req SMTPRequest) (*SMTPRequest, error)

	// Send dispatches the formatted SMTP to the recipient
	Send(ctx context.Context, req SMTPRequest) error
}

type HorizonSMTP struct {
	host     string
	port     int
	username string
	password string
	from     string

	limiterOnce sync.Once
	limiter     *rate.Limiter
}

// NewHorizonSMTP constructs a new HorizonSMTP client
func NewHorizonSMTP(host string, port int, username, password string, from string) SMTPService {
	return &HorizonSMTP{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

// Run implements SMTPService.
func (h *HorizonSMTP) Run(ctx context.Context) error {
	h.limiterOnce.Do(func() {
		h.limiter = rate.NewLimiter(rate.Limit(1000), 100) // 10 rps, burst 5
	})
	return nil
}

// Stop implements SMTPService.
func (h *HorizonSMTP) Stop(ctx context.Context) error {
	h.limiter = nil
	return nil
}

// Format implements SMTPService.
func (h *HorizonSMTP) Format(ctx context.Context, req SMTPRequest) (*SMTPRequest, error) {
	var tmplBody string

	if err := handlers.IsValidFilePath(req.Body); err == nil {
		content, err := os.ReadFile(req.Body)
		if err != nil {
			return nil, eris.Wrap(err, "failed to read template file")
		}
		tmplBody = string(content)
	} else {
		tmplBody = req.Body
	}

	tmpl, err := template.New("email").Parse(tmplBody)
	if err != nil {
		return nil, eris.Wrap(err, "parse template failed")
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, req.Vars); err != nil {
		return nil, eris.Wrap(err, "execute template failed")
	}
	req.Body = buf.String()
	return &req, nil
}

// Send implements SMTPService.
func (h *HorizonSMTP) Send(ctx context.Context, req SMTPRequest) error {
	if !handlers.IsValidEmail(req.To) {
		return eris.New("Recipient email format is invalid")
	}
	if !handlers.IsValidEmail(h.from) {
		return eris.New("Admin email format is invalid")
	}

	// Wait for rate limiter token (blocking)
	if err := h.limiter.Wait(ctx); err != nil {
		return eris.Wrap(err, "rate limit wait failed")
	}

	// Sanitize and format body using bluemonday
	req.Body = bluemonday.UGCPolicy().Sanitize(req.Body)
	finalBody, err := h.Format(ctx, req)
	if err != nil {
		return eris.Wrap(err, "failed to inject variables into body")
	}
	req.Body = finalBody.Body

	// Sanitize headers to prevent injection
	safeFrom := sanitizeHeader(h.from)
	safeTo := sanitizeHeader(req.To)
	safeSubject := sanitizeHeader(req.Subject)

	auth := smtp.PlainAuth("", h.username, h.password, h.host)
	addr := fmt.Sprintf("%s:%d", h.host, h.port)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", safeFrom, safeTo, safeSubject, req.Body)

	if err := smtp.SendMail(addr, auth, safeFrom, []string{safeTo}, []byte(
		bluemonday.UGCPolicy().Sanitize(msg),
	)); err != nil {
		return eris.Wrap(err, "smtp send failed")
	}
	return nil
}
