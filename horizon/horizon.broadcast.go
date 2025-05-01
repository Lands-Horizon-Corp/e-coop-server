package horizon

type HorizonBroadcast struct{}

func NewHorizonBroadcast() (*HorizonBroadcast, error) {
	return &HorizonBroadcast{}, nil
}
