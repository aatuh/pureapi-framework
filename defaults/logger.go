package defaults

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/aatuh/envvar"
	"github.com/aatuh/pureapi-core/event"
	"github.com/aatuh/pureapi-core/logging"
	"github.com/aatuh/pureapi-framework/api/middleware"
)

// EnvVarConfig holds the names of the environment variables used.
type EnvVarConfig struct {
	LoggingCompact   string
	LoggingAnsiCodes string
	LoggingLevel     string
}

// LoggingConfig holds the logging configuration.
type LoggingConfig struct {
	LoggingLevel logging.LogLevel
	Compact      bool
	AnsiCodes    bool
}

var (
	// Variables holding the default logger factories.
	loggerFactory        atomic.Value
	loggerPlainFactory   atomic.Value
	emitterLoggerFactory atomic.Value

	// Environment variable config.
	envVarConfig atomic.Value
)

// init sets the default logger factories and env var config.
func init() {
	loggerFactory.Store(defaultCTXLogger)
	loggerPlainFactory.Store(defaultLoggerPlain)
	emitterLoggerFactory.Store(defaultEmitterLogger)

	envVarConfig.Store(EnvVarConfig{
		LoggingCompact:   "LOGGING_COMPACT",
		LoggingAnsiCodes: "LOGGING_ANSI_CODES",
		LoggingLevel:     "LOGGING_LEVEL",
	})
}

// SetEnvVarConfig allows overriding the default environment variable names.
//
// Parameters:
//   - cfg EnvVarConfig: The environment variable names.
func SetEnvVarConfig(cfg EnvVarConfig) {
	envVarConfig.Store(cfg)
}

// GetEnvVarConfig returns the current environment variable configuration.
//
// Returns:
//   - EnvVarConfig: The environment variable configuration.
func GetEnvVarConfig() EnvVarConfig {
	return envVarConfig.Load().(EnvVarConfig)
}

// NewLoggingConfig returns a new LoggingConfig.
// It reads the environment variables for logging configuration.
//
// Returns:
//   - *LoggingConfig: The logging configuration.
func NewLoggingConfig() *LoggingConfig {
	cfg := GetEnvVarConfig()
	loggingLevel, err := logging.LoggingLevelStrToInt(
		envvar.MustGet(cfg.LoggingLevel),
	)
	if err != nil {
		panic(err)
	}
	return &LoggingConfig{
		LoggingLevel: loggingLevel,
		Compact:      envvar.MustGetBool(cfg.LoggingCompact),
		AnsiCodes:    envvar.MustGetBool(cfg.LoggingAnsiCodes),
	}
}

// SetLoggerFactory allows clients to override the default logger factory.
//
// Parameters:
//   - factory: The logger factory.
func SetLoggerFactory(
	factory func(ctx context.Context) logging.ILogger,
) {
	if factory == nil {
		return
	}
	loggerFactory.Store(factory)
}

// ResetLoggerFactory resets the logger factory to its default.
func ResetLoggerFactory() {
	loggerFactory.Store(defaultCTXLogger)
}

// CtxLogger returns a logger for the given context using the current factory.
//
// Parameters:
//   - ctx: The context to use.
//
// Returns:
//   - ILogger: The logger.
func CtxLogger(ctx context.Context) logging.ILogger {
	factory, ok := loggerFactory.Load().(func(context.Context) logging.ILogger)
	if !ok {
		factory = defaultCTXLogger
	}
	return factory(ctx)
}

// WrapCtxLoggerFactory wraps a context-based logger factory into a
// LoggerFactoryFn.
//
// Parameters:
//   - factoryFn: The context-based logger factory.
//
// Returns:
//   - LoggerFactoryFn: The wrapped logger factory.
func WrapCtxLoggerFactory(
	factoryFn logging.CtxLoggerFactoryFn,
) logging.LoggerFactoryFn {
	return func(params ...any) logging.ILogger {
		var ctx context.Context
		if len(params) > 0 {
			if c, ok := params[0].(context.Context); ok {
				ctx = c
			}
		}
		if ctx == nil {
			ctx = context.Background()
		}
		return factoryFn(ctx)
	}
}

// SetLoggerPlainFactory allows clients to override the default LoggerPlain.
//
// Parameters:
//   - factory: The LoggerPlain factory.
func SetLoggerPlainFactory(factory func(
	startTime time.Time, traceID string, spanID string,
) logging.ILogger,
) {
	if factory == nil {
		return
	}
	loggerPlainFactory.Store(factory)
}

