//go:build wireinject
// +build wireinject

package wire

import (
	"admin-webrtc-go/internal/handler"
	"admin-webrtc-go/internal/repository"
	"admin-webrtc-go/internal/server"
	"admin-webrtc-go/internal/service"
	"admin-webrtc-go/pkg/app"
	"admin-webrtc-go/pkg/jwt"
	"admin-webrtc-go/pkg/log"
	"admin-webrtc-go/pkg/server/http"
	"admin-webrtc-go/pkg/sid"
	"github.com/google/wire"
	"github.com/spf13/viper"
)

var repositorySet = wire.NewSet(
	repository.NewDB,
	//repository.NewRedis,
	repository.NewRepository,
	repository.NewTransaction,
	repository.NewUserRepository,
)

var serviceSet = wire.NewSet(
	service.NewService,
	service.NewUserService,
)

var handlerSet = wire.NewSet(
	handler.NewHandler,
	handler.NewUserHandler,
)

var serverSet = wire.NewSet(
	server.NewHTTPServer,
	server.NewJob,
)

// build App
func newApp(httpServer *http.Server, job *server.Job) *app.App {
	return app.NewApp(
		app.WithServer(httpServer, job),
		app.WithName("demo-server"),
	)
}

func NewWire(*viper.Viper, *log.Logger) (*app.App, func(), error) {

	panic(wire.Build(
		repositorySet,
		serviceSet,
		handlerSet,
		serverSet,
		sid.NewSid,
		jwt.NewJwt,
		newApp,
	))
}
