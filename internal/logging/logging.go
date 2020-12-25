package logging

import (
	"os"
	"sync"
	"time"
)

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// newDefaultProductionLog configures a custom log that is
// intended for use by default if no other log is specified
// in a config. It writes to stderr, uses the console encoder,
// and enables INFO-level logs and higher.
func newDefaultProductionLog() *zap.Logger {
	encCfg := zap.NewProductionEncoderConfig()
	// if interactive terminal, make output more human-readable by default
	encCfg.EncodeTime = func(ts time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(ts.Format("2006/01/02 15:04:05.000"))
	}
	encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	enc := zapcore.NewConsoleEncoder(encCfg)
	core := zapcore.NewCore(enc, zapcore.Lock(os.Stdout), zap.NewAtomicLevelAt(zap.DebugLevel))
	return zap.New(core)

}

// Log returns the current default logger.
func Log() *zap.Logger {
	defaultLoggerMu.RLock()
	defer defaultLoggerMu.RUnlock()
	return defaultLogger
}

var (
	defaultLogger   = newDefaultProductionLog()
	defaultLoggerMu sync.RWMutex
)
