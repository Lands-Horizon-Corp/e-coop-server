package horizon

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

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
	if intVal, err := strconv.Atoi(val); err == nil {
		return intVal
	}
	return defaultVal
}

func GetFloat(key string, defaultVal float64) float64 {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
		return floatVal
	}
	return defaultVal
}

func GetRequestBody(c echo.Context) string {
	body := ""
	if c.Request().Method == http.MethodPost || c.Request().Method == http.MethodPut {
		var err error
		bodyBytes, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return "Error reading body"
		}
		body = string(bodyBytes)
	}
	return body
}
