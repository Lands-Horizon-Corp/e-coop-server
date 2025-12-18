package horizon

import (
	"context"
	"encoding/json"

	"github.com/rotisserie/eris"
)

type QRResult struct {
	Data string `json:"data"`
	Type string `json:"type"`
}

type QRService interface {
	DecodeQR(ctx context.Context, data *QRResult) (*any, error)
	EncodeQR(ctx context.Context, data any, qrType string) (*QRResult, error)
}

type QRServiceImpl struct {
	security SecurityService
}

func NewHorizonQRService(
	security SecurityService,
) QRService {
	return &QRServiceImpl{
		security: security,
	}
}

func (h *QRServiceImpl) DecodeQR(ctx context.Context, data *QRResult) (*any, error) {
	decrypted, err := h.security.Decrypt(ctx, data.Data)
	if err != nil {
		return nil, eris.Wrap(err, "failed to decrypt data")
	}
	var decoded any
	if err := json.Unmarshal([]byte(decrypted), &decoded); err != nil {
		return nil, eris.Wrap(err, "failed to unmarshal JSON")
	}
	return &decoded, nil
}

func (h *QRServiceImpl) EncodeQR(ctx context.Context, data any, qrTYpe string) (*QRResult, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, eris.Wrap(err, "failed to marshal data")
	}
	encrypted, err := h.security.Encrypt(ctx, string(jsonBytes))
	if err != nil {
		return nil, eris.Wrap(err, "failed to encrypt data")
	}
	return &QRResult{
		Data: encrypted,
		Type: qrTYpe,
	}, nil
}
