package horizon

type HorizonSMTP struct{}

func NewHorizonSMTP() (*HorizonSMTP, error) {
	return &HorizonSMTP{}, nil
}
