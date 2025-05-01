package horizon

type HorizonRequest struct{}

func NewHorizonRequest() (*HorizonRequest, error) {
	return &HorizonRequest{}, nil
}
