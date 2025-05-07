package horizon

import (
	"os"
	"path/filepath"

	"github.com/rotisserie/eris"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Category string

var Categories = []Category{
	CategoryAuthentication, CategoryBroadcast, CategoryCache,
	CategoryDatabase, CategoryOTP, CategoryQR,
	CategoryRequest, CategorySchedule, CategorySecurity,
	CategorySMS, CategorySMTP, CategoryStorage,
	CategoryTerminal, CategoryHijack,
}

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

	// unkown request
	CategoryHijack Category = "hijack"
)

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

type LogLevel int

const (
	LevelInfo LogLevel = iota
	LevelDebug
	LevelWarn
	LevelError
	LevelPanic
	LevelFatal
)

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

type LogEntry struct {
	Category Category    // which logger category to use
	Level    LogLevel    // severity of the log
	Message  string      // message to log
	Fields   []zap.Field // optional structured data
}

type HorizonLog struct {
	loggers     map[Category]*zap.Logger
	fallback    *zap.Logger
	config      *HorizonConfig
	mainRotator zapcore.WriteSyncer
}

func NewHorizonLog(config *HorizonConfig) (*HorizonLog, error) {
	mainLogPath := filepath.Join(config.AppLog, "main.log")
	mainRotator := &lumberjack.Logger{
		Filename:   mainLogPath,
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     28,
		Compress:   true,
	}

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
		loggers:     make(map[Category]*zap.Logger),
		fallback:    fallback,
		config:      config,
		mainRotator: zapcore.AddSync(mainRotator),
	}, nil
}

func (hl *HorizonLog) Run() error {
	loggers, err := hl.setupCategories(hl.config.AppLog, Categories)
	if err != nil {
		return eris.Wrap(err, "failed to initialize loggers")
	}

	hl.loggers = loggers
	hl.loggers["default"] = hl.fallback
	return nil
}
func (hl *HorizonLog) Stop() error {
	for category, logger := range hl.loggers {
		if err := logger.Sync(); err != nil {
			if !IsInvalidArgumentError(err) {
				hl.fallback.Warn("failed to sync logger", zap.String("category", string(category)), zap.Error(err))
			}
		}
	}
	return nil
}

func (hl *HorizonLog) Log(entry LogEntry) {
	if !hl.config.CanDebug() {
		return
	}
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

func (hl *HorizonLog) ClearAll() error {
	for category := range hl.loggers {
		if err := hl.ClearCategory(category); err != nil {
			return eris.Wrapf(err, "failed to clear category %q", category)
		}
	}
	return nil
}

func (hl *HorizonLog) ClearCategory(category Category) error {
	logPath := filepath.Join(hl.config.AppLog, string(category)+".log")
	err := os.Remove(logPath)
	if err != nil && !os.IsNotExist(err) {
		return eris.Wrap(err, "failed to remove log file")
	}
	return nil
}

func (hl *HorizonLog) Close() {
	for _, logger := range hl.loggers {
		_ = logger.Sync()
	}
}

func (hl *HorizonLog) setupCategories(logDir string, cats []Category) (map[Category]*zap.Logger, error) {
	loggers := make(map[Category]*zap.Logger, len(cats))
	for _, category := range cats {
		logger, err := hl.newCategoryLogger(logDir, category)
		if err != nil {
			return nil, eris.Wrapf(err, "failed to create logger for category %q", category)
		}
		loggers[category] = logger
	}
	return loggers, nil
}

func (hl *HorizonLog) newCategoryLogger(logDir string, category Category) (*zap.Logger, error) {
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, eris.Wrap(err, "failed to create log directory")
	}

	filePath := filepath.Join(logDir, string(category)+".log")
	categoryRotator := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     28,
		Compress:   true,
	}

	categoryWriter := zapcore.AddSync(categoryRotator)
	consoleWriter := zapcore.AddSync(os.Stdout)

	// Include main log writer (shared)
	mainWriter := hl.mainRotator

	level := zapcore.InfoLevel
	if lvlStr, ok := LogLevels[category]; ok {
		if parsed, err := zapcore.ParseLevel(lvlStr); err == nil {
			level = parsed
		}
	}

	// JSON encoder for files
	jsonEncCfg := zap.NewProductionEncoderConfig()
	jsonEncCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	jsonEncoder := zapcore.NewJSONEncoder(jsonEncCfg)

	// Console encoder (pretty)
	consoleEncCfg := zap.NewProductionEncoderConfig()
	consoleEncCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := NewHorizonPrettyJSONEncoder(consoleEncCfg)

	// Create all cores
	categoryCore := zapcore.NewCore(jsonEncoder, categoryWriter, level)
	mainCore := zapcore.NewCore(jsonEncoder, mainWriter, level)
	consoleCore := zapcore.NewCore(consoleEncoder, consoleWriter, level)

	// Combine them
	core := zapcore.NewTee(categoryCore, mainCore, consoleCore)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).With(
		zap.String("app", hl.config.AppName),
		zap.String("category", string(category)),
	)

	return logger, nil
}
