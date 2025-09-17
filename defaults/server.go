package defaults

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/aatuh/envvar"
	"github.com/aatuh/pureapi-core/endpoint"
	"github.com/aatuh/pureapi-core/event"
	"github.com/aatuh/pureapi-core/server"
)

// ServerConfig holds the server configuration.
type ServerConfig struct {
	Port int
}

// ServerEnvConfig holds the names of the environment variables used.
type ServerEnvConfig struct {
	ServerPort string
}

// serverEnvConfig holds the server environment variable config.
var serverEnvConfig atomic.Value

// init sets the default server environment variable config.
func init() {
	serverEnvConfig.Store(ServerEnvConfig{
		ServerPort: "SERVER_PORT",
	})
}

// SetServerEnvConfig allows clients to override the default server env var
// names.
//
// Parameters:
//   - cfg: The server environment variable names.
func SetServerEnvConfig(cfg ServerEnvConfig) {
	serverEnvConfig.Store(cfg)
}

// GetServerEnvConfig returns the current server env var configuration.
//
// Returns:
//   - ServerEnvConfig: The server environment variable configuration.
func GetServerEnvConfig() ServerEnvConfig {
	return serverEnvConfig.Load().(ServerEnvConfig)
}

// NewServerConfig returns a new instance of the server config.
//
// Returns:
//   - *ServerConfig: The server config.
func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		Port: envvar.MustGetInt(GetServerEnvConfig().ServerPort),
	}
}

// StartServer starts the server.
//
// Parameters:
//   - endpoints: The endpoints to serve.
//
// Returns:
//   - error: An error if the server fails to start.
func StartServer(endpoints []endpoint.Endpoint) error {
	timeStart := time.Now().UTC()
	traceID := (&UUIDGen{}).MustRandom().String()
	spanID := (&UUIDGen{}).MustRandom().String()
	eventEmitter := event.
		NewEventEmitter().
		RegisterGlobalListener(func(event *event.Event) {
			logServerEvent(event, timeStart, traceID, spanID)
		})

	serverHandler := server.NewHandler(EmitterLogger(eventEmitter))
	err := server.StartServer(
		serverHandler,
		server.DefaultHTTPServer(
			serverHandler, NewServerConfig().Port, endpoints,
		),
		nil,
	)
	if err != nil {
		return fmt.Errorf("StartServer: server panic: %w", err)
	}

	return nil
}

// StartServer starts the server. It panics if an error occurs.
//
// Parameters:
//   - endpoints: The endpoints to serve.
func MustStartServer(endpoints []endpoint.Endpoint) {
	if err := StartServer(endpoints); err != nil {
		panic(fmt.Errorf("MustStartServer: %w", err))
	}
}

// logServerEvent logs server events.
func logServerEvent(
	event *event.Event, timeStart time.Time, traceID string, spanID string,
) {
	switch event.Type {
	case server.EventPanic:
		LoggerPlain(timeStart, traceID, spanID).Errorf(
			"Server panic event: %s %+v", event.Message, event.Data,
		)
	case server.EventNotFound:
		LoggerPlain(timeStart, traceID, spanID).Warn(event.Message)
	case server.EventMethodNotAllowed:
		LoggerPlain(timeStart, traceID, spanID).Warn(event.Message)
	default:
		LoggerPlain(timeStart, traceID, spanID).Infof(
			"%s %v", event.Message, event.Data,
		)
	}
}
