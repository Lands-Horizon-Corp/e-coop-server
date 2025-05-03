package horizon

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"

	"github.com/microcosm-cc/bluemonday"
	"github.com/rotisserie/eris"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type HorizonSecurity struct {
	config *HorizonConfig
	log    *HorizonLog
}

func NewHorizonSecurity(
	config *HorizonConfig,
	log *HorizonLog,
) (*HorizonSecurity, error) {
	return &HorizonSecurity{
		config: config,
		log:    log,
	}, nil
}

func (hs *HorizonSecurity) PasswordHash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		hs.log.Log(LogEntry{
			Category: CategorySecurity,
			Level:    LevelError,
			Message:  "Failed to hash password",
			Fields:   []zap.Field{zap.Error(err)},
		})
		return "", eris.Wrap(err, "failed to hash password")
	}
	hs.log.Log(LogEntry{
		Category: CategorySecurity,
		Level:    LevelInfo,
		Message:  "Successfully hashed password",
	})
	return string(hashedPassword), nil
}

func (hs *HorizonSecurity) VerifyPassword(hashedPassword, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		hs.log.Log(LogEntry{
			Category: CategorySecurity,
			Level:    LevelError,
			Message:  "Password verification failed",
			Fields:   []zap.Field{zap.Error(err)},
		})
		return false, eris.Wrap(err, "password verification failed")
	}
	hs.log.Log(LogEntry{
		Category: CategorySecurity,
		Level:    LevelInfo,
		Message:  "Password verified successfully",
	})
	return true, nil
}

func (hs *HorizonSecurity) SanitizeHTML(input string) string {
	p := bluemonday.UGCPolicy()
	clean := p.Sanitize(input)

	hs.log.Log(LogEntry{
		Category: CategorySecurity,
		Level:    LevelInfo,
		Message:  "Sanitized HTML content",
	})
	return clean
}

func (hs *HorizonSecurity) Hash(value string) []byte {
	hash := sha256.New()
	hash.Write([]byte(value))
	return hash.Sum(nil)
}

func (hs *HorizonSecurity) Encrypt(value string) (string, error) {
	key := hs.Hash(hs.config.AppToken)
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		hs.log.Log(LogEntry{
			Category: CategorySecurity,
			Level:    LevelError,
			Message:  "Invalid encryption key length",
			Fields:   []zap.Field{zap.Int("key_length", len(key))},
		})
		return "", eris.New("key must be 16, 24, or 32 bytes after hashing")
	}

	plaintext := []byte(value)

	block, err := aes.NewCipher(key)
	if err != nil {
		hs.log.Log(LogEntry{
			Category: CategorySecurity,
			Level:    LevelError,
			Message:  "Failed to create AES cipher block",
			Fields:   []zap.Field{zap.Error(err)},
		})
		return "", eris.Wrap(err, "failed to create AES cipher block")
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		hs.log.Log(LogEntry{
			Category: CategorySecurity,
			Level:    LevelError,
			Message:  "Failed to create GCM block cipher",
			Fields:   []zap.Field{zap.Error(err)},
		})
		return "", eris.Wrap(err, "failed to create GCM block cipher")
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		hs.log.Log(LogEntry{
			Category: CategorySecurity,
			Level:    LevelError,
			Message:  "Failed to generate nonce",
			Fields:   []zap.Field{zap.Error(err)},
		})
		return "", eris.Wrap(err, "failed to generate nonce")
	}

	ciphertext := aesgcm.Seal(nonce, nonce, plaintext, nil)
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	hs.log.Log(LogEntry{
		Category: CategorySecurity,
		Level:    LevelInfo,
		Message:  "Successfully encrypted value",
	})

	return encoded, nil
}

func (hs *HorizonSecurity) Decrypt(value string) (string, error) {
	key := hs.Hash(hs.config.AppToken)
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		hs.log.Log(LogEntry{
			Category: CategorySecurity,
			Level:    LevelError,
			Message:  "Invalid decryption key length",
			Fields:   []zap.Field{zap.Int("key_length", len(key))},
		})
		return "", eris.New("key must be 16, 24, or 32 bytes after hashing")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		hs.log.Log(LogEntry{
			Category: CategorySecurity,
			Level:    LevelError,
			Message:  "Failed to decode ciphertext from base64",
			Fields:   []zap.Field{zap.Error(err)},
		})
		return "", eris.Wrap(err, "failed to decode ciphertext from base64")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		hs.log.Log(LogEntry{
			Category: CategorySecurity,
			Level:    LevelError,
			Message:  "Failed to create AES cipher block",
			Fields:   []zap.Field{zap.Error(err)},
		})
		return "", eris.Wrap(err, "failed to create AES cipher block")
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		hs.log.Log(LogEntry{
			Category: CategorySecurity,
			Level:    LevelError,
			Message:  "Failed to create GCM block cipher",
			Fields:   []zap.Field{zap.Error(err)},
		})
		return "", eris.Wrap(err, "failed to create GCM block cipher")
	}

	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		hs.log.Log(LogEntry{
			Category: CategorySecurity,
			Level:    LevelError,
			Message:  "Ciphertext too short",
		})
		return "", eris.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		hs.log.Log(LogEntry{
			Category: CategorySecurity,
			Level:    LevelError,
			Message:  "Failed to decrypt ciphertext",
			Fields:   []zap.Field{zap.Error(err)},
		})
		return "", eris.Wrap(err, "failed to decrypt ciphertext")
	}

	hs.log.Log(LogEntry{
		Category: CategorySecurity,
		Level:    LevelInfo,
		Message:  "Successfully decrypted value",
	})

	return string(plaintext), nil
}
