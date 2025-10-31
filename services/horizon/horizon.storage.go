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

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
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
	UploadFromURL(ctx context.Context, url string, cb ProgressCallback) (*Storage, error)
	GeneratePresignedURL(ctx context.Context, storage *Storage, expiry time.Duration) (string, error)
	DeleteFile(ctx context.Context, storage *Storage) error
	RemoveAllFiles(ctx context.Context) error
	GenerateUniqueName(ctx context.Context, originalName string, contentType string) (string, error)
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
	ssl              bool
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

func NewHorizonStorageService(
	accessKey,
	secretKey,
	endpoint,
	bucket,
	region,
	driver string,
	maxSize int64,
	ssl bool,
) StorageService {
	return &HorizonStorage{
		driver:           driver,
		storageAccessKey: accessKey,
		storageSecretKey: secretKey,
		endpoint:         endpoint,
		storageBucket:    bucket,
		region:           region,
		maxFileSize:      maxSize,
		ssl:              ssl,
	}
}

// Run initializes the storage service and ensures the bucket exists
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
		Secure: h.ssl,
		Region: h.region,
		BucketLookup: func() minio.BucketLookupType {
			if h.driver == "s3" {
				return minio.BucketLookupDNS
			}
			return minio.BucketLookupPath
		}(),
	})
	if err != nil {
		return eris.Wrapf(err, "failed to initialize %s client", h.driver)
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

// Stop cleans up the storage service resources
func (h *HorizonStorage) Stop(ctx context.Context) error {
	h.client = nil
	return nil
}

// Read reads data from the underlying reader and reports progress
func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.readSoFar += int64(n)
		percent := pr.readSoFar * 100 / pr.total
		percent = min(percent, 100)
		pr.storage.Progress = percent
		if pr.callback != nil {
			pr.callback(percent, 100, pr.storage)
		}
	}
	return n, err
}

// Upload uploads a file to the storage service based on the input type
func (h *HorizonStorage) Upload(ctx context.Context, file any, onProgress ProgressCallback) (*Storage, error) {
	switch v := file.(type) {
	case string:
		if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") {
			return h.UploadFromURL(ctx, v, onProgress)
		}
		return h.UploadFromPath(ctx, v, onProgress)
	case []byte:
		return h.UploadFromBinary(ctx, v, onProgress)
	case *multipart.FileHeader:
		return h.UploadFromHeader(ctx, v, onProgress)
	default:
		return nil, eris.Errorf("unsupported type: %T", file)
	}
}

// UploadFromPath uploads a file from a local path to the storage service
func (h *HorizonStorage) UploadFromPath(ctx context.Context, path string, cb ProgressCallback) (*Storage, error) {
	if handlers.IsSuspiciousPath(path) {
		return nil, eris.New("suspicious file path")
	}
	file, err := os.Open(path) // #nosec G304 -- path validated above with handlers.IsSuspiciousPath
	if err != nil {
		return nil, eris.Wrapf(err, "failed to open %s", path)
	}
	defer func() { _ = file.Close() }()

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
	if _, err := file.Seek(0, 0); err != nil {
		return nil, eris.Wrap(err, "failed to seek file to beginning")
	}
	fileName, err := h.GenerateUniqueName(ctx, filepath.Base(path), contentType)
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

// UploadFromBinary uploads a file from binary data to the storage service
func (h *HorizonStorage) UploadFromBinary(ctx context.Context, data []byte, cb ProgressCallback) (*Storage, error) {
	contentType := http.DetectContentType(data)
	fileName, err := h.GenerateUniqueName(ctx, "file", contentType)
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

// UploadFromURL uploads a file from a URL to the storage service
func (h *HorizonStorage) UploadFromURL(ctx context.Context, url string, cb ProgressCallback) (*Storage, error) {
	// Download the file from the URL
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, eris.Wrap(err, "failed to create HTTP request for URL")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to download file from URL: %s", url)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, eris.Errorf("failed to download file from URL: %s, status: %s", url, resp.Status)
	}
	fileName := ""
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if parts := strings.Split(cd, "filename="); len(parts) > 1 {
			fileName = strings.Trim(parts[1], "\"")
		}
	}
	if fileName == "" {
		fileName = filepath.Base(resp.Request.URL.Path)
	}
	if fileName == "" {
		fileName = "file"
	}

	// Detect content type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Read file data to buffer to get the size (since MinIO needs content length)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, eris.Wrap(err, "failed to read file data from URL response")
	}

	fileName, err = h.GenerateUniqueName(ctx, fileName, contentType)
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
		reader:   strings.NewReader(string(data)),
		callback: cb,
		total:    int64(len(data)),
		storage:  storage,
	}

	result, err := h.client.PutObject(ctx, h.storageBucket, fileName, pr, int64(len(data)), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return nil, eris.Wrap(err, "upload from URL failed")
	}
	storage.StorageKey = result.Key
	storage.BucketName = result.Bucket

	urlStr, err := h.GeneratePresignedURL(ctx, storage, 24*time.Hour)
	if err != nil {
		return nil, err
	}
	storage.URL = urlStr
	storage.Status = "completed"
	return storage, nil
}

