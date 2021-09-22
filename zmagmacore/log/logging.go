package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/0chain/gosdk/zmagmacore/errors"
)

var (
	// Logger represents main logger implementation used in app.
	Logger = zap.NewNop()

	// logName
	logName string
)

// InitLogging initializes the main Logger consistent with passed log directory and level.
//
// If an error occurs during execution, the program terminates with code 2 and the error will be written in os.Stderr.
//
// InitLogging should be used only once while application is starting.
func InitLogging(development bool, logDir, level string) {
	logName = logDir + "/" + "logs.log"
	var (
		logWriter = getWriteSyncer(logName)
		logCfg    zap.Config
	)

	if development {
		logCfg = zap.NewProductionConfig()
		logCfg.DisableCaller = true
	} else {
		logCfg = zap.NewDevelopmentConfig()
		logCfg.EncoderConfig.LevelKey = "level"
		logCfg.EncoderConfig.NameKey = "name"
		logCfg.EncoderConfig.MessageKey = "msg"
		logCfg.EncoderConfig.CallerKey = "caller"
		logCfg.EncoderConfig.StacktraceKey = "stacktrace"

		logWriter = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), logWriter)
	}
	_ = logCfg.Level.UnmarshalText([]byte(level))
	logCfg.Encoding = consoleEncoderType
	logCfg.EncoderConfig.TimeKey = "timestamp"
	logCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	l, err := logCfg.Build(setOutput(logWriter, logCfg))
	if err != nil {
		errors.ExitErr("error while build logger config", err, 2)
	}

	Logger = l
}

const (
	jsonEncoderType    = "json"
	consoleEncoderType = "console"
)

// setOutput replaces existing Core with new, that writes to passed zapcore.WriteSyncer.
func setOutput(ws zapcore.WriteSyncer, conf zap.Config) zap.Option {
	var enc zapcore.Encoder
	switch conf.Encoding {
	case jsonEncoderType:
		enc = zapcore.NewJSONEncoder(conf.EncoderConfig)
	case consoleEncoderType:
		enc = zapcore.NewConsoleEncoder(conf.EncoderConfig)
	default:
		errors.ExitMsg("error while build logger config", 2)
	}

	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewCore(enc, ws, conf.Level)
	})
}

// getWriteSyncer creates zapcore.WriteSyncer using provided log file.
func getWriteSyncer(logName string) zapcore.WriteSyncer {
	var ioWriter = &lumberjack.Logger{
		Filename:   logName,
		MaxSize:    10, // MB
		MaxBackups: 3,  // number of backups
		MaxAge:     28, // days
		LocalTime:  true,
		Compress:   false, // disabled by default
	}
	_ = ioWriter.Rotate()
	return zapcore.AddSync(ioWriter)
}
