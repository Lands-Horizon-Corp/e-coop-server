package horizon

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"golang.org/x/crypto/argon2"
)

// SecurityUtils provides cryptographic and security-related functions
type SecurityService interface {
	// GenerateUUID creates a new UUIDv4
	GenerateUUID(ctx context.Context) (string, error)

	// HashPassword creates an Argon2 hashed password
	HashPassword(ctx context.Context, password string) (string, error)

	// VerifyPassword compares a password with its hash
	VerifyPassword(ctx context.Context, hash, password string) (bool, error)

	// Encrypt performs AES encryption on plaintext
	Encrypt(ctx context.Context, plaintext string) (string, error)

	// Decrypt performs AES decryption on ciphertext
	Decrypt(ctx context.Context, ciphertext string) (string, error)

	GenerateUUIDv5(ctx context.Context, name string) (string, error)
}

// HorizonSecurity is a concrete implementation of SecurityUtils
type HorizonSecurity struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
	secret      []byte
}

// NewSecurityService returns a new instance of SecurityUtils
func NewSecurityService(
	memory uint32,
	iterations uint32,
	parallelism uint8,
	saltLength uint32,
	keyLength uint32,
	secret []byte,
) SecurityService {
	return &HorizonSecurity{
		memory:      memory,
		iterations:  iterations,
		parallelism: parallelism,
		saltLength:  saltLength,
		keyLength:   keyLength,
		secret:      secret,
	}
}

// Decrypt implements SecurityUtils.
func (h *HorizonSecurity) Decrypt(ctx context.Context, ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher([]byte(Create32ByteKey(h.secret)))
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, encryptedData := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, []byte(encryptedData), nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// Encrypt implements SecurityUtils.
func (h *HorizonSecurity) Encrypt(ctx context.Context, plaintext string) (string, error) {
	block, err := aes.NewCipher([]byte(Create32ByteKey(h.secret)))
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// GenerateUUID implements SecurityUtils.
func (h *HorizonSecurity) GenerateUUID(ctx context.Context) (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// VerifyPassword implements SecurityUtils.
func (h *HorizonSecurity) HashPassword(ctx context.Context, password string) (string, error) {
	salt, err := GenerateRandomBytes(h.saltLength)
	if err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, h.iterations, h.memory, h.parallelism, h.keyLength)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, h.memory, h.iterations, h.parallelism, b64Salt, b64Hash)
	return encodedHash, nil
}

// VerifyPassword implements SecurityUtils.
func (h *HorizonSecurity) VerifyPassword(ctx context.Context, hash string, password string) (bool, error) {
	vals := strings.Split(hash, "$")
	if len(vals) != 6 {
		return false, eris.New("the encoded hash is not in the correct format")
	}

	var version int
	_, err := fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return false, err
	}

	if version != argon2.Version {
		return false, eris.New("incompatible version of argon2")
	}

	var p struct {
		memory      uint32
		iterations  uint32
		parallelism uint8
	}

	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return false, err
	}

	hashed, err := base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return false, err
	}

	otherHash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, h.keyLength)

	if subtle.ConstantTimeCompare(hashed, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

func (h *HorizonSecurity) GenerateUUIDv5(ctx context.Context, name string) (string, error) {
	namespace := uuid.NameSpaceX500
	if name == "" {
		return "", errors.New("name cannot be empty")
	}

	uuid5 := uuid.NewSHA1(namespace, []byte(name))
	return uuid5.String(), nil
}
