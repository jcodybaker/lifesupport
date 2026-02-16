package temporallog

import (
	"github.com/rs/zerolog"
	temporalLog "go.temporal.io/sdk/log"
)

// zerologAdapter wraps zerolog.Logger to implement Temporal's Logger interface
type zerologAdapter struct {
	logger zerolog.Logger
}

func (z *zerologAdapter) Debug(msg string, keyvals ...interface{}) {
	z.log(z.logger.Debug(), msg, keyvals...)
}

func (z *zerologAdapter) Info(msg string, keyvals ...interface{}) {
	z.log(z.logger.Info(), msg, keyvals...)
}

func (z *zerologAdapter) Warn(msg string, keyvals ...interface{}) {
	z.log(z.logger.Warn(), msg, keyvals...)
}

func (z *zerologAdapter) Error(msg string, keyvals ...interface{}) {
	z.log(z.logger.Error(), msg, keyvals...)
}

func (z *zerologAdapter) log(event *zerolog.Event, msg string, keyvals ...interface{}) {
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			key := keyvals[i].(string)
			val := keyvals[i+1]
			event = event.Interface(key, val)
		}
	}
	event.Msg(msg)
}

// NewTemporalLogger creates a Temporal logger that wraps a zerolog.Logger
func NewTemporalLogger(logger zerolog.Logger) temporalLog.Logger {
	return &zerologAdapter{logger: logger}
}
