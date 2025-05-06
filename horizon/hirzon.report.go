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

// Generate will generate a CSV from the provided data (interface{}) and upload it to storage
func (hr *HorizonReport) Generate(data []any, onProgress ProgressCallback, name string) (*Storage, error) {

	return nil, nil
}