// UploadFromHeader uploads a file from a multipart header to the storage service
func (h *HorizonStorage) UploadFromHeader(ctx context.Context, header *multipart.FileHeader, cb ProgressCallback) (*Storage, error) {
	file, err := header.Open()
	if err != nil {
		return nil, eris.Wrap(err, "failed to open multipart file")
	}
	defer func() { _ = file.Close() }()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	fileName, err := h.GenerateUniqueName(ctx, header.Filename, contentType)
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

// GeneratePresignedURL generates a presigned URL for accessing the stored file
func (h *HorizonStorage) GeneratePresignedURL(ctx context.Context, storage *Storage, expiry time.Duration) (string, error) {
	_, err := h.client.StatObject(ctx, storage.BucketName, storage.StorageKey, minio.StatObjectOptions{})
	if err != nil {
		return "", eris.Wrapf(err, "failed to generate presigned URL for key %s in bucket %s", storage.StorageKey, storage.BucketName)
	}
	presignedURL, err := h.client.PresignedGetObject(ctx, storage.BucketName, storage.StorageKey, expiry, nil)
	if err != nil {
		return "", eris.Wrap(err, "failed to generate presigned URL")
	}
	return presignedURL.String(), nil
}

// DeleteFile deletes a file from the storage service
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

// RemoveAllFiles removes all files from the storage bucket
func (h *HorizonStorage) RemoveAllFiles(ctx context.Context) error {
	if h.client == nil {
		return eris.New("not initialized")
	}
	if strings.TrimSpace(h.storageBucket) == "" {
		return eris.New("empty bucket name")
	}

	objectsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objectsCh)
		opts := minio.ListObjectsOptions{
			Recursive: true,
		}
		for object := range h.client.ListObjects(ctx, h.storageBucket, opts) {
			if object.Err != nil {
				continue
			}
			select {
			case objectsCh <- object:
			case <-ctx.Done():
				return
			}
		}
	}()
	errorCh := h.client.RemoveObjects(ctx, h.storageBucket, objectsCh, minio.RemoveObjectsOptions{})
	var errors []error
	for err := range errorCh {
		if err.Err != nil {
			errors = append(errors, eris.Wrapf(err.Err, "failed to delete object %s", err.ObjectName))
		}
	}
	if len(errors) > 0 {
		return eris.Errorf("failed to delete %d objects from bucket %s", len(errors), h.storageBucket)
	}
	return nil
}

// GenerateUniqueName generates a unique file name based on the original name and content type
func (h *HorizonStorage) GenerateUniqueName(_ context.Context, original string, contentType string) (string, error) {
	ext := filepath.Ext(original)
	base := strings.TrimSuffix(original, ext)

	// If no extension in original filename, try to get it from content type
	if ext == "" && contentType != "" {
		ext = handlers.GetExtensionFromContentType(contentType)
	}

	return fmt.Sprintf("%s%d-%s%s", h.prefix, time.Now().UnixNano(), base, ext), nil
}
