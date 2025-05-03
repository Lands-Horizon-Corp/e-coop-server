package horizon

type HorizonDatabase struct{}

func NewHorizonDatabase() (*HorizonDatabase, error) {
	return &HorizonDatabase{}, nil
}

func (hd *HorizonDatabase) Run() {}

func (hd *HorizonDatabase) Ping() {}
