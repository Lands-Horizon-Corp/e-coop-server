// Package seeder provides utilities to populate the database with
// sample data for development and testing environments.
package seeder

import (
	"context"
	crand "crypto/rand"
	"io/fs"
	"math/big"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/modelcore"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/horizon"
	"github.com/rotisserie/eris"
)

// loadImagePaths scans the seeder/images directory and loads all image file paths
func (s *Seeder) loadImagePaths() error {
	imagesDir := "seeder/images"
	supportedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}

	err := filepath.WalkDir(imagesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if supportedExtensions[ext] {
				s.imagePaths = append(s.imagePaths, path)
			}
		}

		return nil
	})

	if err != nil {
		return eris.Wrap(err, "failed to scan images directory")
	}

	if len(s.imagePaths) == 0 {
		return eris.New("no image files found in seeder/images directory")
	}

	return nil
}

func (s *Seeder) createImageMedia(ctx context.Context, imageType string) (*modelcore.Media, error) {
	if len(s.imagePaths) == 0 {
		return nil, eris.New("no image files available for seeding")
	}

	// Randomly choose one image from the loaded paths using crypto/rand
	maxInt := big.NewInt(int64(len(s.imagePaths)))
	nBig, err := crand.Int(crand.Reader, maxInt)
	if err != nil {
		return nil, eris.Wrap(err, "failed to generate secure random index for image selection")
	}
	randomIndex := int(nBig.Int64())
	imagePath := s.imagePaths[randomIndex]

	// Upload the image from local path
	storage, err := s.provider.Service.Storage.UploadFromPath(ctx, imagePath, func(_ int64, _ int64, _ *horizon.Storage) {})
	if err != nil {
		return nil, eris.Wrapf(err, "failed to upload image from path %s for %s", imagePath, imageType)
	} // Create media record
	media := &modelcore.Media{
		FileName:   storage.FileName,
		FileType:   storage.FileType,
		FileSize:   storage.FileSize,
		StorageKey: storage.StorageKey,
		URL:        storage.URL,
		BucketName: storage.BucketName,
		Status:     "completed",
		Progress:   100,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	if err := s.modelcore.MediaManager.Create(ctx, media); err != nil {
		return nil, eris.Wrap(err, "failed to create media record")
	}

	return media, nil
}

func ptr[T any](v T) *T {
	return &v
}
