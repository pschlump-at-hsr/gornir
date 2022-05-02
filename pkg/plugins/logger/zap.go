package logger

import (
	"github.com/nornir-automation/gornir/pkg/gornir"
	"go.uber.org/zap"
)

// Zap uses go.uber.org/zap to log messages. Implements gornir.Logger Interface
type Zap struct {
	zapLogger     *zap.Logger
	sugaredLogger *zap.SugaredLogger
}

// NewZap instantiates a new Zap logger
func NewZap(loglevel string) (*Zap, error) {
	config := zap.NewProductionConfig()
	var err error
	config.Level, err = zap.ParseAtomicLevel(loglevel)
	if err != nil {
		return nil, err
	}

	zap, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &Zap{zapLogger: zap}, nil
}

// NewZapFromLogger instantiates a new Zap logger
func NewZapFromLogger(zapLogger *zap.Logger) *Zap {
	return &Zap{zapLogger: zapLogger}
}

// WithField implements gornir.Logger interface
func (z *Zap) WithField(field string, value interface{}) gornir.Logger {
	return &Zap{sugaredLogger: z.zapLogger.Sugar().With(field, value)}
}

// Info implements gornir.Logger interface
func (z *Zap) Info(args ...interface{}) {
	z.zapLogger.Sugar().Info(args)
}

// Debug implements gornir.Logger interface
func (z *Zap) Debug(args ...interface{}) {
	z.zapLogger.Sugar().Debug(args)
}

// Error implements gornir.Logger interface
func (z *Zap) Error(args ...interface{}) {
	z.zapLogger.Sugar().Error(args)
}

// Warn implements gornir.Logger interface
func (z *Zap) Warn(args ...interface{}) {
	z.zapLogger.Sugar().Warn(args)
}

// Fatal implements gornir.Logger interface
func (z *Zap) Fatal(args ...interface{}) {
	z.zapLogger.Sugar().Fatal(args)
}
