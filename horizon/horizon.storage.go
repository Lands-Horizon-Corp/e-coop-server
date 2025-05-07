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
		// Calculate percentage 0-100
		percent := pr.readSoFar * 100 / pr.total
		if percent > 100 {
			percent = 100
		}
		pr.storage.Progress = percent
		if pr.callback != nil {
			// Report percent completion
			pr.callback(percent, 100, pr.storage)
		}
	}
	return n, err
}

func NewHorizonStorage(config *HorizonConfig, log *HorizonLog, security *HorizonSecurity) (*HorizonStorage, error) {
	return &HorizonStorage{config: config, log: log, security: security}, nil
}

// Run initializes the MinIO client and ensures the bucket exists.
func (hs *HorizonStorage) Run() error {
	ctx := context.Background()
	client, err := minio.New(hs.config.StorageEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(hs.config.StorageAccessKey, hs.config.StorageSecretKey, ""),
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
		return eris.Wrap(err, "failed to check bucket exists")
	}
	if !exists {
		err = client.MakeBucket(ctx, hs.config.StorageBucket, minio.MakeBucketOptions{Region: hs.config.StorageRegion})
		if err != nil {
			return eris.Wrapf(err, "failed to create bucket %s", hs.config.StorageBucket)
		}
	}
	return nil
}

// Stop cleans up the storage client.
func (hs *HorizonStorage) Stop() error {
	hs.storage = nil
	return nil
}

// Upload dispatches to the correct upload method based on input type.
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
			err = eris.Errorf("invalid string input %s", v)
		}
	case BinaryFileInput:
		storage, err = hs.UploadFromBinary(v, onProgress)
	default:
		err = eris.New("unsupported file input type")
	}
	if err == nil {
		hs.log.Log(LogEntry{Category: CategoryStorage, Level: LevelInfo, Message: fmt.Sprintf("uploaded %s", storage.FileName), Fields: []zap.Field{zap.String("key", storage.StorageKey)}})
	}
	return storage, err
}

// UploadFromBinary handles io.Reader streams.
func (hs *HorizonStorage) UploadFromBinary(input BinaryFileInput, onProgress ProgressCallback) (*Storage, error) {
	if input.Data == nil {
		return nil, eris.New("binary input nil")
	}
	contentType := input.ContentType
	if contentType == "" {
		buf := make([]byte, 512)
		n, _ := input.Data.Read(buf)
		contentType = http.DetectContentType(buf[:n])
		if s, ok := input.Data.(io.Seeker); ok {
			s.Seek(0, io.SeekStart)
		} else {
			input.Data = io.MultiReader(bytes.NewReader(buf[:n]), input.Data)
		}
	}
	fileName := hs.storageName(input.Name)
	storage := &Storage{FileName: fileName, FileSize: input.Size, FileType: contentType, Status: StorageStatusProgress}
	pr := &progressReader{reader: input.Data, callback: onProgress, total: input.Size, storage: storage}
	ctx := context.Background()
	info, err := hs.storage.PutObject(ctx, hs.config.StorageBucket, fileName, pr, input.Size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return nil, eris.Wrap(err, "upload binary failed")
	}
	url, _ := hs.GeneratePresignedURL(fileName)
	storage.StorageKey, storage.URL, storage.BucketName = info.Key, url, hs.config.StorageBucket
	storage.Status, storage.Progress = StorageStatusCompleted, info.Size
	return storage, nil
}

// UploadFromHeader handles multipart uploads.
func (hs *HorizonStorage) UploadFromHeader(file *multipart.FileHeader, onProgress ProgressCallback) (*Storage, error) {
	src, err := file.Open()
	if err != nil {
		return nil, eris.Wrapf(err, "open %s failed", file.Filename)
	}
	defer src.Close()
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	fileName := hs.storageName(file.Filename)
	storage := &Storage{FileName: fileName, FileSize: file.Size, FileType: contentType, Status: StorageStatusProgress}
	pr := &progressReader{reader: src, callback: onProgress, total: file.Size, storage: storage}
	ctx := context.Background()
	info, err := hs.storage.PutObject(ctx, hs.config.StorageBucket, fileName, pr, file.Size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return nil, eris.Wrapf(err, "upload %s failed", file.Filename)
	}
	url, _ := hs.GeneratePresignedURL(fileName)
	storage.StorageKey, storage.URL, storage.BucketName = info.Key, url, hs.config.StorageBucket
	storage.Status, storage.Progress = StorageStatusCompleted, info.Size
	return storage, nil
}

