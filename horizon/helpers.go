package horizon

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

func GetBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	val = strings.ToLower(val)
	trueVals := map[string]bool{
		"1": true, "true": true, "yes": true, "on": true,
	}
	falseVals := map[string]bool{
		"0": false, "false": false, "no": false, "off": false,
	}
	if b, ok := trueVals[val]; ok {
		return b
	}
	if b, ok := falseVals[val]; ok {
		return b
	}
	return defaultVal
}

func GetString(key string, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func GetInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intVal
}

func GetFloat(key string, defaultVal float64) float64 {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	floatVal, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return defaultVal
	}
	return floatVal
}

func GetRequestBody(c echo.Context) string {
	if c.Request().Method == http.MethodPost || c.Request().Method == http.MethodPut {
		bodyBytes, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return "Error reading body"
		}
		return string(bodyBytes)
	}
	return ""
}

func Hash(keyStr string) []byte {
	hash := sha256.New()
	hash.Write([]byte(keyStr))
	return hash.Sum(nil)
}

func Encrypt(keyStr, plaintextStr string) (string, error) {
	key := Hash(keyStr)
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return "", eris.New("key must be 16, 24, or 32 bytes after hashing")
	}

	plaintext := []byte(plaintextStr)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", eris.Wrap(err, "failed to create AES cipher block")
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", eris.Wrap(err, "failed to create GCM block cipher")
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", eris.Wrap(err, "failed to generate nonce")
	}

	ciphertext := aesgcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(keyStr, encryptedStr string) (string, error) {
	key := Hash(keyStr)
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return "", eris.New("key must be 16, 24, or 32 bytes after hashing")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedStr)
	if err != nil {
		return "", eris.Wrap(err, "failed to decode ciphertext from base64")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", eris.Wrap(err, "failed to create AES cipher block")
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", eris.Wrap(err, "failed to create GCM block cipher")
	}

	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", eris.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", eris.Wrap(err, "failed to decrypt ciphertext")
	}

	return string(plaintext), nil
}

func IsInvalidArgumentError(err error) bool {
	if runtime.GOOS == "windows" {
		var errno syscall.Errno
		if errors.As(err, &errno) {
			return errno == syscall.EINVAL
		}
	}
	return false
}
