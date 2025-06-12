package horizon_test

import (
	"context"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"fmt"

	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/stretchr/testify/assert"
)

// go test -v services/horizon_test/horizon.request_test.go
var apiPort = horizon.GetFreePort()

func TestMain(m *testing.M) {
	env := horizon.NewEnvironmentService("../../.env")

	metricsPort := horizon.GetFreePort()
	clientUrl := env.GetString("APP_CLIENT_URL", "http://localhost:3000")
	clientName := env.GetString("APP_CLIENT_NAME", "test-client")
	baseURL := "http://localhost:" + fmt.Sprint(apiPort)

	// Assign package-level variables, do NOT use := to avoid shadowing
	testCtx := context.Background()

	service := horizon.NewHorizonAPIService(apiPort, metricsPort, clientUrl, clientName, false)
	go func() {
		if err := service.Run(testCtx); err != nil {
			println("Server exited with error:", err.Error())
		}
	}()

	// Wait for server to be ready
	if !waitForServerReady(baseURL+"/health", 10*time.Second) {
		panic("server did not become ready in time")
	}

	// Run all tests
	code := m.Run()

	time.Sleep(100 * time.Millisecond) // allow graceful shutdown

	os.Exit(code)
}

func waitForServerReady(url string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

func TestNewHorizonAPIService_HealthCheck(t *testing.T) {

	baseURL := "http://localhost:" + fmt.Sprint(apiPort)
	resp, err := http.Get(baseURL + "/health")
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "OK", string(body))
}

func TestNewHorizonAPIService_SuspiciousPath(t *testing.T) {

	baseURL := "http://localhost:" + fmt.Sprint(apiPort)
	resp, err := http.Get(baseURL + "/config.yaml")
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	assert.Equal(t, "Access forbidden", string(body))
}

func TestNewHorizonAPIService_WellKnownPath(t *testing.T) {

	baseURL := "http://localhost:" + fmt.Sprint(apiPort)
	resp, err := http.Get(baseURL + "/.well-known/security.txt")
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, "Path not found", string(body))
}
