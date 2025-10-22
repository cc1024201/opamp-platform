package opamp

import (
	"context"

	"github.com/open-telemetry/opamp-go/client/types"
	"go.uber.org/zap"
)

// loggerAdapter adapts zap.Logger to opamp-go's Logger interface
type loggerAdapter struct {
	logger *zap.Logger
}

func newLoggerAdapter(logger *zap.Logger) types.Logger {
	return &loggerAdapter{logger: logger}
}

func (l *loggerAdapter) Debugf(ctx context.Context, format string, v ...interface{}) {
	l.logger.Sugar().Debugf(format, v...)
}

func (l *loggerAdapter) Errorf(ctx context.Context, format string, v ...interface{}) {
	l.logger.Sugar().Errorf(format, v...)
}
