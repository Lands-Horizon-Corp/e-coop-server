package horizon

type HorizonCache struct{}

func NewHorizonCache() (*HorizonCache, error) {
	return &HorizonCache{}, nil
}
