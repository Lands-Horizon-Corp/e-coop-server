package horizon

type HorizonSMS struct{}

func NewHorizonSMS() (*HorizonSMS, error) {
	return &HorizonSMS{}, nil
}
