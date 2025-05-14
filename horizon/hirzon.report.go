package horizon

// HorizonReport struct holds the storage reference
type HorizonReport struct {
	storage *HorizonStorage
}

// NewHorizonReport creates a new instance of HorizonReport
func NewHorizonReport(storage *HorizonStorage) (*HorizonReport, error) {
	return &HorizonReport{
		storage: storage,
	}, nil
}

func (hr *HorizonReport) Generate(data []any, onProgress ProgressCallback, name string) (*Storage, error) {
	return nil, nil
}