// UploadLocalFile handles file paths.
func (hs *HorizonStorage) UploadLocalFile(filePath string, onProgress ProgressCallback) (*Storage, error) {
	filePath = strings.TrimSpace(filePath)
	if err := isValidFilePath(filePath); err != nil {
		return nil, err
	}
	file, err := os.Open(filePath)
	if err != nil {
		return nil, eris.Wrapf(err, "open %s failed", filePath)
	}
	defer file.Close()
	stat, _ := file.Stat()
	buf := make([]byte, 512)
	file.Read(buf)
	contentType := http.DetectContentType(buf)
	file.Seek(0, io.SeekStart)
	fileName := hs.storageName(filepath.Base(filePath))
	storage := &Storage{FileName: fileName, FileSize: stat.Size(), FileType: contentType, Status: StorageStatusProgress}
	pr := &progressReader{reader: file, callback: onProgress, total: stat.Size(), storage: storage}
	ctx := context.Background()
	info, err := hs.storage.PutObject(ctx, hs.config.StorageBucket, fileName, pr, stat.Size(), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return nil, eris.Wrap(err, "upload local failed")
	}
	url, _ := hs.GeneratePresignedURL(fileName)
	storage.StorageKey, storage.URL, storage.BucketName = info.Key, url, hs.config.StorageBucket
	storage.Status, storage.Progress = StorageStatusCompleted, info.Size
	return storage, nil
}

// UploadFromURL handles remote URLs.
func (hs *HorizonStorage) UploadFromURL(value string, onProgress ProgressCallback) (*Storage, error) {
	value = strings.TrimSpace(value)
	if !isValidURL(value) {
		return nil, eris.Errorf("invalid URL %s", value)
	}
	resp, err := http.Get(value)
	if err != nil {
		return nil, eris.Wrapf(err, "download %s failed", value)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, eris.Errorf("download status %d", resp.StatusCode)
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	fileName := hs.storageName(path.Base(value))
	storage := &Storage{FileName: fileName, FileSize: resp.ContentLength, FileType: contentType, Status: StorageStatusProgress}
	pr := &progressReader{reader: resp.Body, callback: onProgress, total: resp.ContentLength, storage: storage}
	ctx := context.Background()
	info, err := hs.storage.PutObject(ctx, hs.config.StorageBucket, fileName, pr, resp.ContentLength, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return nil, eris.Wrapf(err, "upload URL %s failed", value)
	}
	url, _ := hs.GeneratePresignedURL(fileName)
	storage.StorageKey, storage.URL, storage.BucketName = info.Key, url, hs.config.StorageBucket
	storage.Status, storage.Progress = StorageStatusCompleted, info.Size
	return storage, nil
}

// GeneratePresignedURL returns a temporary GET URL.
func (hs *HorizonStorage) GeneratePresignedURL(fileName string) (string, error) {
	ctx := context.Background()
	u, err := hs.storage.PresignedGetObject(ctx, hs.config.StorageBucket, fileName, 24*time.Hour, nil)
	if err != nil {
		return "", eris.Wrap(err, "presign failed")
	}
	return u.String(), nil
}

// DeleteFile removes an object.
func (hs *HorizonStorage) DeleteFile(key string) error {
	if hs.storage == nil {
		return eris.New("not initialized")
	}
	if strings.TrimSpace(key) == "" {
		return eris.New("empty key")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := hs.storage.RemoveObject(ctx, hs.config.StorageBucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return eris.Wrapf(err, "delete %s failed", key)
	}
	hs.log.Log(LogEntry{Category: CategoryStorage, Level: LevelInfo, Message: fmt.Sprintf("deleted %s", key)})
	return nil
}

func (hs *HorizonStorage) storageName(original string) string {
	return fmt.Sprintf("%s-%d-%s", time.Now().Format("20060102150405"), os.Getpid(), original)
}

func isValidFilePath(p string) error {
	info, err := os.Stat(p)
	if os.IsNotExist(err) {
		return errors.New("not exist")
	}
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errors.New("is dir")
	}
	return nil
}

func isValidURL(u string) bool {
	_, err := url.ParseRequestURI(u)
	return err == nil
}
