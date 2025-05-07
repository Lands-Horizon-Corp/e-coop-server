package horizon

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rotisserie/eris"
	"go.uber.org/zap"
)

type StorageStatus string

const (
	StorageStatusPending   StorageStatus = "pending"
	StorageStatusCancelled StorageStatus = "cancelled"
	StorageStatusCorrupt   StorageStatus = "corrupt"
	StorageStatusCompleted StorageStatus = "completed"
	StorageStatusProgress  StorageStatus = "progress"
)

type Storage struct {
	FileName   string
	FileSize   int64
	FileType   string
	StorageKey string
	URL        string
	BucketName string
	Status     StorageStatus
	Progress   int64
}

type BinaryFileInput struct {
	Data        io.Reader
	Size        int64
	Name        string
	ContentType string
}

type HorizonStorage struct {
	config   *HorizonConfig
	log      *HorizonLog
	security *HorizonSecurity
	storage  *minio.Client
}

type ProgressCallback func(progress int64, total int64, storage *Storage)

type progressReader struct {
	reader   io.Reader
	callback ProgressCallback
	total    int64
	storage  *Storage
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	if pr.callback != nil {
		pr.storage.Progress = int64(n)
		pr.callback(int64(n), pr.total, pr.storage)
	}
	return n, err
}

func NewHorizonStorage(config *HorizonConfig, log *HorizonLog, security *HorizonSecurity) (*HorizonStorage, error) {
	return &HorizonStorage{
		config:   config,
		log:      log,
		security: security,
	}, nil
}

// Initializes the MinIO storage client and ensures the bucket exists.
func (hs *HorizonStorage) Run() error {
	ctx := context.Background()

	client, err := minio.New(hs.config.StorageEndpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			hs.config.StorageAccessKey,
			hs.config.StorageSecretKey,
			"",
		),
		Secure: false,
		Region: hs.config.StorageRegion,
		BucketLookup: func() minio.BucketLookupType {
			if hs.config.StorageDriver == "s3" {
				return minio.BucketLookupDNS
			}
			return minio.BucketLookupPath
		}(),
	})
	if err != nil {
		return eris.Wrap(err, "failed to initialize MinIO client")
	}

	hs.storage = client

	exists, err := client.BucketExists(ctx, hs.config.StorageBucket)
	if err != nil {
		return eris.Wrap(err, "failed to check if bucket exists")
	}

	if !exists {
		err = client.MakeBucket(ctx, hs.config.StorageBucket, minio.MakeBucketOptions{
			Region: hs.config.StorageRegion,
		})
		if err != nil {
			return eris.Wrapf(err, "failed to create bucket: %s", hs.config.StorageBucket)
		}
	}

	return nil
}

// Releases storage client.
func (hs *HorizonStorage) Stop() {
	hs.storage = nil
}

// Upload determines the source of file input and delegates to the appropriate method.
func (hs *HorizonStorage) Upload(file any, onProgress ProgressCallback) (*Storage, error) {
	var storage *Storage
	var err error

	switch v := file.(type) {
	case *multipart.FileHeader:
		storage, err = hs.UploadFromHeader(v, onProgress)
	case string:
		v = strings.TrimSpace(v)
		switch {
		case isValidURL(v):
			storage, err = hs.UploadFromURL(v, onProgress)
		case isValidFilePath(v) == nil:
			storage, err = hs.UploadLocalFile(v, onProgress)
		default:
			err = eris.Errorf("invalid string input: must be a valid file path or URL: %s", v)
		}
	case BinaryFileInput:
		storage, err = hs.UploadFromBinary(v, onProgress)
	default:
		err = eris.New("unsupported file input type")
	}
	if err != nil {
		hs.log.Log(LogEntry{
			Category: CategoryStorage,
			Level:    LevelInfo,
			Message:  fmt.Sprintf("successfully uploaded %s", storage.FileName),
			Fields: []zap.Field{
				zap.String("file-name", storage.FileName),
				zap.Int64("file-size", storage.FileSize),
				zap.String("file-type", storage.FileType),
				zap.String("storage-key", storage.StorageKey),
				zap.String("url", storage.URL),
			},
		})
	}
	return storage, err
}

func (hs *HorizonStorage) UploadFromBinary(input BinaryFileInput, onProgress ProgressCallback) (*Storage, error) {
	if input.Data == nil {
		return nil, eris.New("binary input data is nil")
	}
	contentType := input.ContentType
	if contentType == "" {
		buf := make([]byte, 512)
		n, err := input.Data.Read(buf)
		if err != nil && err != io.EOF {
			return nil, eris.Wrap(err, "failed to read binary input")
		}
		contentType = http.DetectContentType(buf[:n])
		if seeker, ok := input.Data.(io.Seeker); ok {
			_, _ = seeker.Seek(0, io.SeekStart)
		} else {
			input.Data = io.MultiReader(bytes.NewReader(buf[:n]), input.Data)
		}
	}

	fileName := hs.storageName(input.Name)

	progress := &progressReader{
		reader:   input.Data,
		callback: onProgress,
		total:    input.Size,
	}

	ctx := context.Background()
	info, err := hs.storage.PutObject(ctx, hs.config.StorageBucket, fileName, progress, input.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to upload binary input")
	}

	url, err := hs.GeneratePresignedURL(fileName)
	if err != nil {
		return nil, eris.Wrap(err, "failed to generate presigned URL for binary input")
	}

	return &Storage{
		FileName:   input.Name,
		FileSize:   info.Size,
		FileType:   contentType,
		StorageKey: info.Key,
		URL:        url,
		BucketName: hs.config.StorageBucket,
		Status:     StorageStatusCompleted,
		Progress:   progress.total,
	}, nil
}

