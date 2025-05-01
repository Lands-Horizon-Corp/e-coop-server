package horizon

type HorizonAuthentication struct{}

func NewHorizonAuthentication() (*HorizonAuthentication, error) {
	return &HorizonAuthentication{}, nil
}
