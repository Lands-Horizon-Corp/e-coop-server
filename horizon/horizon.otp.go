package horizon

type HorizonOTP struct{}

func NewHorizonOTP() (*HorizonOTP, error) {
	return &HorizonOTP{}, nil
}

func SendSMSOTP()   {}
func VerifySMSOTP() {}

func SendSMTPOTP()   {}
func VerifySMTPOTP() {}
