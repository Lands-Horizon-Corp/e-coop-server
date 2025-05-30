package horizon

import (
	"context"
	"encoding/json"
	"fmt"
)

type QRResult struct {
	Data string `json:"data"`
	Type string `json:"type"`
}

type QRService interface {
	DecodeQR(ctx context.Context, data *QRResult) (*any, error)
	EncodeQR(ctx context.Context, data any, qrType string) (*QRResult, error)
}

type HorizonQRService struct {
	security SecurityService
}

func NewHorizonQRService(
	security SecurityService,
) QRService {
	return &HorizonQRService{
		security: security,
	}
}

func (h *HorizonQRService) DecodeQR(ctx context.Context, data *QRResult) (*any, error) {
	decrypted, err := h.security.Decrypt(ctx, data.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}
	var decoded any
	if err := json.Unmarshal([]byte(decrypted), &decoded); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return &decoded, nil
}

func (h *HorizonQRService) EncodeQR(ctx context.Context, data any, qrTYpe string) (*QRResult, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	encrypted, err := h.security.Encrypt(ctx, string(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}
	return &QRResult{
		Data: encrypted,
		Type: qrTYpe,
	}, nil
}
