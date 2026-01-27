package log

import (
	"strings"

	"github.com/nunoOliveiraqwe/micro-proxy/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(logFlags *Flags) {
	var cfg zap.Config

	if logFlags.Debug {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()

	}

	internalLogLevel := getLogLevelFromLogStr(logFlags.LogLevel)

	zapLevel := internalLogLevel.toZapLevel()

	cfg.Level = zap.NewAtomicLevelAt(zapLevel)

	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stderr"}

	if logFlags.LogPath != "" {
		exists := util.FileExists(logFlags.LogPath)
		if exists {
			cfg.OutputPaths = []string{logFlags.LogPath, "stdout"}
			cfg.ErrorOutputPaths = []string{logFlags.LogPath, "stderr"}
		}
	}

	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	log, err := cfg.Build()
	if err != nil {
		log = zap.NewNop()
		zap.S().Errorf("Failed to initialize logger: %v", err)
	}
	zap.RedirectStdLog(log)
	zap.ReplaceGlobals(log)
}

func getLogLevelFromLogStr(logLevelStr string) LogLevel {
	all := allLevels()
	for _, level := range all {
		if strings.EqualFold(strings.ToLower(level.String()), strings.ToLower(logLevelStr)) {
			return level
		}
	}
	return defaultLogLevel
}
