package horizon

type HorizonCache struct{}

func NewHorizonCache() (*HorizonCache, error) {
	return &HorizonCache{}, nil
}
func (hc *HorizonCache) Delete() {}
func (hc *HorizonCache) Exist()  {}
func (hc *HorizonCache) Set()    {}
func (hc *HorizonCache) Get()    {}
func (hc *HorizonCache) Ping()   {}
