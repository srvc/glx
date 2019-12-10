package main

import (
	"context"
	"fmt"
	"os"

	"github.com/izumin5210/clig/pkg/clib"
	"github.com/spf13/cobra"
	"github.com/srvc/glx/pkg/proxy"
	"go.uber.org/zap"
)

func main() {
	defer clib.Close()

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	return newCommand(clib.Stdio()).Execute()
}

func newCommand(io clib.IO) *cobra.Command {
	cfg := struct {
		Src, Dest string
		Network   string
	}{
		Src:     ":80",
		Dest:    ":8080",
		Network: "tcp",
	}

	cmd := &cobra.Command{
		RunE: func(c *cobra.Command, args []string) error {
			zap.L().Info("load config successfully", zap.Any("config", cfg))

			var server interface {
				Serve(context.Context) error
			}

			switch cfg.Network {
			case "tcp":
				var err error
				server, err = proxy.NewTCPServer(cfg.Src, cfg.Dest)
				if err != nil {
					return fmt.Errorf("failed to create a server: %w", err)
				}
			case "udp":
				return fmt.Errorf("unsupported network type: %s", cfg.Network)
			default:
				panic("unreachable")
			}

			return server.Serve(context.Background())
		},
	}

	cmd.SetOut(io.Out())
	cmd.SetErr(io.Err())
	cmd.SetIn(io.In())
	clib.AddLoggingFlags(cmd)

	cmd.Flags().StringVar(&cfg.Src, "src-addr", "", "")
	cmd.Flags().StringVar(&cfg.Dest, "dest-addr", "", "")
	cmd.Flags().StringVar(&cfg.Network, "network", "", "")

	return cmd
}
