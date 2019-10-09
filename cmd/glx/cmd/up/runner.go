package up

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/srvc/glx"
	api_pb "github.com/srvc/glx/api"
)

type Runner interface {
	Run(ctx context.Context) error
}

type Factory interface {
	GetRunner(app *glx.App, appPb *api_pb.App) Runner
}

func New(
	appAPI api_pb.AppServiceClient,
	local *LocalRunnerFactory,
	docker *DockerRunnerFactory,
) *RunnerFacade {
	return &RunnerFacade{
		appAPI: appAPI,
		local:  local,
		docker: docker,
		log:    zap.L(),
	}
}

type RunnerFacade struct {
	appAPI        api_pb.AppServiceClient
	local, docker Factory
	log           *zap.Logger
}

func (r *RunnerFacade) Run(ctx context.Context, app *glx.App) error {
	log := r.log.With(zap.String("app_name", app.Name))

	appPb, err := r.appAPI.CreateApp(ctx, &api_pb.CreateAppRequest{App: app.Pb()})
	if err != nil {
		r.log.Error("failed to register app", zap.Error(err))
		return err
	}

	log = r.log.With(zap.Any("app", appPb))
	log.Debug("a new app is registered")

	defer func() {
		_, err = r.appAPI.DeleteApp(context.Background(), &api_pb.DeleteAppRequest{AppId: appPb.GetAppId()})
		if err != nil {
			log.Error("failed to delete app", zap.Error(err))
		}
		log.Debug("an app is deleted")
	}()

	switch {
	case app.Local != nil:
		return r.local.GetRunner(app, appPb).Run(ctx)
	case app.Docker != nil:
		return r.docker.GetRunner(app, appPb).Run(ctx)
	case app.Kubernetes != nil:
		return errors.New("not yet implemented")
	default:
		return errors.New("unknown app type")
	}
}
