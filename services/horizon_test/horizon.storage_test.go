package horizon_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/stretchr/testify/require"
)

// go test -v ./services/horizon_test/horizon.storage_test.go
func TestHorizonStorage_UploadFromBinary(t *testing.T) {
	ctx := context.Background()
	hs := createTestService(t)

	content := []byte("test file content")

	storage, err := hs.UploadFromBinary(ctx, content, nil)
	require.NoError(t, err)

	// Make sure the file name includes "test" (or skip if name is always unique)
	require.NotEmpty(t, storage.FileName)
	require.Equal(t, int64(len(content)), storage.FileSize)
	require.NotEmpty(t, storage.URL)

	// Optional: If you have a way to retrieve the uploaded content for validation
	// actualContent := hs.DownloadFile(storage.URL) // Replace with real download function
	// require.Equal(t, string(content), string(actualContent))
}
func TestHorizonStorage_UploadFromPath(t *testing.T) {
	ctx := context.Background()
	hs := createTestService(t)

	// Create temporary test file
	tmpfile, err := os.CreateTemp("../../config", "sample-*.text")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name()) // Remove the file when done

	content := []byte("test content")
	_, err = tmpfile.Write(content)
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	storage, err := hs.UploadFromPath(ctx, tmpfile.Name(), nil)
	require.NoError(t, err)

	// Get the original base filename without the extension
	originalFileNameBase := strings.TrimSuffix(filepath.Base(tmpfile.Name()), filepath.Ext(tmpfile.Name()))

	// Assert that the generated FileName contains the original base filename
	require.Contains(t, storage.FileName, originalFileNameBase)
	require.Equal(t, int64(len(content)), storage.FileSize)
	require.NotEmpty(t, storage.URL)

}

func TestHorizonStorage_GeneratePresignedURL(t *testing.T) {
	env := horizon.NewEnvironmentService("../../.env")

	testBucket := env.GetString("STORAGE_BUCKET", "my-bucket")

	ctx := context.Background()
	hs := createTestService(t)

	// First upload a test file
	data := []byte("presigned url test content")
	storage, err := hs.UploadFromBinary(ctx, data, nil)
	require.NoError(t, err)

	// Generate presigned URL
	url, err := hs.GeneratePresignedURL(ctx, storage, 5*time.Minute)
	require.NoError(t, err)
	require.Contains(t, url, testBucket)
	require.Contains(t, url, storage.StorageKey)
}

func TestHorizonStorage_DeleteFile(t *testing.T) {
	ctx := context.Background()
	hs := createTestService(t)

	// Upload test file
	data := []byte("file to delete")
	storage, err := hs.UploadFromBinary(ctx, data, nil)
	require.NoError(t, err)

	// Delete the file
	err = hs.DeleteFile(ctx, storage)
	require.NoError(t, err)

	// Try to generate URL after deletion
	_, err = hs.GeneratePresignedURL(ctx, storage, 5*time.Minute)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to presign URL")
}

func TestHorizonStorage_GenerateUniqueName(t *testing.T) {
	ctx := context.Background()
	hs := createTestService(t)

	original := "testfile.txt"
	name1, err := hs.GenerateUniqueName(ctx, original)
	require.NoError(t, err)

	name2, err := hs.GenerateUniqueName(ctx, original)
	require.NoError(t, err)

	require.NotEqual(t, name1, name2)
	require.Contains(t, name1, original)
	require.Contains(t, name2, original)
}

func createTestService(t *testing.T) *horizon.HorizonStorage {

	env := horizon.NewEnvironmentService("../../.env")

	accessKey := env.GetString("STORAGE_ACCESS_KEY", "minioadmin")
	secretKey := env.GetString("STORAGE_SECRET_KEY", "minioadmin")
	testBucket := env.GetString("STORAGE_BUCKET", "my-bucket")
	prefix := env.GetString("STORAGE_PREFIX", "test-")

	if accessKey == "" || secretKey == "" {
		t.Fatal("Missing required environment variables for B2 testing")
	}

	hs := horizon.NewHorizonStorageService(
		accessKey,
		secretKey,
		testBucket,
		prefix,
		1024*1024*10, // 10MB max size
	).(*horizon.HorizonStorage)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := hs.Run(ctx)
	require.NoError(t, err)

	return hs
}
