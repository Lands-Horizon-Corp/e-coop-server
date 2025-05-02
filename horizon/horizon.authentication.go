package horizon

type HorizonAuthentication struct{}

func NewHorizonAuthentication() (*HorizonAuthentication, error) {
	return &HorizonAuthentication{}, nil
}

func (ha *HorizonAuthentication) GenerateToken() {}

func (ha *HorizonAuthentication) VerifyToken() {}

func (ha *HorizonAuthentication) DeleteToken() {}
