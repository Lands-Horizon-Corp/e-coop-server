package horizon

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Category represents a logging category.
type Category string

const (
	CategoryAuthentication Category = "authentication"
	CategoryBroadcast      Category = "broadcast"
	CategoryCache          Category = "cache"
	CategoryDatabase       Category = "database"
	CategoryOTP            Category = "otp"
	CategoryQR             Category = "qr"
	CategoryRequest        Category = "request"
	CategorySchedule       Category = "schedule"
	CategorySecurity       Category = "security"
	CategorySMS            Category = "sms"
	CategorySMTP           Category = "smtp"
	CategoryStorage        Category = "storage"
	CategoryTerminal       Category = "terminal"
)

// LogLevels defines default per-category log levels (used if none provided in config)
var LogLevels = map[Category]string{
	CategoryAuthentication: "info",
	CategoryBroadcast:      "info",
	CategoryCache:          "info",
	CategoryDatabase:       "info",
	CategoryOTP:            "info",
	CategoryQR:             "info",
	CategoryRequest:        "info",
	CategorySchedule:       "info",
	CategorySecurity:       "info",
	CategorySMS:            "info",
	CategorySMTP:           "info",
	CategoryStorage:        "info",
	CategoryTerminal:       "info",
}

// LogLevel represents the severity level for logging.
type LogLevel int

const (
	LevelInfo LogLevel = iota
	LevelDebug
	LevelWarn
	LevelError
	LevelPanic
	LevelFatal
)

// String returns the string representation of the LogLevel.
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelPanic:
		return "panic"
	case LevelFatal:
		return "fatal"
	default:
		return "info"
	}
}

// LogEntry represents a single log invocation.
type LogEntry struct {
	Category Category    // which logger category to use
	Level    LogLevel    // severity of the log
	Message  string      // message to log
	Fields   []zap.Field // optional structured data
}

// HorizonLog manages loggers per category.
type HorizonLog struct {
	loggers  map[Category]*zap.Logger
	fallback *zap.Logger
	config   *HorizonConfig
}

// NewHorizonLog initializes HorizonLog with the given config.
// If config.LogLevels is nil, the package-level LogLevels will be used.
func NewHorizonLog(config *HorizonConfig) (*HorizonLog, error) {

	// Fallback logger includes caller info
	fallbackEncoderCfg := zap.NewProductionEncoderConfig()
	fallbackEncoderCfg.TimeKey = "timestamp"
	fallbackEncoderCfg.CallerKey = "caller"
	fallbackEncoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	fallbackEncoderCfg.EncodeCaller = zapcore.ShortCallerEncoder

	fallbackCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(fallbackEncoderCfg),
		zapcore.AddSync(os.Stdout),
		zapcore.ErrorLevel,
	)
	fallback := zap.New(fallbackCore, zap.AddCaller(), zap.AddCallerSkip(1)).With(
		zap.String("app", config.AppName),
	)

	return &HorizonLog{
		loggers:  make(map[Category]*zap.Logger),
		fallback: fallback,
		config:   config,
	}, nil
}

// Run sets up loggers for each category in the config.
func (hl *HorizonLog) Run() error {
	categories := []Category{
		CategoryAuthentication, CategoryBroadcast, CategoryCache,
		CategoryDatabase, CategoryOTP, CategoryQR,
		CategoryRequest, CategorySchedule, CategorySecurity,
		CategorySMS, CategorySMTP, CategoryStorage,
		CategoryTerminal,
	}
	loggers, err := hl.setupCategories(hl.config.AppLog, categories)
	if err != nil {
		return fmt.Errorf("failed to initialize loggers: %w", err)
	}

	hl.loggers = loggers
	hl.loggers["default"] = hl.fallback
	return nil
}

// setupCategories creates a zap.Logger for each category using lumberjack for rotation.
func (hl *HorizonLog) setupCategories(logDir string, cats []Category) (map[Category]*zap.Logger, error) {
	loggers := make(map[Category]*zap.Logger, len(cats))
	for _, category := range cats {
		logger, err := hl.newCategoryLogger(logDir, category)
		if err != nil {
			return nil, fmt.Errorf("failed to create logger for category %q: %w", category, err)
		}
		loggers[category] = logger
	}
	return loggers, nil
}

// newCategoryLogger builds a logger for a single category with rotation, console output, and caller info.
func (hl *HorizonLog) newCategoryLogger(logDir string, category Category) (*zap.Logger, error) {
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, err
	}

	filePath := filepath.Join(logDir, string(category)+".log")
	rotator := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     28,
		Compress:   true,
	}

	writeSyncer := zapcore.AddSync(rotator)
	consoleSyncer := zapcore.AddSync(os.Stdout)

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "timestamp"
	encCfg.CallerKey = "caller"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encCfg.EncodeCaller = zapcore.ShortCallerEncoder

	// Determine per-category log level
	level := zapcore.InfoLevel
	if lvlStr, ok := LogLevels[category]; ok {
		if parsed, err := zapcore.ParseLevel(lvlStr); err == nil {
			level = parsed
		}
	}

	jsonCore := zapcore.NewCore(zapcore.NewJSONEncoder(encCfg), writeSyncer, level)
	consoleCore := zapcore.NewCore(zapcore.NewConsoleEncoder(encCfg), consoleSyncer, level)
	core := zapcore.NewTee(jsonCore, consoleCore)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).With(
		zap.String("app", hl.config.AppName),
		zap.String("category", string(category)),
	)
	return logger, nil
}

// Log sends a LogEntry to the appropriate category at its level.
func (hl *HorizonLog) Log(entry LogEntry) {
	logger, ok := hl.loggers[entry.Category]
	if !ok {
		logger = hl.loggers["default"]
	}
	switch entry.Level {
	case LevelDebug:
		logger.Debug(entry.Message, entry.Fields...)
	case LevelWarn:
		logger.Warn(entry.Message, entry.Fields...)
	case LevelError:
		logger.Error(entry.Message, entry.Fields...)
	case LevelPanic:
		logger.Panic(entry.Message, entry.Fields...)
	case LevelFatal:
		logger.Fatal(entry.Message, entry.Fields...)
	default:
		logger.Info(entry.Message, entry.Fields...)
	}
}

// Close gracefully flushes all loggers.
func (hl *HorizonLog) Close() {
	for _, logger := range hl.loggers {
		_ = logger.Sync()
	}
}
