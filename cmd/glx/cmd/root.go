package cmd

import (
	"github.com/izumin5210/clig/pkg/clib"
	"github.com/spf13/cobra"
)

func New(io clib.IO) *cobra.Command {
	cmd := &cobra.Command{
		Use: "glx",
		RunE: func(c *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.SetOut(io.Out())
	cmd.SetErr(io.Err())
	cmd.SetIn(io.In())
	clib.AddLoggingFlags(cmd)

	cmd.AddCommand(
		newStartCmd(),
		newPsCmd(),
		newDaemonCmd(),
		newUpCmd(),
	)

	return cmd
}
