package horizon

type HorizonSchedule struct{}

func NewHorizonSchedule() (*HorizonSchedule, error) {
	return &HorizonSchedule{}, nil
}
