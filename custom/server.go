package custom

import (
	"fmt"
	"time"

	endpointtypes "github.com/pureapi/pureapi-core/endpoint/types"
	"github.com/pureapi/pureapi-core/server"
	"github.com/pureapi/pureapi-core/util"
	"github.com/pureapi/pureapi-core/util/types"
	"github.com/pureapi/pureapi-util/envvar"
)

const envVarServerPort = "SERVER_PORT"

type ServerConfig struct {
	Port int
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		Port: envvar.MustGetInt(envVarServerPort),
	}
}

func MustStartServer(endpoints []endpointtypes.Endpoint) {
	timeStart := time.Now().UTC()
	traceID := (&UUIDGen{}).MustRandom().String()

	eventEmitter := util.NewEventEmitter().RegisterListener(
		server.EventRegisterURL,
		func(event *types.Event) {
			LoggerPlain(timeStart, traceID).Infof(
				"%s %v", event.Message, event.Data,
			)
		},
	).RegisterListener(
		server.EventPanic,
		func(event *types.Event) {
			LoggerPlain(timeStart, traceID).Error(
				"Server panic event: %s %+v", event.Message, event.Data,
			)
		},
	).RegisterListener(
		server.EventNotFound,
		func(event *types.Event) {
			LoggerPlain(timeStart, traceID).Error(event.Message)
		},
	).RegisterListener(
		server.EventMethodNotAllowed,
		func(event *types.Event) {
			LoggerPlain(timeStart, traceID).Error(event.Message)
		},
	).RegisterListener(
		server.EventStart,
		func(event *types.Event) {
			LoggerPlain(timeStart, traceID).Info(event.Message)
		},
	).RegisterListener(
		server.EventErrorStart,
		func(event *types.Event) {
			LoggerPlain(timeStart, traceID).Error(event.Message)
		},
	).RegisterListener(
		server.EventShutDownStarted,
		func(event *types.Event) {
			LoggerPlain(timeStart, traceID).Info(event.Message)
		},
	).RegisterListener(
		server.EventShutDown,
		func(event *types.Event) {
			LoggerPlain(timeStart, traceID).Info(event.Message)
		},
	)

	emitterLogger := EmitterLogger(eventEmitter)
	serverHandler := server.NewHandler(emitterLogger)

	err := server.StartServer(
		serverHandler,
		server.DefaultHTTPServer(
			serverHandler, NewServerConfig().Port, endpoints,
		),
		nil,
	)
	if err != nil {
		panic(fmt.Sprintf("Server error: %s", err))
	}
}
