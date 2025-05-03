package horizon

type HorizonSMS struct{}

type SMSRequest struct {
	To      string             `json:"to"`
	Subject string             `json:"subject"`
	Body    string             `json:"body"`
	Vars    *map[string]string `json:"vars,omitempty"`
}

func NewHorizonSMS() (*HorizonSMS, error) {
	return &HorizonSMS{}, nil
}

func (hs *SMSRequest) Send(er *SMSRequest) {}
