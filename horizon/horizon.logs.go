package horizon

type HorizonLog struct{}

func NewHorizonLog() (*HorizonLog, error) {
	return &HorizonLog{}, nil
}
