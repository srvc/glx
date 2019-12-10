package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/srvc/appctx"
	ery_pb "github.com/srvc/glx/api/ery"
	"github.com/srvc/glx/pkg/ery"
)

func newPsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ps",
		Short: "List apps",
		RunE: func(c *cobra.Command, args []string) error {
			ctx := appctx.Global()

			conn, err := grpc.DialContext(ctx, ery.Hostname, grpc.WithInsecure())
			if err != nil {
				return err
			}
			appAPI := ery_pb.NewAppServiceClient(conn)
			resp, err := appAPI.ListApps(ctx, new(ery_pb.ListAppsRequest))
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
