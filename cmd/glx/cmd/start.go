package cmd

import (
	"github.com/spf13/cobra"
	"github.com/srvc/appctx"

	"github.com/srvc/glx/pkg/ery"
)

func newStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start server",
		RunE: func(c *cobra.Command, args []string) error {
			ctx := appctx.Global()

			ery, closeFunc, err := ery.New(ctx)
			if err != nil {
				return err
			}
			defer closeFunc()

			return ery.Run(ctx)
		},
	}

	return cmd
}
