package horizon

import (
	"fmt"
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
	var auth smtp.Auth
	if hs.config.SMTPUsername != "" && hs.config.SMTPPassword != "" {
		auth = smtp.PlainAuth("", hs.config.SMTPUsername, hs.config.SMTPPassword, hs.config.SMTPHost)
	}
	cleaned := hs.security.SanitizeHTML(req.Body)
	addr := fmt.Sprintf("%s:%d", hs.config.SMTPHost, hs.config.SMTPPort)
	msg := fmt.Appendf(nil, "To: %s\r\nSubject: %s\r\n\r\n%s", req.To, req.Subject, cleaned)
	err := smtp.SendMail(addr, auth, hs.config.SMTPFrom, []string{req.To}, msg)
	if err != nil {
		hs.log.Log(LogEntry{
			Category: CategorySMTP,
			Level:    LevelError,
			Message:  fmt.Sprintf("SMTP send failed: %v", err),
			Fields: []zap.Field{
				zap.String("to", req.To),
				zap.String("from", hs.config.SMTPFrom),
				zap.String("subject", req.Subject),
				zap.String("body", cleaned),
			},
		})
		return err
	}
	hs.log.Log(LogEntry{
		Category: CategorySMTP,
		Level:    LevelInfo,
		Message:  "SMTP email sent successfully",
		Fields: []zap.Field{
			zap.String("to", req.To),
			zap.String("from", hs.config.SMTPFrom),
			zap.String("subject", req.Subject),
			zap.String("body", req.Body),
		},
	})

	return nil
}
