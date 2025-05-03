package horizon

type HorizonReport struct{}

func NewHorizonReport() (*HorizonReport, error) {
	return &HorizonReport{}, nil
}
