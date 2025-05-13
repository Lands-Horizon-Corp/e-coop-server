package horizon

import (
	"encoding/json"

	"github.com/rotisserie/eris"
	"go.uber.org/zap"
)

type QRResult struct {
	QRCode string `json:"qr_code"`
	Type   string `json:"type"`
}

type HorizonQR struct {
	config   *HorizonConfig
	log      *HorizonLog
	security *HorizonSecurity
}

func NewHorizonQR(
	config *HorizonConfig,
	log *HorizonLog,
	security *HorizonSecurity,
) (*HorizonQR, error) {
	return &HorizonQR{
		config:   config,
		log:      log,
		security: security,
	}, nil
}

func (hq *HorizonQR) Encode(data any, typeName string) (*QRResult, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		hq.log.Log(LogEntry{
			Category: CategoryQR,
			Level:    LevelError,
			Message:  "failed to marshal original data to JSON",
			Fields: []zap.Field{
				zap.String("type", typeName),
				zap.Any("data", data),
				zap.Error(err),
			},
		})
		return nil, eris.Wrap(err, "failed to marshal data")
	}
	var dataMap map[string]any
	if err := json.Unmarshal(dataBytes, &dataMap); err != nil {
		hq.log.Log(LogEntry{
			Category: CategoryQR,
			Level:    LevelError,
			Message:  "failed to unmarshal data into map[string]any",
			Fields: []zap.Field{
				zap.String("type", typeName),
				zap.ByteString("json", dataBytes),
				zap.Error(err),
			},
		})
		return nil, eris.Wrap(err, "failed to convert data to map")
	}
	wrapped := map[string]any{
		"type": typeName,
		"data": dataMap,
	}
	finalBytes, err := json.Marshal(wrapped)
	if err != nil {
		hq.log.Log(LogEntry{
			Category: CategoryQR,
			Level:    LevelError,
			Message:  "failed to marshal wrapped data",
			Fields: []zap.Field{
				zap.String("type", typeName),
				zap.Any("wrapped", wrapped),
				zap.Error(err),
			},
		})
		return nil, eris.Wrap(err, "failed to marshal wrapped data")
	}
	encryptedStr, err := hq.security.Encrypt(string(finalBytes))
	if err != nil {
		hq.log.Log(LogEntry{
			Category: CategoryQR,
			Level:    LevelError,
			Message:  "failed to encrypt marshaled data",
			Fields: []zap.Field{
				zap.String("type", typeName),
				zap.ByteString("json", finalBytes),
				zap.Error(err),
			},
		})
		return nil, eris.Wrap(err, "encryption failed")
	}
	decoded, err := hq.Decode(encryptedStr)
	if err != nil {
		hq.log.Log(LogEntry{
			Category: CategoryQR,
			Level:    LevelError,
			Message:  "failed to decode encrypted data",
			Fields: []zap.Field{
				zap.String("type", typeName),
				zap.String("encrypted_str", encryptedStr),
				zap.Error(err),
			},
		})
		return nil, eris.Wrap(err, "failed to decode encrypted data")
	}
	hq.log.Log(LogEntry{
		Category: CategoryQR,
		Level:    LevelInfo,
		Message:  "Successfully converted data to QR",
		Fields: []zap.Field{
			zap.String("qr_code", encryptedStr),
			zap.String("type", typeName),
			zap.Any("decoded_data", decoded),
		},
	})

	return &QRResult{
		QRCode: encryptedStr,
		Type:   typeName,
	}, nil
}

func (hq *HorizonQR) Decode(qrCodeBase64 string) (any, error) {
	decryptedStr, err := hq.security.Decrypt(qrCodeBase64)
	if err != nil {
		hq.log.Log(LogEntry{
			Category: CategoryQR,
			Level:    LevelError,
			Message:  "Decryption failed",
			Fields: []zap.Field{
				zap.String("qr_code_base64", qrCodeBase64),
				zap.Error(err),
			},
		})

		return nil, eris.Wrap(err, "decryption failed")
	}

	var wrapper struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal([]byte(decryptedStr), &wrapper); err != nil {
		hq.log.Log(LogEntry{
			Category: CategoryQR,
			Level:    LevelError,
			Message:  "Failed to parse wrapped data",
			Fields: []zap.Field{
				zap.String("decrypted_str", decryptedStr),
				zap.Error(err),
			},
		})
		return nil, eris.Wrap(err, "failed to parse wrapped data")
	}
	return wrapper.Data, err
}
