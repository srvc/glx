package cmd

import (
	"context"
	"fmt"
	"sync"

	"github.com/docker/docker/client"
	"github.com/izumin5210/clig/pkg/clib"
	"github.com/spf13/cobra"
	"github.com/srvc/appctx"
	"github.com/srvc/glx/cmd/glx/cmd/up"
	"github.com/srvc/glx/pkg/ery"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/srvc/glx"
	ery_pb "github.com/srvc/glx/api/ery"
)

func newUpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Up server",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := appctx.Global()

			fs := glx.NewFs()
			uFs, err := glx.NewUnionFs(fs)
			if err != nil {
				return err
			}
			viper := glx.NewViper(uFs)
			cfg, err := glx.NewConfig(viper)
			if err != nil {
				return err
			}

			conn, err := grpc.DialContext(ctx, ery.Hostname, grpc.WithInsecure())
			if err != nil {
				return err
			}
			appAPI := ery_pb.NewAppServiceClient(conn)

			dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil {
				return err
			}

			proj := cfg.FindProject(args[0])
			if proj == nil {
				return fmt.Errorf("Project %q was not found", args[0])
			}

			io := &clib.IOContainer{
				InR:  c.InOrStdin(),
				OutW: c.OutOrStdout(),
				ErrW: c.OutOrStderr(),
			}

			runner := up.New(
				appAPI,
				up.NewLocalRunnerFactory(
					cfg.Root,
					io,
				),
				up.NewDockerRunnerFactory(
					cfg.Root,
					io,
					dockerClient,
				),
			)

			var wg sync.WaitGroup
			for _, app := range proj.Apps {
				app := app
				wg.Add(1)
				go func() {
					defer wg.Done()

					err := runner.Run(ctx, app)
					if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
						zap.L().Error("unexpected exit app", zap.Any("app", app), zap.Error(err))
					}
					// TODO: auto restart?
				}()
			}

			wg.Wait()

			return nil
		},
	}

	return cmd
}
