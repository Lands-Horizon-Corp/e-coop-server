/*
	err = sms.Send(&SMSRequest{
		To:      "+639194893088",
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
	"regexp"
	"sync"

	"github.com/rotisserie/eris"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type HorizonSMS struct {
	log         *HorizonLog
	security    *HorizonSecurity
	limiter     *rate.Limiter
	limiterOnce sync.Once
	ctx         context.Context
	cancel      context.CancelFunc
}

type SMSRequest struct {
	To      string             `json:"to"`
	Subject string             `json:"subject"`
	Body    string             `json:"body"`
	Vars    *map[string]string `json:"vars,omitempty"`
}

func NewHorizonSMS(
	log *HorizonLog,
	security *HorizonSecurity,
) (*HorizonSMS, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &HorizonSMS{
		log:      log,
		security: security,

		limiter: rate.NewLimiter(1, 3),
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

func (hs *HorizonSMS) run() {
	hs.limiterOnce.Do(func() {
		hs.limiter = rate.NewLimiter(1, 3)
	})
}

func (hs *HorizonSMS) stop() {
	hs.log.Log(LogEntry{
		Category: CategorySMTP,
		Level:    LevelInfo,
		Message:  "Stopping HorizonSMS service gracefully...",
	})
	hs.cancel()
}

func (hs *HorizonSMS) Send(req *SMSRequest) error {
	if !hs.isValidPhoneNumber(req.To) {
		err := fmt.Errorf("invalid phone number format: %s", req.To)
		hs.log.Log(LogEntry{
			Category: CategorySMS,
			Level:    LevelError,
			Message:  err.Error(),
			Fields: []zap.Field{
				zap.String("to", req.To),
				zap.String("subject", req.Subject),
				zap.String("body", req.Body),
			},
		})
		return err

	}

	if len(req.Body) > maxSMSCharacters {
		err := fmt.Errorf("SMS body exceeds %d characters (actual: %d)", maxSMSCharacters, len(req.Body))
		hs.log.Log(LogEntry{
			Category: CategorySMS,
			Level:    LevelError,
			Message:  err.Error(),
			Fields: []zap.Field{
				zap.String("to", req.To),
				zap.String("subject", req.Subject),
				zap.String("body", req.Body),
			},
		})
		return err
	}

	if !hs.limiter.Allow() {
		err := fmt.Errorf("rate limit exceeded for sending SMS")
		hs.log.Log(LogEntry{
			Category: CategorySMS,
			Level:    LevelError,
			Message:  err.Error(),
			Fields: []zap.Field{
				zap.String("to", req.To),
				zap.String("subject", req.Subject),
				zap.String("body", req.Body),
			},
		})
		return err
	}
	bodyWithVars, err := hs.injectVarsIntoBody(req.Body, req.Vars)
	if err != nil {
		hs.log.Log(LogEntry{
			Category: CategorySMS,
			Level:    LevelError,
			Message:  eris.ToString(err, true),
			Fields:   []zap.Field{zap.String("body", req.Body)},
		})
		return eris.Wrap(err, "failed to inject vars into body")
	}
	cleaned := hs.security.SanitizeHTML(bodyWithVars)
	hs.log.Log(LogEntry{
		Category: CategorySMS,
		Level:    LevelInfo,
		Message:  "SMS sent successfully",
		Fields: []zap.Field{
			zap.String("to", req.To),
			zap.String("subject", req.Subject),
			zap.String("body", cleaned),
		},
	})
	return nil

}
func (hs *HorizonSMS) injectVarsIntoBody(body string, vars *map[string]string) (string, error) {
	if vars == nil {
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
func (hs *HorizonSMS) isValidPhoneNumber(phoneNumber string) bool {
	re := regexp.MustCompile(`^\+(\d{1,4})\d{7,14}$`)
	return re.MatchString(phoneNumber)
}
