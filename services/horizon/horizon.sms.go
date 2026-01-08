package horizon

import (
	"bytes"
	"context"
	"html/template"
	"log"
	"os"
	"sync"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/microcosm-cc/bluemonday"
	"github.com/rotisserie/eris"
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"

	"golang.org/x/time/rate"
)

type SMSRequest struct {
	To   string            // Recipient phone number
	Body string            // Message template body
	Vars map[string]string // Dynamic variables for template interpolation
}

type SMSService interface {
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
	Format(ctx context.Context, req SMSRequest) (*SMSRequest, error)
	Send(ctx context.Context, req SMSRequest) error
}

type SMS struct {
	limiterOnce sync.Once
	limiter     *rate.Limiter
	twilio      *twilio.RestClient

	accountSID    string // Twilio Account SID
	authToken     string // Twilio Auth Token
	sender        string // Sender phone number registered with Twilio
	maxCharacters int32  // Maximum allowed length for SMS body
	secured       bool   // Whether to use secured connection
}

func NewSMS(accountSID, authToken, sender string, maxCharacters int32, secured bool) SMSService {
	return &SMS{
		accountSID:    accountSID,
		authToken:     authToken,
		sender:        sender,
		maxCharacters: maxCharacters,
		secured:       secured,
	}
}

func (h *SMS) Run(_ context.Context) error {
	h.limiterOnce.Do(func() {
		h.limiter = rate.NewLimiter(rate.Limit(1000), 100) // 10 rps, burst 5
	})
	h.twilio = twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: h.accountSID,
		Password: h.authToken,
	})
	return nil
}

func (h *SMS) Stop(_ context.Context) error {
	h.twilio = nil
	h.limiter = nil
	return nil
}

func (h *SMS) Format(_ context.Context, req SMSRequest) (*SMSRequest, error) {
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

	tmpl, err := template.New("sms").Parse(tmplBody)
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

func (h *SMS) Send(ctx context.Context, req SMSRequest) error {
	if !handlers.IsValidPhoneNumber(req.To) {
		return eris.Errorf("invalid recipient phone number format: %s", req.To)
	}
	if !handlers.IsValidPhoneNumber(h.sender) {
		return eris.Errorf("invalid sender phone number format: %s", h.sender)
	}

	formatted, err := h.Format(ctx, req)
	if err != nil {
		return eris.Wrap(err, "template formatting failed")
	}
	req.Body = formatted.Body

	if len(req.Body) > int(h.maxCharacters) {
		return eris.Errorf("SMS body exceeds %d characters (actual: %d)", h.maxCharacters, len(req.Body))
	}

	// 🔐 NOT SECURED → LOG ONLY (LOCAL / LAN MODE)
	if !h.secured {
		log.Printf(
			"[SMS MOCK MODE] To=%s | Message=%s",
			req.To,
			req.Body,
		)
		return nil
	}

	if !h.limiter.Allow() {
		return eris.New("rate limit exceeded for sending SMS")
	}

	req.Body = bluemonday.UGCPolicy().Sanitize(req.Body)

	params := &openapi.CreateMessageParams{}
	params.SetTo(req.To)
	params.SetFrom(h.sender)
	params.SetBody(req.Body)

	_, err = h.twilio.Api.CreateMessage(params)
	if err != nil {
		return eris.Wrap(err, "failed to send SMS via Twilio")
	}
	return nil
}
