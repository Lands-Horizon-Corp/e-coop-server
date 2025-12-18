package horizon

import (
	"log"
	"math"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type EnvironmentService interface {
	Get(key string, defaultValue any) any
	GetString(key string, defaultValue string) string
	GetByteSlice(key string, defaultValue string) []byte
	GetBool(key string, defaultValue bool) bool
	GetInt(key string, defaultValue int) int
	GetInt16(key string, defaultValue int16) int16
	GetInt32(key string, defaultValue int32) int32
	GetInt64(key string, defaultValue int64) int64
	GetUint8(key string, defaultValue uint8) uint8
	GetUint(key string, defaultValue uint) uint
	GetUint16(key string, defaultValue uint16) uint16
	GetUint32(key string, defaultValue uint32) uint32
	GetUint64(key string, defaultValue uint64) uint64
	GetFloat64(key string, defaultValue float64) float64
	GetTime(key string, defaultValue time.Time) time.Time
	GetDuration(key string, defaultValue time.Duration) time.Duration
	GetIntSlice(key string, defaultValue []int) []int
	GetStringSlice(key string, defaultValue []string) []string
	GetStringMap(key string, defaultValue map[string]any) map[string]any
	GetStringMapString(key string, defaultValue map[string]string) map[string]string
	GetStringMapStringSlice(key string, defaultValue map[string][]string) map[string][]string
	GetSizeInBytes(key string, defaultValue uint) uint
}

type EnvironmentServiceImpl struct{}

func NewEnvironmentService(path string) EnvironmentService {
	if !handlers.FileExists(path) {
		log.Printf("Info: Provided .env path is empty or does not exist. Falling back to default: .env")
		path = ".env"
	}

	err := godotenv.Load(path)
	if err != nil {
		log.Printf("Warning: .env file not loaded from path: %s, err: %v", path, err)
	}

	viper.AutomaticEnv()
	return EnvironmentServiceImpl{}
}

func (h EnvironmentServiceImpl) GetInt16(key string, defaultValue int16) int16 {
	viper.SetDefault(key, defaultValue)
	val := viper.GetInt(key)
	if val < math.MinInt16 || val > math.MaxInt16 {
		return defaultValue
	}
	return int16(val)
}

func (h EnvironmentServiceImpl) GetByteSlice(key string, defaultValue string) []byte {
	viper.SetDefault(key, defaultValue)
	value := h.GetString(key, defaultValue)
	return []byte(value)
}

func (h EnvironmentServiceImpl) Get(key string, defaultValue any) any {
	viper.SetDefault(key, defaultValue)
	return viper.Get(key)
}

func (h EnvironmentServiceImpl) GetBool(key string, defaultValue bool) bool {
	viper.SetDefault(key, defaultValue)
	return viper.GetBool(key)
}

func (h EnvironmentServiceImpl) GetDuration(key string, defaultValue time.Duration) time.Duration {
	viper.SetDefault(key, defaultValue)
	return viper.GetDuration(key)
}

func (h EnvironmentServiceImpl) GetFloat64(key string, defaultValue float64) float64 {
	viper.SetDefault(key, defaultValue)
	return viper.GetFloat64(key)
}

func (h EnvironmentServiceImpl) GetInt(key string, defaultValue int) int {
	viper.SetDefault(key, defaultValue)
	return viper.GetInt(key)
}

func (h EnvironmentServiceImpl) GetInt32(key string, defaultValue int32) int32 {
	viper.SetDefault(key, defaultValue)
	return viper.GetInt32(key)
}

func (h EnvironmentServiceImpl) GetInt64(key string, defaultValue int64) int64 {
	viper.SetDefault(key, defaultValue)
	return viper.GetInt64(key)
}

func (h EnvironmentServiceImpl) GetIntSlice(key string, defaultValue []int) []int {
	viper.SetDefault(key, defaultValue)
	return viper.GetIntSlice(key)
}

func (h EnvironmentServiceImpl) GetSizeInBytes(key string, defaultValue uint) uint {
	viper.SetDefault(key, defaultValue)
	return viper.GetUint(key)
}

func (h EnvironmentServiceImpl) GetString(key string, defaultValue string) string {
	viper.SetDefault(key, defaultValue)
	return viper.GetString(key)
}

func (h EnvironmentServiceImpl) GetStringMap(key string, defaultValue map[string]any) map[string]any {
	viper.SetDefault(key, defaultValue)
	return viper.GetStringMap(key)
}

func (h EnvironmentServiceImpl) GetStringMapString(key string, defaultValue map[string]string) map[string]string {
	viper.SetDefault(key, defaultValue)
	return viper.GetStringMapString(key)
}

func (h EnvironmentServiceImpl) GetStringMapStringSlice(key string, defaultValue map[string][]string) map[string][]string {
	viper.SetDefault(key, defaultValue)
	return viper.GetStringMapStringSlice(key)
}

func (h EnvironmentServiceImpl) GetStringSlice(key string, defaultValue []string) []string {
	viper.SetDefault(key, defaultValue)
	return viper.GetStringSlice(key)
}

func (h EnvironmentServiceImpl) GetTime(key string, defaultValue time.Time) time.Time {
	viper.SetDefault(key, defaultValue)
	return viper.GetTime(key)
}

func (h EnvironmentServiceImpl) GetUint(key string, defaultValue uint) uint {
	viper.SetDefault(key, defaultValue)
	return viper.GetSizeInBytes(key)
}

func (h EnvironmentServiceImpl) GetUint16(key string, defaultValue uint16) uint16 {
	viper.SetDefault(key, defaultValue)
	return viper.GetUint16(key)
}

func (h EnvironmentServiceImpl) GetUint32(key string, defaultValue uint32) uint32 {
	viper.SetDefault(key, defaultValue)
	return viper.GetUint32(key)
}

func (h EnvironmentServiceImpl) GetUint64(key string, defaultValue uint64) uint64 {
	viper.SetDefault(key, defaultValue)
	return viper.GetUint64(key)
}

func (h EnvironmentServiceImpl) GetUint8(key string, defaultValue uint8) uint8 {
	viper.SetDefault(key, defaultValue)
	return viper.GetUint8(key)
}