func (hs *HorizonStorage) UploadFromHeader(file *multipart.FileHeader, onProgress ProgressCallback) (*Storage, error) {
	src, err := file.Open()
	if err != nil {
		return nil, eris.Wrapf(err, "failed to open file: %s", file.Filename)
	}
	defer src.Close()

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	fileName := hs.storageName(file.Filename)

	progress := &progressReader{
		reader:   src,
		callback: onProgress,
		total:    file.Size,
	}

	ctx := context.Background()
	info, err := hs.storage.PutObject(ctx, hs.config.StorageBucket, fileName, progress, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, eris.Wrapf(err, "failed to upload file: %s", file.Filename)
	}

	url, err := hs.GeneratePresignedURL(fileName)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to generate presigned URL for: %s", file.Filename)
	}

	return &Storage{
		FileName:   file.Filename,
		FileSize:   info.Size,
		FileType:   contentType,
		StorageKey: info.Key,
		URL:        url,
		BucketName: hs.config.StorageBucket,
		Status:     StorageStatusCompleted,
		Progress:   progress.total,
	}, nil
}

func (hs *HorizonStorage) UploadLocalFile(filePath string, onProgress ProgressCallback) (*Storage, error) {
	filePath = strings.TrimSpace(filePath)

	if err := isValidFilePath(filePath); err != nil {
		return nil, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to open file: %s", filePath)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, eris.Wrapf(err, "failed to stat file: %s", filePath)
	}

	buf := make([]byte, 512)
	_, _ = file.Read(buf)
	contentType := http.DetectContentType(buf)
	_, _ = file.Seek(0, io.SeekStart)

	fileName := hs.storageName(filepath.Base(filePath))

	progress := &progressReader{
		reader:   file,
		callback: onProgress,
		total:    stat.Size(),
	}

	ctx := context.Background()
	info, err := hs.storage.PutObject(ctx, hs.config.StorageBucket, fileName, progress, stat.Size(), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to upload file to storage")
	}

	url, err := hs.GeneratePresignedURL(fileName)
	if err != nil {
		return nil, eris.Wrap(err, "failed to generate presigned URL")
	}

	return &Storage{
		FileName:   info.Key,
		FileSize:   info.Size,
		FileType:   contentType,
		StorageKey: info.Key,
		URL:        url,
		BucketName: hs.config.StorageBucket,
		Status:     StorageStatusCompleted,
		Progress:   progress.total,
	}, nil
}

func (hs *HorizonStorage) UploadFromURL(value string, onProgress ProgressCallback) (*Storage, error) {
	value = strings.TrimSpace(value)

	if !isValidURL(value) {
		return nil, eris.Errorf("invalid URL: %s", value)
	}

	resp, err := http.Get(value)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to download file from URL: %s", value)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, eris.Errorf("non-200 response when downloading file: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	fileName := hs.storageName(path.Base(value))

	progress := &progressReader{
		reader:   resp.Body,
		callback: onProgress,
		total:    resp.ContentLength,
	}

	ctx := context.Background()
	info, err := hs.storage.PutObject(ctx, hs.config.StorageBucket, fileName, progress, resp.ContentLength, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, eris.Wrapf(err, "failed to upload file from URL: %s", value)
	}

	url, err := hs.GeneratePresignedURL(fileName)
	if err != nil {
		return nil, eris.Wrap(err, "failed to generate presigned URL")
	}

	return &Storage{
		FileName:   fileName,
		FileSize:   info.Size,
		FileType:   contentType,
		StorageKey: info.Key,
		URL:        url,
		BucketName: hs.config.StorageBucket,
	}, nil
}

func (hs *HorizonStorage) GeneratePresignedURL(fileName string) (string, error) {
	ctx := context.Background()

	presignedURL, err := hs.storage.PresignedGetObject(ctx, hs.config.StorageBucket, fileName, 24*time.Hour, nil)
	if err != nil {
		return "", eris.Wrap(err, "failed to generate presigned URL")
	}

	return presignedURL.String(), nil
}

// DeleteFile deletes a file from the storage bucket using the given storage key.
func (hs *HorizonStorage) DeleteFile(storageKey string) error {
	if hs.storage == nil {
		return eris.New("storage client is not initialized")
	}

	if strings.TrimSpace(storageKey) == "" {
		return eris.New("storage key is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := hs.storage.RemoveObject(ctx, hs.config.StorageBucket, storageKey, minio.RemoveObjectOptions{})
	if err != nil {
		return eris.Wrapf(err, "failed to delete file: %s", storageKey)
	}

	hs.log.Log(LogEntry{
		Category: CategoryStorage,
		Level:    LevelInfo,
		Message:  fmt.Sprintf("successfully deleted file: %s", storageKey),
		Fields: []zap.Field{
			zap.String("storage-key", storageKey),
			zap.String("bucket", hs.config.StorageBucket),
		},
	})

	return nil
}

func (hs *HorizonStorage) storageName(originalFileName string) string {
	return fmt.Sprintf("%s-%d-%s", time.Now().Format("20060102150405"), os.Getpid(), originalFileName)
}

func isValidFilePath(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return errors.New("file does not exist")
	}
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errors.New("path is a directory, not a file")
	}
	return nil
}

func isValidURL(value string) bool {
	_, err := url.ParseRequestURI(value)
	return err == nil
}
