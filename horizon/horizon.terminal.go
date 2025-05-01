package horizon

type HorizonTerminal struct{}

func NewHorizonTerminal() (*HorizonTerminal, error) {
	// expose server
	// expose server swagger
	// ping server

	// expose mailhog ui
	// expose mailing smtp server
	// health check mailing

	// expose redis ui
	// expose redis server
	// ping redis server

	// expose minio ui
	// expose minio server
	// ping minio

	// expose sms server
	// expose sms ui
	// health check sms

	// expose database ui
	// expose database server
	// ping database
	return &HorizonTerminal{}, nil
}
