package horizon

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Backblaze/blazer/b2"
	"github.com/rotisserie/eris"
)

type StorageService interface {
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
	Upload(ctx context.Context, file any, opts ProgressCallback) (*Storage, error)
	UploadFromBinary(ctx context.Context, data []byte, opts ProgressCallback) (*Storage, error)
	UploadFromHeader(ctx context.Context, hdr *multipart.FileHeader, opts ProgressCallback) (*Storage, error)
	UploadFromPath(ctx context.Context, path string, opts ProgressCallback) (*Storage, error)
	GeneratePresignedURL(ctx context.Context, storage *Storage, expiry time.Duration) (string, error)
	DeleteFile(ctx context.Context, storage *Storage) error
	GenerateUniqueName(ctx context.Context, originalName string) (string, error)
}

type Storage struct {
	FileName   string
	FileSize   int64
	FileType   string
	StorageKey string
	URL        string
	BucketName string
	Status     string
	Progress   int64
}

type ProgressCallback func(progress int64, total int64, storage *Storage)

type HorizonStorage struct {
	storageAccessKey string
	storageSecretKey string
	storageBucket    string
	prefix           string
	maxFileSize      int64
	b2Client         *b2.Client
	bucket           *b2.Bucket
}

func NewHorizonStorageService(
	accessKey,
	secretKey,
	prefix,
	bucket string,
	maxSize int64,
) StorageService {
	return &HorizonStorage{
		storageAccessKey: accessKey,
		storageSecretKey: secretKey,
		prefix:           prefix,
		storageBucket:    bucket,
		maxFileSize:      maxSize,
	}
}

func (h *HorizonStorage) Run(ctx context.Context) error {
	client, err := b2.NewClient(
		ctx,
		h.storageAccessKey,
		h.storageSecretKey,
	)
	if err != nil {
		return eris.Wrap(err, "B2 authentication failed")
	}
	h.b2Client = client

	// Try to get existing bucket first
	if bucket, err := client.Bucket(ctx, h.storageBucket); err == nil {
		h.bucket = bucket
		return nil
	}

	// Create new bucket if it doesn't exist
	bucket, err := client.NewBucket(ctx, h.storageBucket, &b2.BucketAttrs{
		Type: b2.Private,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to create bucket %s", h.storageBucket)
	}
	h.bucket = bucket
	return nil
}

func (h *HorizonStorage) Stop(ctx context.Context) error {
	h.b2Client = nil
	h.bucket = nil
	return nil
}

type progressReader struct {
	reader    io.Reader
	callback  ProgressCallback
	total     int64
	readSoFar int64
	storage   *Storage
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.readSoFar += int64(n)
		percent := pr.readSoFar * 100 / pr.total
		if percent > 100 {
			percent = 100
		}
		pr.storage.Progress = percent
		if pr.callback != nil {
			pr.callback(percent, 100, pr.storage)
		}
	}
	return n, err
}

func (h *HorizonStorage) Upload(ctx context.Context, file any, cb ProgressCallback) (*Storage, error) {
	switch v := file.(type) {
	case string:
		return h.UploadFromPath(ctx, v, cb)
	case []byte:
		return h.UploadFromBinary(ctx, v, cb)
	case *multipart.FileHeader:
		return h.UploadFromHeader(ctx, v, cb)
	default:
		return nil, eris.Errorf("unsupported type: %T", file)
	}
}

func (h *HorizonStorage) UploadFromPath(ctx context.Context, path string, cb ProgressCallback) (*Storage, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to open %s", path)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, eris.Wrapf(err, "failed to stat %s", path)
	}

	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil && err != io.EOF {
		return nil, eris.Wrap(err, "content type detection failed")
	}
	contentType := http.DetectContentType(buf)
	file.Seek(0, 0)

	fileName, err := h.GenerateUniqueName(ctx, filepath.Base(path))
	if err != nil {
		return nil, err
	}

	storage := &Storage{
		FileName:   fileName,
		FileSize:   info.Size(),
		FileType:   contentType,
		StorageKey: fileName,
		BucketName: h.storageBucket,
		Status:     "progress",
	}

	pr := &progressReader{
		reader:   file,
		callback: cb,
		total:    info.Size(),
		storage:  storage,
	}

	obj := h.bucket.Object(fileName)
	w := obj.NewWriter(ctx)

	if _, err := io.Copy(w, pr); err != nil {
		return nil, eris.Wrap(err, "upload failed")
	}
	if err := w.Close(); err != nil {
		return nil, eris.Wrap(err, "upload finalization failed")
	}

	url, err := h.GeneratePresignedURL(ctx, storage, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	storage.URL = url
	storage.Status = "completed"
	return storage, nil
}

