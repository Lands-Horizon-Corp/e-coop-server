package horizon

type HorizonOTP struct{}

func NewHorizonOTP() (*HorizonOTP, error) {
	return &HorizonOTP{}, nil
}
