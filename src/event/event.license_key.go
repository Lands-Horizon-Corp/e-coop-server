package event

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	core_admin "github.com/Lands-Horizon-Corp/e-coop-server/src/db/admin"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

func ActivateLicense(
	ctx context.Context,
	service *horizon.HorizonService,
	licenseKey string,
	fingerprint string,
) (string, error) {
	license, err := core_admin.LicenseManager(service).FindOne(ctx, &types.License{
		LicenseKey: licenseKey,
	})
	if err != nil {
		return "", eris.Wrap(err, "license key not found")
	}
	if license.IsUsed {
		return "", eris.New("license key already used")
	}
	if license.IsRevoked {
		return "", eris.New("license key is revoked")
	}
	if license.ExpiresAt != nil && license.ExpiresAt.Before(time.Now().UTC()) {
		return "", eris.New("license key has expired")
	}
	now := time.Now().UTC()
	license.IsUsed = true
	license.UsedAt = &now
	if err := core_admin.LicenseManager(service).UpdateByID(ctx, license.ID, license); err != nil {
		return "", eris.Wrap(err, "failed to update license")
	}
	secretKey := uuid.New().String()
	if fingerprint != "" {
		hashedFingerprint := fmt.Sprintf("%x", sha256.Sum256([]byte(fingerprint)))
		redisKey := fmt.Sprintf("license:activation:%s:%s", hashedFingerprint, secretKey)
		ttl := 365 * 24 * time.Hour
		licenseData, err := json.Marshal(license)
		if err != nil {
			return "", eris.Wrap(err, "failed to marshal license data")
		}
		if err := service.Cache.Set(ctx, redisKey, licenseData, ttl); err != nil {
			return "", eris.Wrap(err, "failed to store activation in Redis")
		}
		reverseKey := fmt.Sprintf("license:secret:%s", secretKey)
		if err := service.Cache.Set(ctx, reverseKey, hashedFingerprint, ttl); err != nil {
			return "", eris.Wrap(err, "failed to store reverse mapping in Redis")
		}
	}
	return secretKey, nil
}

func VerifyLicenseByFingerprint(
	ctx context.Context,
	service *horizon.HorizonService,
	secretKey string,
	fingerprint string,
) (*types.License, error) {
	if secretKey == "" {
		return nil, eris.New("secret key cannot be empty")
	}
	if fingerprint == "" {
		return nil, eris.New("fingerprint cannot be empty")
	}
	hashedFingerprint := fmt.Sprintf("%x", sha256.Sum256([]byte(fingerprint)))
	reverseKey := fmt.Sprintf("license:secret:%s", secretKey)
	storedFingerprint, err := service.Cache.Get(ctx, reverseKey)
	if err != nil {
		redisKey := fmt.Sprintf("license:activation:%s:%s", hashedFingerprint, secretKey)
		licenseData, err := service.Cache.Get(ctx, redisKey)
		if err != nil {
			return nil, eris.Wrap(err, "failed to get license from Redis")
		}
		if licenseData == nil {
			return nil, eris.New("license not found for fingerprint and secret key")
		}

		var license types.License
		if err := json.Unmarshal(licenseData, &license); err != nil {
			return nil, eris.Wrap(err, "failed to unmarshal license data")
		}
		return &license, nil
	}
	if string(storedFingerprint) != hashedFingerprint {
		return nil, eris.New("fingerprint does not match secret key")
	}
	redisKey := fmt.Sprintf("license:activation:%s:%s", hashedFingerprint, secretKey)
	licenseData, err := service.Cache.Get(ctx, redisKey)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get license from Redis")
	}
	if licenseData == nil {
		return nil, eris.New("license not found")
	}
	var license types.License
	if err := json.Unmarshal(licenseData, &license); err != nil {
		return nil, eris.Wrap(err, "failed to unmarshal license data")
	}
	return &license, nil
}

func DeactivateLicense(
	ctx context.Context,
	service *horizon.HorizonService,
	secretKey string,
	fingerprint string,
) error {
	if secretKey == "" {
		return eris.New("secret key cannot be empty")
	}
	if fingerprint == "" {
		return eris.New("fingerprint cannot be empty")
	}
	hashedFingerprint := fmt.Sprintf("%x", sha256.Sum256([]byte(fingerprint)))
	redisKey := fmt.Sprintf("license:activation:%s:%s", hashedFingerprint, secretKey)
	licenseData, err := service.Cache.Get(ctx, redisKey)
	if err != nil {
		return eris.Wrap(err, "failed to get license from Redis")
	}
	if licenseData == nil {
		return eris.New("license activation not found")
	}
	var license *types.License
	if err := json.Unmarshal(licenseData, &license); err != nil {
		return eris.Wrap(err, "failed to unmarshal license data")
	}
	license.IsUsed = false
	license.UsedAt = nil
	license.IsRevoked = true
	if err := core_admin.LicenseManager(service).UpdateByID(ctx, license.ID, license); err != nil {
		return eris.Wrap(err, "failed to update license in database")
	}
	if err := service.Cache.Delete(ctx, redisKey); err != nil {
		return eris.Wrap(err, "failed to remove activation from Redis")
	}
	reverseKey := fmt.Sprintf("license:secret:%s", secretKey)
	if err := service.Cache.Delete(ctx, reverseKey); err != nil {
		return eris.Wrap(err, "Warning: failed to delete reverse mapping")
	}
	return nil
}
