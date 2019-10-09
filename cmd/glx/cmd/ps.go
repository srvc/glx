package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/srvc/appctx"
	api_pb "github.com/srvc/glx/api"
)

func newPsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ps",
		Short: "List apps",
		RunE: func(c *cobra.Command, args []string) error {
			ctx := appctx.Global()

			conn, err := grpc.DialContext(ctx, "api.glx.local:80", grpc.WithInsecure())
			if err != nil {
				return err
			}
			appAPI := api_pb.NewAppServiceClient(conn)
			resp, err := appAPI.ListApps(ctx, new(api_pb.ListAppsRequest))
			if err != nil {
				return err
			}

			for _, app := range resp.GetApps() {
				fmt.Fprintf(c.OutOrStdout(), "%s\t%s\t%s -> %s\n", app.GetAppId()[:7], app.GetName(), app.GetHostname(), app.GetIp())
			}
			return nil
		},
	}

	return cmd
}
