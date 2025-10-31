package horizon

import (
	"context"
	"encoding/json"

	"github.com/rotisserie/eris"
)

// QRResult represents the structure of QR code data.
type QRResult struct {
	Data string `json:"data"`
	Type string `json:"type"`
}

// QRService defines the interface for QR code operations.
type QRService interface {
	DecodeQR(ctx context.Context, data *QRResult) (*any, error)
	EncodeQR(ctx context.Context, data any, qrType string) (*QRResult, error)
}

// HorizonQRService provides an implementation of QRService.
type HorizonQRService struct {
	security SecurityService
}

// NewHorizonQRService creates a new QR service instance.
func NewHorizonQRService(
	security SecurityService,
) QRService {
	return &HorizonQRService{
		security: security,
	}
}

// DecodeQR decodes and decrypts QR code data.
func (h *HorizonQRService) DecodeQR(ctx context.Context, data *QRResult) (*any, error) {
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

// EncodeQR encrypts and encodes data for QR code generation.
func (h *HorizonQRService) EncodeQR(ctx context.Context, data any, qrTYpe string) (*QRResult, error) {
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
