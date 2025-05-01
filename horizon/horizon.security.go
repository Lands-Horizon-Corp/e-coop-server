package horizon

type HorizonSecurity struct{}

func NewHorizonSecurity() (*HorizonSecurity, error) {
	return &HorizonSecurity{}, nil
}
