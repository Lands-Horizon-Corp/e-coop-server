package horizon

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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
	driver           string
	storageAccessKey string
	storageSecretKey string
	storageBucket    string
	endpoint         string
	prefix           string
	region           string
	maxFileSize      int64
	client           *minio.Client
}

func NewHorizonStorageService(
	accessKey,
	secretKey,
	endpoint,
	bucket,
	region string,
	maxSize int64,
) StorageService {
	return &HorizonStorage{
		driver:           "linodes",
		storageAccessKey: accessKey,
		storageSecretKey: secretKey,
		endpoint:         endpoint,
		storageBucket:    bucket,
		region:           region,
		maxFileSize:      maxSize,
	}
}

func (h *HorizonStorage) Run(ctx context.Context) error {
	// Check for missing keys
	if h.endpoint == "" {
		return eris.New("missing storage endpoint")
	}
	if h.storageAccessKey == "" {
		return eris.New("missing storage access key")
	}
	if h.storageSecretKey == "" {
		return eris.New("missing storage secret key")
	}
	if h.storageBucket == "" {
		return eris.New("missing storage bucket name")
	}

	// Initialize MinIO client
	client, err := minio.New(h.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(h.storageAccessKey, h.storageSecretKey, ""),
		Secure: false,
		Region: h.region,
		BucketLookup: func() minio.BucketLookupType {
			if h.driver == "s3" {
				return minio.BucketLookupDNS
			}
			return minio.BucketLookupPath
		}(),
	})
	if err != nil {
		return eris.Wrap(err, "failed to initialize MinIO client")
	}
	h.client = client

	// Check whether the bucket exists
	exists, err := client.BucketExists(ctx, h.storageBucket)
	if err != nil {
		return eris.Wrap(err, "failed to check bucket exists")
	}
	if !exists {
		// Create the bucket if it does not exist
		err = client.MakeBucket(ctx, h.storageBucket, minio.MakeBucketOptions{Region: h.region})
		if err != nil {
			return eris.Wrapf(err, "failed to create bucket %s", h.storageBucket)
		}
	}
	return nil
}

func (h *HorizonStorage) Stop(ctx context.Context) error {
	h.client = nil
	return nil
}

type progressReader struct {
	reader    io.Reader
	callback  ProgressCallback
	total     int64
	readSoFar int64
	storage   *Storage
}

type BinaryFileInput struct {
	Data        io.Reader
	Size        int64
	Name        string
	ContentType string
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

func (h *HorizonStorage) Upload(ctx context.Context, file any, onProgress ProgressCallback) (*Storage, error) {
	switch v := file.(type) {
	case string:
		return h.UploadFromPath(ctx, v, onProgress)
	case []byte:
		return h.UploadFromBinary(ctx, v, onProgress)
	case *multipart.FileHeader:
		return h.UploadFromHeader(ctx, v, onProgress)
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

	result, err := h.client.PutObject(ctx, h.storageBucket, fileName, pr, info.Size(), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return nil, eris.Wrap(err, "upload local failed")
	}
	storage.StorageKey = result.Key
	storage.BucketName = result.Bucket
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

	result, err := h.client.PutObject(ctx, h.storageBucket, fileName, pr, storage.FileSize, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return nil, eris.Wrap(err, "upload local failed")
	}
	storage.StorageKey = result.Key
	storage.BucketName = result.Bucket
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
	result, err := h.client.PutObject(ctx, h.storageBucket, fileName, pr, storage.FileSize, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return nil, eris.Wrap(err, "upload local failed")
	}
	storage.StorageKey = result.Key
	storage.BucketName = result.Bucket
	url, err := h.GeneratePresignedURL(ctx, storage, 24*time.Hour)
	if err != nil {
		return nil, err
	}
	storage.URL = url
	storage.Status = "completed"
	return storage, nil
}

func (h *HorizonStorage) GeneratePresignedURL(ctx context.Context, storage *Storage, expiry time.Duration) (string, error) {
	// Check if the file exists before generating the presigned URL
	_, err := h.client.StatObject(ctx, storage.BucketName, storage.StorageKey, minio.StatObjectOptions{})
	if err != nil {
		return "", eris.Wrapf(err, "failed to generate presigned URL for key %s in bucket %s", storage.StorageKey, storage.BucketName)
	}

	u, err := h.client.PresignedGetObject(ctx, storage.BucketName, storage.FileName, expiry, nil)
	if err != nil {
		return "", eris.Wrap(err, "presign failed")
	}
	return u.String(), nil
}

func (h *HorizonStorage) DeleteFile(ctx context.Context, storage *Storage) error {
	if h.client == nil {
		return eris.New("not initialized")
	}
	if strings.TrimSpace(storage.StorageKey) == "" {
		return eris.New("empty key")
	}
	err := h.client.RemoveObject(ctx, storage.BucketName, storage.StorageKey, minio.RemoveObjectOptions{})
	if err != nil {
		return eris.Wrapf(err, "failed to delete key %s from bucket %s", storage.StorageKey, storage.BucketName)
	}
	return nil
}

func (h *HorizonStorage) GenerateUniqueName(ctx context.Context, original string) (string, error) {
	ext := filepath.Ext(original)
	base := strings.TrimSuffix(original, ext)
	return fmt.Sprintf("%s%d-%s%s", h.prefix, time.Now().UnixNano(), base, ext), nil
}
