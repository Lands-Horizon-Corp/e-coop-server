package horizon

type SMTPRequest struct {
	To      string             `json:"to"`
	Subject string             `json:"subject"`
	Body    string             `json:"body"`
	Vars    *map[string]string `json:"vars,omitempty"`
}

type HorizonSMTP struct{}

func NewHorizonSMTP() (*HorizonSMTP, error) {
	return &HorizonSMTP{}, nil
}

func (hs *HorizonSMTP) Send(er *SMTPRequest) {}
