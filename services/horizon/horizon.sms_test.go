package horizon

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func injectMockTwilio(h *SMS) {
	err := h.Run(context.Background())
	if err != nil {
		panic(err)
	}
}

func TestSendSMS(t *testing.T) {
	env := NewEnvironmentService("../../.env")

	accountSID := env.GetString("TWILIO_ACCOUNT_SID", "")
	authToken := env.GetString("TWILIO_AUTH_TOKEN", "")
	sender := env.GetString("TWILIO_SENDER", "")
	receiver := env.GetString("TWILIO_TEST_RECIEVER", "")
	if accountSID == "" || authToken == "" {
		return
	}

	h := NewSMS(accountSID, authToken, sender, 160).(*SMS)
	injectMockTwilio(h)

	tests := []struct {
		name    string
		req     SMSRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: SMSRequest{
				To:   receiver,
				Body: "Hi {{.Name}}, alert {{.Code}}: {{.Message}}",
				Vars: map[string]string{
					"Name":    "Alice",
					"Code":    "A001",
					"Message": "Test alert",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid recipient number",
			req: SMSRequest{
				To:   "invalid-number",
				Body: "Test",
				Vars: map[string]string{},
			},
			wantErr: true,
		},
		{
			name: "body too long",
			req: SMSRequest{
				To:   receiver,
				Body: string(make([]byte, 200)), // Exceeds 160 chars
				Vars: map[string]string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := h.Send(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err, "expected an error")
			} else {
				assert.NoError(t, err, "did not expect an error")
			}
		})
	}
}
