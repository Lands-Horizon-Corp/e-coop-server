package horizon

type HorizonBroadcast struct{}

func NewHorizonBroadcast() (*HorizonBroadcast, error) {
	return &HorizonBroadcast{}, nil
}

func (h *HorizonBroadcast) Subscribe(topic string)   {}
func (h *HorizonBroadcast) Publish(topic string)     {}
func (h *HorizonBroadcast) Unsubscribe(topic string) {}
func (h *HorizonBroadcast) Close(topic string)       {}
