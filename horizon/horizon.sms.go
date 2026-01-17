package horizon

import (
	"bytes"
	"context"
	"html/template"
	"log"
	"os"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/microcosm-cc/bluemonday"
	"github.com/rotisserie/eris"
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type SMSRequest struct {
	To   string
	Body string
	Vars map[string]string
}

type SMSImpl struct {
	twilio *twilio.RestClient

	accountSID    string // Twilio Account SID
	authToken     string // Twilio Auth Token
	sender        string // Sender phone number registered with Twilio
	maxCharacters int32  // Maximum allowed length for SMS body
	secured       bool   // Whether to use secured connection
}

func NewSMSImpl(accountSID, authToken, sender string, maxCharacters int32, secured bool) *SMSImpl {
	return &SMSImpl{
		accountSID:    accountSID,
		authToken:     authToken,
		sender:        sender,
		maxCharacters: maxCharacters,
		secured:       secured,
	}
}

func (h *SMSImpl) Format(req SMSRequest) (*SMSRequest, error) {
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

func (h *SMSImpl) Send(ctx context.Context, req SMSRequest) error {
	if !helpers.IsValidPhoneNumber(req.To) {
		return eris.Errorf("invalid recipient phone number format: %s", req.To)
	}
	if !helpers.IsValidPhoneNumber(h.sender) {
		return eris.Errorf("invalid sender phone number format: %s", h.sender)
	}
	formatted, err := h.Format(req)
	if err != nil {
		return eris.Wrap(err, "template formatting failed")
	}
	req.Body = formatted.Body
	if len(req.Body) > int(h.maxCharacters) {
		return eris.Errorf("SMS body exceeds %d characters (actual: %d)", h.maxCharacters, len(req.Body))
	}
	if !h.secured {
		log.Printf("[SMS MOCK MODE] To=%s | Message=%s", req.To, req.Body)
		return nil
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
