package horizon

import (
	"errors"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
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

func IsInvalidArgumentError(err error) bool {
	if runtime.GOOS == "windows" {
		var errno syscall.Errno
		if errors.As(err, &errno) {
			return errno == syscall.EINVAL
		}
	}
	return false
}

func GetFirstNChars(str string, n int) string {
	if len(str) == 0 {
		return ""
	}
	runes := []rune(str)
	if n > len(runes) {
		n = len(runes)
	}
	return string(runes[:n])
}

// Returns true if both are nil, or both are non-nil and equal.
func DateEqual(t1, t2 *time.Time) bool {
	switch {
	case t1 == nil && t2 == nil:
		return true
	case t1 == nil || t2 == nil:
		return false
	default:
		return t1.Equal(*t2)
	}
}

func Dateformat(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format(time.RFC3339)
	return &formatted
}

func StringFormat(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