// ResetLoggerPlainFactory resets the LoggerPlain factory to its default.
func ResetLoggerPlainFactory() {
	loggerPlainFactory.Store(defaultLoggerPlain)
}

// LoggerPlain returns a logger with plain formatting using the current factory.
//
// Parameters:
//   - startTime: The start time.
//   - traceID: The trace ID.
//   - spanID: The span ID.
//
// Returns:
//   - ILogger: The logger.
func LoggerPlain(
	startTime time.Time, traceID string, spanID string,
) logging.ILogger {
	factory, ok := loggerPlainFactory.
		Load().(func(time.Time, string, string) logging.ILogger)
	if !ok {
		factory = defaultLoggerPlain
	}
	return factory(startTime, traceID, spanID)
}

// SetEmitterLoggerFactory allows clients to override the default EmitterLogger.
//
// Parameters:
//   - factory: The EmitterLogger factory.
func SetEmitterLoggerFactory(
	factory func([]event.EventEmitter) event.EmitterLogger,
) {
	if factory == nil {
		return
	}
	emitterLoggerFactory.Store(factory)
}

// ResetEmitterLoggerFactory resets the EmitterLogger factory to its default.
func ResetEmitterLoggerFactory() {
	emitterLoggerFactory.Store(defaultEmitterLogger)
}

// EmitterLogger returns an emitter logger using the current factory.
//
// Parameters:
//   - eventEmitter: The event emitter.
//
// Returns:
//   - EmitterLogger: The emitter logger.
func EmitterLogger(
	eventEmitter ...event.EventEmitter,
) event.EmitterLogger {
	factory, ok := emitterLoggerFactory.
		Load().(func([]event.EventEmitter) event.EmitterLogger)
	if !ok {
		factory = defaultEmitterLogger
	}
	return factory(eventEmitter)
}

// defaultCTXLogger is the default implementation of the logger factory.
func defaultCTXLogger(ctx context.Context) logging.ILogger {
	opts := defaultLogOpts(func(ctx context.Context) *logging.ExtraData {
		return extraData(ctx)
	})
	return logging.NewCtxLogger(ctx, opts)
}

// defaultLoggerPlain is the default implementation for LoggerPlain.
func defaultLoggerPlain(
	startTime time.Time, traceID string, spanID string,
) logging.ILogger {
	opts := defaultLogOpts(func(ctx context.Context) *logging.ExtraData {
		return nonCtxExtraData(startTime, traceID, spanID)
	})
	return logging.NewCtxLogger(context.Background(), opts)
}

// defaultEmitterLogger is the default implementation for EmitterLogger.
func defaultEmitterLogger(
	eventEmitter []event.EventEmitter,
) event.EmitterLogger {
	var useEventEmitter event.EventEmitter
	if len(eventEmitter) > 0 && eventEmitter[0] != nil {
		useEventEmitter = eventEmitter[0]
	} else {
		useEventEmitter = event.NewEventEmitter()
	}
	return event.NewEmitterLogger(
		useEventEmitter,
		WrapCtxLoggerFactory(CtxLogger),
	)
}

// extraData returns the extra data for the logger.
func extraData(ctx context.Context) *logging.ExtraData {
	requestMetadata := middleware.GetRequestMetadata(ctx)
	if requestMetadata == nil {
		return nil
	}
	return nonCtxExtraData(
		requestMetadata.TimeStart,
		requestMetadata.TraceID,
		requestMetadata.SpanID,
	)
}

// nonCtxExtraData returns the extra data for the logger without the context.
func nonCtxExtraData(
	timeStart time.Time, traceID string, spanID string,
) *logging.ExtraData {
	now := time.Now().UTC()
	return &logging.ExtraData{
		Time:      &now,
		TimeStart: &timeStart,
		TimeDelta: now.Sub(timeStart).String(),
		TraceID:   traceID,
		SpanID:    spanID,
	}
}

// defaultLogOpts returns the default log options.
func defaultLogOpts(
	getExtraDataFn func(context.Context) *logging.ExtraData,
) *logging.LogOpts {
	config := NewLoggingConfig()
	return &logging.LogOpts{
		LoggingLevel: config.LoggingLevel,
		Compact:      config.Compact,
		AnsiCodes:    config.AnsiCodes,
		LogLevelOpts: logging.DefaultLogLevelOpts(),
		GetExtraData: getExtraDataFn,
	}
}
