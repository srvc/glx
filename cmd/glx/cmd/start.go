package cmd

import (
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"github.com/srvc/appctx"
	"golang.org/x/sync/errgroup"

	"github.com/srvc/glx/pkg/glx/infra/local"
	"github.com/srvc/glx/pkg/glx/infra/mem"
	"github.com/srvc/glx/pkg/server/api"
	"github.com/srvc/glx/pkg/server/dns"
	"github.com/srvc/glx/pkg/server/proxy"
)

func newStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start server",
		RunE: func(c *cobra.Command, args []string) error {
			ctx := appctx.Global()

			ipPool, err := local.NewIPPool()
			if err != nil {
				return err
			}
			portPool := local.NewPortPool()
			appRepo := mem.NewAppRepository(ipPool, portPool)
			dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil {
				return err
			}
			proxies := proxy.NewManager(dockerClient)
			dns := dns.NewServer(appRepo)
			api := api.NewServer(appRepo, proxies)

			eg, ctx := errgroup.WithContext(ctx)
			eg.Go(func() error { return proxies.Serve(ctx) })
			eg.Go(func() error { return dns.Serve(ctx) })
			eg.Go(func() error { return api.Serve(ctx) })

			return eg.Wait()
		},
	}

	return cmd
}
