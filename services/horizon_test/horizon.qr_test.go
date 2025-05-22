package horizon

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/stretchr/testify/assert"
)

// go test ./services/horizon_test/horizon.qr_test.go

func setupSecurityUtilsQR() horizon.SecurityService {
	env := horizon.NewEnvironmentService("../../.env")
	token := env.GetByteSlice("APP_TOKEN", "")
	return horizon.NewSecurityService(
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
	qrService := horizon.NewHorizonQRService(mockSecurity)

	// Sample data to encode
	inputData := map[string]interface{}{
		"user": "john_doe",
		"role": "admin",
	}

	// Encode the data
	qrResult, err := qrService.EncodeQR(ctx, inputData, "user_data")
	assert.NoError(t, err)
	assert.Equal(t, "user_data", qrResult.Type)

	// Decode the data
	decodedData, err := qrService.DecodeQR(ctx, qrResult)
	assert.NoError(t, err)

	// Convert decodedData (*any) to a map
	decodedMap, ok := (*decodedData).(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "john_doe", decodedMap["user"])
	assert.Equal(t, "admin", decodedMap["role"])

	// Optionally: Re-marshal and compare the JSON string representations
	originalJSON, _ := json.Marshal(inputData)
	decodedJSON, _ := json.Marshal(decodedMap)
	assert.JSONEq(t, string(originalJSON), string(decodedJSON))
}
