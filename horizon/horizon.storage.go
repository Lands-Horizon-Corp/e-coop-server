package horizon

type HorizonStorage struct{}

func NewHorizonStorage() (*HorizonStorage, error) {
	return &HorizonStorage{}, nil
}
