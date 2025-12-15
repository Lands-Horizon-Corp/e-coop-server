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

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/rotisserie/eris"
	"golang.org/x/time/rate"
)

func sanitizeHeader(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r", ""), "\n", "")
}

type SMTPRequest struct {
	To      string            // Recipient SMTP address
	Subject string            // SMTP subject line
	Body    string            // Template body with placeholders
	Vars    map[string]string // Dynamic variables for template interpolation
}

type SMTPService interface {
	Run(ctx context.Context) error

	Stop(ctx context.Context) error

	Format(ctx context.Context, req SMTPRequest) (*SMTPRequest, error)

	Send(ctx context.Context, req SMTPRequest) error
}

type SMTP struct {
	host     string
	port     int
	username string
	password string
	from     string

	limiterOnce sync.Once
	limiter     *rate.Limiter
}

func NewSMTP(host string, port int, username, password string, from string) SMTPService {
	return &SMTP{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (h *SMTP) Run(_ context.Context) error {
	h.limiterOnce.Do(func() {
		h.limiter = rate.NewLimiter(rate.Limit(1000), 100) // 10 rps, burst 5
	})
	return nil
}

func (h *SMTP) Stop(_ context.Context) error {
	h.limiter = nil
	return nil
}

func (h *SMTP) Format(_ context.Context, req SMTPRequest) (*SMTPRequest, error) {
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

func (h *SMTP) Send(ctx context.Context, req SMTPRequest) error {
	if !handlers.IsValidEmail(req.To) {
		return eris.New("Recipient email format is invalid")
	}
	if !handlers.IsValidEmail(h.from) {
		return eris.New("Admin email format is invalid")
	}

	if err := h.limiter.Wait(ctx); err != nil {
		return eris.Wrap(err, "rate limit wait failed")
	}

	req.Body = handlers.Sanitize(req.Body)
	finalBody, err := h.Format(ctx, req)
	if err != nil {
		return eris.Wrap(err, "failed to inject variables into body")
	}
	req.Body = finalBody.Body

	safeFrom := sanitizeHeader(h.from)
	safeTo := sanitizeHeader(req.To)
	safeSubject := sanitizeHeader(req.Subject)

	auth := smtp.PlainAuth("", h.username, h.password, h.host)
	addr := fmt.Sprintf("%s:%d", h.host, h.port)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", safeFrom, safeTo, safeSubject, req.Body)

	if err := smtp.SendMail(addr, auth, safeFrom, []string{safeTo}, []byte(
		handlers.Sanitize(msg),
	)); err != nil {
		return eris.Wrap(err, "smtp send failed")
	}
	return nil
}
