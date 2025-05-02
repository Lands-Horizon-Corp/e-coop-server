package horizon

type HorizonStorage struct{}

func NewHorizonStorage() (*HorizonStorage, error) {
	return &HorizonStorage{}, nil
}

func (hs *HorizonStorage) CreateBucketIfNotExists() {}
func (hs *HorizonStorage) UploadFile()              {}
func (hs *HorizonStorage) UploadLocalFile()         {}
func (hs *HorizonStorage) UploadFromURL()           {}
func (hs *HorizonStorage) DeleteFile()              {}
func (hs *HorizonStorage) GeneratePresignedURL()    {}
