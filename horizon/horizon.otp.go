package horizon

type HorizonOtp struct{}

func NewHorizonOtp() (*HorizonOtp, error) {
	return &HorizonOtp{}, nil
}
