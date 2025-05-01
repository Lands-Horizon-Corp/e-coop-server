package horizon

type HorizonDatabase struct{}

func NewHorizonDatabase() (*HorizonDatabase, error) {
	return &HorizonDatabase{}, nil
}
