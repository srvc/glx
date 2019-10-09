package up

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/izumin5210/clig/pkg/clib"
	"github.com/izumin5210/execx"
	"go.uber.org/zap"

	"github.com/srvc/glx"
	api_pb "github.com/srvc/glx/api"
	"github.com/srvc/glx/pkg/util/prefixer"
)

func NewLocalRunnerFactory(
	rootDir string,
	io clib.IO,
) *LocalRunnerFactory {
	return &LocalRunnerFactory{
		rootDir: rootDir,
		io:      io,
	}
}

type LocalRunnerFactory struct {
	rootDir string
	io      clib.IO
	log     *zap.Logger
}

func (f *LocalRunnerFactory) GetRunner(app *glx.App, appPb *api_pb.App) Runner {
	return &LocalRunner{
		LocalRunnerFactory: f,
		app:                app,
		appPb:              appPb,
		log: zap.L().With(
			zap.String("app_name", app.Name),
			zap.String("app_type", "local"),
			zap.Any("app", appPb),
		),
	}
}

type LocalRunner struct {
	*LocalRunnerFactory
	app   *glx.App
	appPb *api_pb.App
	log   *zap.Logger
}

func (r *LocalRunner) Run(ctx context.Context) error {
	cmd := execx.CommandContext(ctx, r.app.Local.Cmd[0], r.app.Local.Cmd[1:]...)
	cmd.Dir = filepath.Join(r.rootDir, r.app.Local.Path)
	cmd.Env = os.Environ()
	for _, port := range r.appPb.Ports {
		if port.Env != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%d", port.Env, port.InternalPort))
		}
	}
	cmd.Stdin = r.io.In()
	cmd.Stdout = prefixer.NewWriter(r.io.Out(), r.app.Name, "")
	cmd.Stderr = prefixer.NewWriter(r.io.Err(), r.app.Name, "")
	r.log.Info("start")
	err := cmd.Run()
	if err != nil {
		r.log.Warn("shutdown", zap.Error(err))
	}
	return nil
}
