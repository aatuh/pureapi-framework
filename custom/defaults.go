package custom

import (
	"context"
	"time"

	databasetypes "github.com/pureapi/pureapi-core/database/types"
	"github.com/pureapi/pureapi-core/util"
	utiltypes "github.com/pureapi/pureapi-core/util/types"
	"github.com/pureapi/pureapi-framework/middleware"
	repositorytypes "github.com/pureapi/pureapi-framework/repository/types"
	"github.com/pureapi/pureapi-sqlite/errorchecker"
	"github.com/pureapi/pureapi-sqlite/query"
	"github.com/pureapi/pureapi-util/envvar"
	"github.com/pureapi/pureapi-util/logging"
)

const (
	loggingCompactEnvVar   = "LOGGING_COMPACT"
	loggingAnsiCodesEnvVar = "LOGGING_ANSI_CODES"
	loggingLevelEnvVar     = "LOGGING_LEVEL"
)

type LoggingConfig struct {
	LoggingLevel logging.LogLevel
	Compact      bool
	AnsiCodes    bool
}

func NewLoggingConfig() *LoggingConfig {
	loggingLevel, err := logging.LoggingLevelStrToInt(envvar.MustGet(
		loggingLevelEnvVar,
	))
	if err != nil {
		panic(err)
	}
	return &LoggingConfig{
		LoggingLevel: loggingLevel,
		Compact:      envvar.MustGetBool(loggingCompactEnvVar),
		AnsiCodes:    envvar.MustGetBool(loggingAnsiCodesEnvVar),
	}
}

// WrapCtxLoggerFactory wraps a context-based logger factory into a
// LoggerFactoryFn.
func WrapCtxLoggerFactory(fn utiltypes.CtxLoggerFactoryFn) utiltypes.LoggerFactoryFn {
	return func(params ...any) utiltypes.ILogger {
		var ctx context.Context
		if len(params) > 0 {
			if c, ok := params[0].(context.Context); ok {
				ctx = c
			}
		}
		if ctx == nil {
			ctx = context.Background()
		}
		return fn(ctx)
	}
}

func CtxLogger(ctx context.Context) utiltypes.ILogger {
	opts := defaultLogOpts()
	opts.GetExtraData = func(ctx context.Context) *logging.ExtraData {
		return extraData(ctx)
	}
	return logging.NewContextLogger(ctx, opts)
}

func LoggerPlain(startTime time.Time, traceID string) utiltypes.ILogger {
	opts := defaultLogOpts()
	opts.GetExtraData = func(ctx context.Context) *logging.ExtraData {
		return nonCtxExtraData(startTime, traceID)
	}
	return logging.NewContextLogger(context.Background(), opts)
}

func EmitterLogger(eventEmitter ...utiltypes.EventEmitter) utiltypes.EmitterLogger {
	var useEventEmitter utiltypes.EventEmitter
	if len(eventEmitter) > 0 && eventEmitter[0] != nil {
		useEventEmitter = eventEmitter[0]
	} else {
		useEventEmitter = util.NewEventEmitter()
	}
	return util.NewEmitterLogger(
		useEventEmitter,
		WrapCtxLoggerFactory(CtxLogger),
	)
}

func QueryBuilder() repositorytypes.QueryBuilder {
	return &query.Query{}
}

func QueryErrorChecker() databasetypes.ErrorChecker {
	return errorchecker.NewErrorChecker()
}

func extraData(ctx context.Context) *logging.ExtraData {
	requestMetadata := middleware.GetRequestMetadata(ctx)
	if requestMetadata == nil {
		return nil
	}
	return nonCtxExtraData(
		requestMetadata.TimeStart,
		requestMetadata.TraceID,
	)
}

func nonCtxExtraData(
	timeStart time.Time,
	traceID string,
) *logging.ExtraData {
	now := time.Now().UTC()
	return &logging.ExtraData{
		Time:      &now,
		TimeStart: &timeStart,
		TimeDelta: now.Sub(timeStart).String(),
		TraceID:   traceID,
	}
}

func defaultLogOpts() *logging.LogOpts {
	config := NewLoggingConfig()
	return &logging.LogOpts{
		LoggingLevel: config.LoggingLevel,
		Compact:      config.Compact,
		AnsiCodes:    config.AnsiCodes,
		LogLevelOpts: logging.DefaultLogLevelOpts(),
	}
}
