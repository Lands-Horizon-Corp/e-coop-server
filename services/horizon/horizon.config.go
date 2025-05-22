package horizon

import (
	"log"
	"time"

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
type HorizonEnvironmentService struct{}

func NewEnvironmentService(path string) EnvironmentService {
	err := godotenv.Load(path)
	if err != nil {
		log.Printf("Warning: .env file not loaded from path: %s, err: %v", path, err)
	}
	viper.AutomaticEnv()
	return HorizonEnvironmentService{}
}

func (h HorizonEnvironmentService) GetInt16(key string, defaultValue int16) int16 {
	viper.SetDefault(key, defaultValue)
	return int16(viper.GetInt(key))
}

func (h HorizonEnvironmentService) GetByteSlice(key string, defaultValue string) []byte {
	viper.SetDefault(key, defaultValue)
	value := h.GetString(key, defaultValue)
	return []byte(value)
}

// Get implements EnvironmentService.
func (h HorizonEnvironmentService) Get(key string, defaultValue any) any {
	viper.SetDefault(key, defaultValue)
	return viper.Get(key)
}

// GetBool implements EnvironmentService.
func (h HorizonEnvironmentService) GetBool(key string, defaultValue bool) bool {
	viper.SetDefault(key, defaultValue)
	return viper.GetBool(key)
}

// GetDuration implements EnvironmentService.
func (h HorizonEnvironmentService) GetDuration(key string, defaultValue time.Duration) time.Duration {
	viper.SetDefault(key, defaultValue)
	return viper.GetDuration(key)
}

// GetFloat64 implements EnvironmentService.
func (h HorizonEnvironmentService) GetFloat64(key string, defaultValue float64) float64 {
	viper.SetDefault(key, defaultValue)
	return viper.GetFloat64(key)
}

// GetInt implements EnvironmentService.
func (h HorizonEnvironmentService) GetInt(key string, defaultValue int) int {
	viper.SetDefault(key, defaultValue)
	return viper.GetInt(key)
}

// GetInt32 implements EnvironmentService.
func (h HorizonEnvironmentService) GetInt32(key string, defaultValue int32) int32 {
	viper.SetDefault(key, defaultValue)
	return viper.GetInt32(key)
}

// GetInt64 implements EnvironmentService.
func (h HorizonEnvironmentService) GetInt64(key string, defaultValue int64) int64 {
	viper.SetDefault(key, defaultValue)
	return viper.GetInt64(key)
}

// GetIntSlice implements EnvironmentService.
func (h HorizonEnvironmentService) GetIntSlice(key string, defaultValue []int) []int {
	viper.SetDefault(key, defaultValue)
	return viper.GetIntSlice(key)
}

// GetSizeInBytes implements EnvironmentService.
func (h HorizonEnvironmentService) GetSizeInBytes(key string, defaultValue uint) uint {
	viper.SetDefault(key, defaultValue)
	return viper.GetUint(key)
}

// GetString implements EnvironmentService.
func (h HorizonEnvironmentService) GetString(key string, defaultValue string) string {
	viper.SetDefault(key, defaultValue)
	return viper.GetString(key)
}

// GetStringMap implements EnvironmentService.
func (h HorizonEnvironmentService) GetStringMap(key string, defaultValue map[string]any) map[string]any {
	viper.SetDefault(key, defaultValue)
	return viper.GetStringMap(key)
}

// GetStringMapString implements EnvironmentService.
func (h HorizonEnvironmentService) GetStringMapString(key string, defaultValue map[string]string) map[string]string {
	viper.SetDefault(key, defaultValue)
	return viper.GetStringMapString(key)
}

// GetStringMapStringSlice implements EnvironmentService.
func (h HorizonEnvironmentService) GetStringMapStringSlice(key string, defaultValue map[string][]string) map[string][]string {
	viper.SetDefault(key, defaultValue)
	return viper.GetStringMapStringSlice(key)
}

// GetStringSlice implements EnvironmentService.
func (h HorizonEnvironmentService) GetStringSlice(key string, defaultValue []string) []string {
	viper.SetDefault(key, defaultValue)
	return viper.GetStringSlice(key)
}

// GetTime implements EnvironmentService.
func (h HorizonEnvironmentService) GetTime(key string, defaultValue time.Time) time.Time {
	viper.SetDefault(key, defaultValue)
	return viper.GetTime(key)
}

// GetUint implements EnvironmentService.
func (h HorizonEnvironmentService) GetUint(key string, defaultValue uint) uint {
	viper.SetDefault(key, defaultValue)
	return viper.GetSizeInBytes(key)
}

// GetUint16 implements EnvironmentService.
func (h HorizonEnvironmentService) GetUint16(key string, defaultValue uint16) uint16 {
	viper.SetDefault(key, defaultValue)
	return viper.GetUint16(key)
}

// GetUint32 implements EnvironmentService.
func (h HorizonEnvironmentService) GetUint32(key string, defaultValue uint32) uint32 {
	viper.SetDefault(key, defaultValue)
	return viper.GetUint32(key)
}

// GetUint64 implements EnvironmentService.
func (h HorizonEnvironmentService) GetUint64(key string, defaultValue uint64) uint64 {
	viper.SetDefault(key, defaultValue)
	return viper.GetUint64(key)
}

// GetUint8 implements EnvironmentService.
func (h HorizonEnvironmentService) GetUint8(key string, defaultValue uint8) uint8 {
	viper.SetDefault(key, defaultValue)
	return viper.GetUint8(key)
}
