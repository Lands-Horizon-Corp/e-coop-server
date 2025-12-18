package horizon

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupSecurityUtilsQR() SecurityService {
	env := NewEnvironmentService("../../.env")
	token := env.GetByteSlice("APP_TOKEN", "")
	return NewSecurityService(
		env.GetUint32("PASSWORD_MEMORY", 65536),  // memory (e.g., 64MB)
		env.GetUint32("PASSWORD_ITERATIONS", 3),  // iterations
		env.GetUint8("PASSWORD_PARALLELISM", 2),  // parallelism
		env.GetUint32("PASSWORD_SALT_LENTH", 16), // salt length in bytes
		env.GetUint32("PASSWORD_KEY_LENGTH", 32), // key length in bytes
		token,
	)
}

func TestHorizonQRService_EncodeDecode(t *testing.T) {
	ctx := context.Background()
	mockSecurity := setupSecurityUtilsQR()
	qrService := NewHorizonQRService(mockSecurity)

	inputData := map[string]any{
		"user": "john_doe",
		"role": "admin",
	}

	qrResult, err := qrService.EncodeQR(ctx, inputData, "user_data")
	assert.NoError(t, err)
	assert.Equal(t, "user_data", qrResult.Type)

	decodedData, err := qrService.DecodeQR(ctx, qrResult)
	assert.NoError(t, err)

	decodedMap, ok := (*decodedData).(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "john_doe", decodedMap["user"])
	assert.Equal(t, "admin", decodedMap["role"])

	originalJSON, _ := json.Marshal(inputData)
	decodedJSON, _ := json.Marshal(decodedMap)
	assert.JSONEq(t, string(originalJSON), string(decodedJSON))
}
