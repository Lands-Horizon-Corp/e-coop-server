package horizon

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStorageImpl_UploadFromBinary(t *testing.T) {
	ctx := context.Background()
	hs := createTestService(t)

	content := []byte("test file content")

	storage, err := hs.UploadFromBinary(ctx, content, nil)
	require.NoError(t, err)

	require.NotEmpty(t, storage.FileName)
	require.Equal(t, int64(len(content)), storage.FileSize)
	require.NotEmpty(t, storage.URL)

}
func TestStorageImpl_UploadFromPath(t *testing.T) {
	ctx := context.Background()
	hs := createTestService(t)

	tmpfile, err := os.CreateTemp("../../config", "sample-*.text")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpfile.Name()) }() // Remove the file when done

	content := []byte("test content")
	_, err = tmpfile.Write(content)
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	storage, err := hs.UploadFromPath(ctx, tmpfile.Name(), nil)
	require.NoError(t, err)

	originalFileNameBase := strings.TrimSuffix(filepath.Base(tmpfile.Name()), filepath.Ext(tmpfile.Name()))

	require.Contains(t, storage.FileName, originalFileNameBase)
	require.Equal(t, int64(len(content)), storage.FileSize)
	require.NotEmpty(t, storage.URL)

}

func TestStorageImpl_GeneratePresignedURL(t *testing.T) {
	env := NewEnvironmentService("../../.env")

	testBucket := env.GetString("STORAGE_BUCKET", "cooperatives-development")

	ctx := context.Background()
	hs := createTestService(t)

	data := []byte("presigned url test content")
	storage, err := hs.UploadFromBinary(ctx, data, nil)
	require.NoError(t, err)

	url, err := hs.GeneratePresignedURL(ctx, storage, 5*time.Minute)
	require.NoError(t, err)
	require.Contains(t, url, testBucket)
	require.Contains(t, url, storage.StorageKey)
}

func TestStorageImpl_DeleteFile(t *testing.T) {
	ctx := context.Background()
	hs := createTestService(t)

	data := []byte("file to delete")
	storage, err := hs.UploadFromBinary(ctx, data, nil)
	require.NoError(t, err)

	err = hs.DeleteFile(ctx, storage)
	require.NoError(t, err)

	_, err = hs.GeneratePresignedURL(ctx, storage, 5*time.Minute)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate presigned URL")
}

func TestStorageImpl_GenerateUniqueName(t *testing.T) {
	ctx := context.Background()
	hs := createTestService(t)

	original := "testfile.txt"
	contentType := "text/plain"

	name1, err := hs.GenerateUniqueName(ctx, original, contentType)
	require.NoError(t, err)

	name2, err := hs.GenerateUniqueName(ctx, original, contentType)
	require.NoError(t, err)

	require.NotEqual(t, name1, name2)
	require.Contains(t, name1, "testfile")
	require.Contains(t, name2, "testfile")
	require.True(t, strings.HasSuffix(name1, ".txt"))
	require.True(t, strings.HasSuffix(name2, ".txt"))
}

func TestStorageImpl_GenerateUniqueNameWithoutExtension(t *testing.T) {
	ctx := context.Background()
	hs := createTestService(t)

	original := "testfile"
	contentType := "image/jpeg"

	name, err := hs.GenerateUniqueName(ctx, original, contentType)
	require.NoError(t, err)

	require.Contains(t, name, "testfile")
	require.True(t, strings.HasSuffix(name, ".jpg"))
}

func TestStorageImpl_GenerateUniqueNameEmptyContentType(t *testing.T) {
	ctx := context.Background()
	hs := createTestService(t)

	original := "testfile"
	contentType := ""

	name, err := hs.GenerateUniqueName(ctx, original, contentType)
	require.NoError(t, err)

	require.Contains(t, name, "testfile")
	require.False(t, strings.Contains(name, "."))
}

func createTestService(t *testing.T) *StorageImpl {

	env := NewEnvironmentService("../../.env")

	accessKey := env.GetString("STORAGE_ACCESS_KEY", "minioadmin")
	secretKey := env.GetString("STORAGE_SECRET_KEY", "minioadmin")
	testBucket := env.GetString("STORAGE_BUCKET", "cooperatives-development")
	endpoint := env.GetString("STORAGE_URL", "")
	region := env.GetString("STORAGE_REGION", "")
	driver := env.GetString("STORAGE_DRIVER", "")
	isStaging := env.GetString("APP_ENV", "development") == "staging"

	if accessKey == "" || secretKey == "" {
		t.Fatal("Missing required environment variables for B2 testing")
	}

	hs := NewStorageImplService(
		accessKey,
		secretKey,
		endpoint,
		testBucket,
		region,
		driver,
		1024*1024*10,
		isStaging,
	).(*StorageImpl)

	ctx := context.Background()
	err := hs.Run(ctx)
	require.NoError(t, err)

	return hs
}
