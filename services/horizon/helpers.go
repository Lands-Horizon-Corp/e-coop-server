package horizon

import (
	"errors"
	"math/rand"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

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

// IsValidEmail checks if the provided string is a valid email address format
func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}
	_, err := mail.ParseAddress(email)
	return err == nil
}

func IsValidPhoneNumber(phoneNumber string) bool {
	re := regexp.MustCompile(`^\+?(?:\d{1,4})?\d{7,14}$`)
	return re.MatchString(phoneNumber)
}

func GenerateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func Create32ByteKey(key []byte) string {
	if len(key) > 32 {
		return string(key[:32])
	}
	padded := make([]byte, 32)
	copy(padded, key)
	return string(padded)
}

func IsValidURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if u.Host == "" {
		return false
	}
	if strings.ContainsAny(rawURL, " <>\"") {
		return false
	}

	return true
}

func GenerateRandomDigits(size int) (int, error) {
	if size > 8 {
		return 0, errors.New("size must not exceed 8 digits")
	}
	if size <= 0 {
		return 0, errors.New("size must be a positive integer")
	}

	min := intPow(10, size-1)
	max := intPow(10, size) - 1

	// Use a local source (deterministic if seeded)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(max-min+1) + min, nil
}

func intPow(a, b int) int {
	result := 1
	for i := 0; i < b; i++ {
		result *= a
	}
	return result
}

func Capitalize(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

func MergeString(defaults, overrides []string) []string {
	totalCap := len(defaults) + len(overrides)
	seen := make(map[string]struct{}, totalCap)
	out := make([]string, 0, totalCap)
	for _, slice := range [][]string{defaults, overrides} {
		for _, p := range slice {
			cp := Capitalize(p)
			if cp == "" {
				continue
			}
			if _, exists := seen[cp]; !exists {
				seen[cp] = struct{}{}
				out = append(out, cp)
			}
		}
	}
	return out
}

func EngineUUIDParam(ctx echo.Context, idParam string) (*uuid.UUID, error) {
	param := ctx.Param(idParam)
	id, err := uuid.Parse(param)
	if err != nil {
		return nil, ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid feedback ID"})
	}
	return &id, nil
}

func ParseUUID(s *string) uuid.UUID {
	if s == nil || strings.TrimSpace(*s) == "" {
		return uuid.Nil
	}
	if id, err := uuid.Parse(*s); err == nil {
		return id
	}
	return uuid.Nil
}
