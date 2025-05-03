package horizon

type HorizonOTP struct {
	config *HorizonConfig
}

func NewHorizonOTP(config *HorizonConfig) (*HorizonOTP, error) {
	return &HorizonOTP{
		config: config,
	}, nil
}

func (ho *HorizonOTP) GenerateOTP(key string, value string) {
	// token := ho.config.AppToken
	// send the generated JWT string to redis
}
func (ho *HorizonOTP) VerifyOTP(key string, value string) {
	// token := ho.config.AppToken
	// send the generate JWT string to redis
}
