package main

import (
	"context"
	"fmt"
	"os"

	"github.com/izumin5210/clig/pkg/clib"
	"github.com/spf13/cobra"
	"github.com/srvc/glx"
	"github.com/srvc/glx/pkg/server/proxy"
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
		Src, Dest *glx.Addr
		Network   glx.Network
	}{
		Src:     &glx.Addr{Port: 80},
		Dest:    &glx.Addr{Port: 8080},
		Network: glx.NetworkTCP,
	}

	cmd := &cobra.Command{
		RunE: func(c *cobra.Command, args []string) error {
			zap.L().Info("load config successfully", zap.Any("config", cfg))

			var server interface {
				Serve(context.Context) error
			}

			switch cfg.Network {
			case glx.NetworkTCP:
				server = proxy.NewTCPServer(cfg.Src, cfg.Dest)
			case glx.NetworkUDP:
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

	cmd.Flags().Var(cfg.Src, "src-addr", "")
	cmd.Flags().Var(cfg.Dest, "dest-addr", "")
	cmd.Flags().Var(&cfg.Network, "network", "")

	return cmd
}
