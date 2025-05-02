package horizon

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type HorizonLog struct {
	loggers map[string]*zap.Logger
	config  *HorizonConfig
}

func NewHorizonLog(config *HorizonConfig) (*HorizonLog, error) {
	return &HorizonLog{
		loggers: make(map[string]*zap.Logger),
		config:  config,
	}, nil
}

func (hl *HorizonLog) Run() error {
	categories := []string{
		"authentication", "broadcast", "cache", "database", "otp",
		"qr", "request", "schedule", "security", "sms", "smtp",
		"storage", "terminal",
	}
	loggers, err := hl.categories(hl.config.AppLog, categories)
	if err != nil {
		return err
	}
	hl.loggers = loggers
	return nil
}

func (hl *HorizonLog) LogAuthentication(message, level string, fields ...zap.Field) {
	hl.logAsync("authentication", level, message, fields...)
}

func (hl *HorizonLog) LogBroadcast(message, level string, fields ...zap.Field) {
	hl.logAsync("broadcast", level, message, fields...)
}

func (hl *HorizonLog) LogCache(message, level string, fields ...zap.Field) {
	hl.logAsync("cache", level, message, fields...)
}

func (hl *HorizonLog) LogDatabase(message, level string, fields ...zap.Field) {
	hl.logAsync("database", level, message, fields...)
}

func (hl *HorizonLog) LogOTP(message, level string, fields ...zap.Field) {
	hl.logAsync("otp", level, message, fields...)
}

func (hl *HorizonLog) LogQR(message, level string, fields ...zap.Field) {
	hl.logAsync("qr", level, message, fields...)
}

func (hl *HorizonLog) LogRequest(message, level string, fields ...zap.Field) {
	hl.logAsync("request", level, message, fields...)
}

func (hl *HorizonLog) LogSchedule(message, level string, fields ...zap.Field) {
	hl.logAsync("schedule", level, message, fields...)
}

func (hl *HorizonLog) LogSecurity(message, level string, fields ...zap.Field) {
	hl.logAsync("security", level, message, fields...)
}

func (hl *HorizonLog) LogSMS(message, level string, fields ...zap.Field) {
	hl.logAsync("sms", level, message, fields...)
}

func (hl *HorizonLog) LogSMTP(message, level string, fields ...zap.Field) {
	hl.logAsync("smtp", level, message, fields...)
}

func (hl *HorizonLog) LogStorage(message, level string, fields ...zap.Field) {
	hl.logAsync("storage", level, message, fields...)
}

func (hl *HorizonLog) LogTerminal(message, level string, fields ...zap.Field) {
	hl.logAsync("terminal", level, message, fields...)
}

func (hl *HorizonLog) logAsync(category, level string, message string, fields ...zap.Field) {
	go func() {
		if logger, ok := hl.loggers[category]; ok {
			switch level {
			case "info":
				logger.Info(message, fields...)
			case "error":
				logger.Error(message, fields...)
			case "warn":
				logger.Warn(message, fields...)
			case "debug":
				logger.Debug(message, fields...)
			default:
				logger.Info(message, fields...)
			}
		} else {
			fmt.Printf("No logger found for category: %s\n", category)
		}
	}()
}

func (hl *HorizonLog) categories(logDir string, cats []string) (map[string]*zap.Logger, error) {
	loggers := make(map[string]*zap.Logger, len(cats))
	for _, category := range cats {
		logger, err := hl.newCategoryLogger(logDir, category)
		if err != nil {
			return nil, fmt.Errorf("failed to create logger for category %q: %w", category, err)
		}
		loggers[category] = logger
	}
	return loggers, nil
}

func (hl *HorizonLog) newCategoryLogger(logDir, category string) (*zap.Logger, error) {
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, err
	}

	filePath := filepath.Join(logDir, category+".log")
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	writeSyncer := zapcore.AddSync(file)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writeSyncer,
		zapcore.InfoLevel, // Change this based on your desired default level
	)

	logger := zap.New(core)
	return logger, nil
}
