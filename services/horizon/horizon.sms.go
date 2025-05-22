package horizon

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"sync"

	"github.com/microcosm-cc/bluemonday"
	"github.com/rotisserie/eris"
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	"golang.org/x/time/rate"
)

// SMSRequest represents a templated SMS message with dynamic variables
type SMSRequest struct {
	To   string            // Recipient phone number
	Body string            // Message template body
	Vars map[string]string // Dynamic variables for template interpolation
}

// SMSService defines the interface for SMS processing and delivery
type SMSService interface {
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
	Format(ctx context.Context, req SMSRequest) (*SMSRequest, error)
	Send(ctx context.Context, req SMSRequest) error
}

// HorizonSMS is the default implementation of SMSService using Twilio
type HorizonSMS struct {
	limiterOnce sync.Once
	limiter     *rate.Limiter
	twilio      *twilio.RestClient

	accountSID    string // Twilio Account SID
	authToken     string // Twilio Auth Token
	sender        string // Sender phone number registered with Twilio
	maxCharacters int32  // Maximum allowed length for SMS body
}

// NewHorizonSMS constructs a new HorizonSMS
func NewHorizonSMS(accountSID, authToken, sender string, maxCharacters int32) SMSService {
	return &HorizonSMS{
		accountSID:    accountSID,
		authToken:     authToken,
		sender:        sender,
		maxCharacters: maxCharacters,
	}
}

// Run initializes the rate limiter and Twilio client (once)
func (h *HorizonSMS) Run(ctx context.Context) error {
	h.limiterOnce.Do(func() {
		h.limiter = rate.NewLimiter(rate.Limit(10), 5) // 10 rps, burst 5
	})
	h.twilio = twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: h.accountSID,
		Password: h.authToken,
	})
	return nil
}

// Stop clears the client and limiter
func (h *HorizonSMS) Stop(ctx context.Context) error {
	h.twilio = nil
	h.limiter = nil
	return nil
}

func (h *HorizonSMS) Format(ctx context.Context, req SMSRequest) (*SMSRequest, error) {
	var tmplBody string
	if err := isValidFilePath(req.Body); err == nil {
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

// Send formats, sanitizes, rate-limits, and dispatches the SMS
func (h *HorizonSMS) Send(ctx context.Context, req SMSRequest) error {
	// Validate phone numbers
	if !IsValidPhoneNumber(req.To) {
		return fmt.Errorf("invalid recipient phone number format: %s", req.To)
	}
	if !IsValidPhoneNumber(h.sender) {
		return fmt.Errorf("invalid sender phone number format: %s", h.sender)
	}

	// Format template vars
	formatted, err := h.Format(ctx, req)
	if err != nil {
		return eris.Wrap(err, "template formatting failed")
	}
	req.Body = formatted.Body

	// Check length
	if len(req.Body) > int(h.maxCharacters) {
		return fmt.Errorf("SMS body exceeds %d characters (actual: %d)", h.maxCharacters, len(req.Body))
	}

	// Rate limiting: allow up to limiter capacity
	if !h.limiter.Allow() {
		return fmt.Errorf("rate limit exceeded for sending SMS")
	}

	// Sanitize content
	req.Body = bluemonday.UGCPolicy().Sanitize(req.Body)

	// Build Twilio params
	params := &openapi.CreateMessageParams{}
	params.SetTo(req.To)
	params.SetFrom(h.sender)
	params.SetBody(req.Body)

	// Send via Twilio
	_, err = h.twilio.Api.CreateMessage(params)
	if err != nil {
		return eris.Wrap(err, "failed to send SMS via Twilio")
	}
	return nil
}
