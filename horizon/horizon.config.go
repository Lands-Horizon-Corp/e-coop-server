package horizon

type HorizonConfig struct{}

func NewHorizonConfig() (*HorizonConfig, error) {
	return &HorizonConfig{}, nil
}