func (h *HorizonStorage) UploadFromBinary(ctx context.Context, data []byte, cb ProgressCallback) (*Storage, error) {
	contentType := http.DetectContentType(data)
	fileName, err := h.GenerateUniqueName(ctx, "file")
	if err != nil {
		return nil, err
	}

	storage := &Storage{
		FileName:   fileName,
		FileSize:   int64(len(data)),
		FileType:   contentType,
		StorageKey: fileName,
		BucketName: h.storageBucket,
		Status:     "progress",
	}

	pr := &progressReader{
		reader:   bytes.NewReader(data),
		callback: cb,
		total:    int64(len(data)),
		storage:  storage,
	}

	obj := h.bucket.Object(fileName)
	w := obj.NewWriter(ctx)

	if _, err := io.Copy(w, pr); err != nil {
		return nil, eris.Wrap(err, "upload failed")
	}
	if err := w.Close(); err != nil {
		return nil, eris.Wrap(err, "upload finalization failed")
	}

	url, err := h.GeneratePresignedURL(ctx, storage, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	storage.URL = url
	storage.Status = "completed"
	return storage, nil
}

func (h *HorizonStorage) UploadFromHeader(ctx context.Context, header *multipart.FileHeader, cb ProgressCallback) (*Storage, error) {
	file, err := header.Open()
	if err != nil {
		return nil, eris.Wrap(err, "failed to open multipart file")
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	fileName, err := h.GenerateUniqueName(ctx, header.Filename)
	if err != nil {
		return nil, err
	}

	storage := &Storage{
		FileName:   fileName,
		FileSize:   header.Size,
		FileType:   contentType,
		StorageKey: fileName,
		BucketName: h.storageBucket,
		Status:     "progress",
	}

	pr := &progressReader{
		reader:   file,
		callback: cb,
		total:    header.Size,
		storage:  storage,
	}

	obj := h.bucket.Object(fileName)
	w := obj.NewWriter(ctx)

	if _, err := io.Copy(w, pr); err != nil {
		return nil, eris.Wrap(err, "upload failed")
	}
	if err := w.Close(); err != nil {
		return nil, eris.Wrap(err, "upload finalization failed")
	}

	url, err := h.GeneratePresignedURL(ctx, storage, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	storage.URL = url
	storage.Status = "completed"
	return storage, nil
}

func (h *HorizonStorage) GeneratePresignedURL(ctx context.Context, storage *Storage, expiry time.Duration) (string, error) {
	obj := h.bucket.Object(storage.StorageKey)
	_, err := obj.Attrs(ctx)
	if err != nil {
		return "", eris.Wrap(errors.New("failed to presign URL"), "file does not exist")
	}
	authToken, err := h.bucket.AuthToken(ctx, h.prefix, expiry)
	if err != nil {
		return "", eris.Wrap(errors.New("failed to generate authorization token"), err.Error())
	}

	downloadURL := fmt.Sprintf("%s/file/%s/%s?Authorization=%s",
		h.bucket.BaseURL(),
		h.storageBucket,
		storage.StorageKey,
		authToken,
	)
	return downloadURL, nil
}
func (h *HorizonStorage) DeleteFile(ctx context.Context, storage *Storage) error {
	obj := h.bucket.Object(storage.StorageKey)
	if err := obj.Delete(ctx); err != nil {
		return eris.Wrap(err, "file deletion failed")
	}
	return nil
}

func (h *HorizonStorage) GenerateUniqueName(ctx context.Context, original string) (string, error) {
	ext := filepath.Ext(original)
	base := strings.TrimSuffix(original, ext)
	return fmt.Sprintf("%s%d-%s%s", h.prefix, time.Now().UnixNano(), base, ext), nil
}
