package horizon

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
	"strings"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/rotisserie/eris"
)

type SMTPRequest struct {
	To      string
	Subject string
	Body    string
	Vars    map[string]string
}

type SMTPImpl struct {
	host     string
	port     int
	username string
	password string
	from     string

	secured bool
}

func NewSMTPImpl(host string, port int, username, password string, from string, secured bool) *SMTPImpl {
	return &SMTPImpl{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		secured:  secured,
	}
}

func (h *SMTPImpl) Format(req SMTPRequest) (*SMTPRequest, error) {
	var tmplBody string
	if err := helpers.IsValidFilePath(req.Body); err == nil {
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

func (h *SMTPImpl) Send(context context.Context, req SMTPRequest) error {
	if !helpers.IsValidEmail(req.To) {
		return eris.New("Recipient email format is invalid")
	}
	if !helpers.IsValidEmail(h.from) {
		return eris.New("Admin email format is invalid")
	}
	req.Body = helpers.Sanitize(req.Body)
	finalBody, err := h.Format(req)
	if err != nil {
		return eris.Wrap(err, "failed to inject variables into body")
	}
	req.Body = finalBody.Body
	if !h.secured {
		log.Printf(
			"[SMTP MOCK MODE] Sending email\nTo: %s\nSubject: %s\nBody:\n%s\n",
			req.To,
			req.Subject,
			req.Body,
		)
	}
	sintizer := func(s string) string {
		return strings.ReplaceAll(strings.ReplaceAll(s, "\r", ""), "\n", "")
	}
	safeFrom := sintizer(h.from)
	safeTo := sintizer(req.To)
	safeSubject := sintizer(req.Subject)
	addr := fmt.Sprintf("%s:%d", h.host, h.port)
	var auth smtp.Auth
	if h.secured && h.username != "" && h.password != "" {
		auth = smtp.PlainAuth("", h.username, h.password, h.host)
	} else {
		auth = nil
	}

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s",
		safeFrom,
		safeTo,
		safeSubject,
		req.Body,
	)
	if err := smtp.SendMail(
		addr,
		auth,
		safeFrom,
		[]string{safeTo},
		[]byte(msg),
	); err != nil {
		return eris.Wrap(err, "smtp send failed")
	}
	return nil
}
