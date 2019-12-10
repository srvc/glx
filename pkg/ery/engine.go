package ery

import (
	"context"
	"fmt"

	"github.com/izumin5210/grapi/pkg/grapiserver"

	ery_pb "github.com/srvc/glx/api/ery"
	"github.com/srvc/glx/pkg/ery/server"
)

type Engine struct {
	appServiceServer server.AppServiceServer
}

var (
	AppName  = "github.com/srvc/glx/pkg/ery"
	Port     = 80
	Hostname = "ery.glx.srvc.local"
)

func (e *Engine) Run(ctx context.Context) error {
	app := &ery_pb.App{
		Name:     AppName,
		Hostname: Hostname,
		Ports: []*ery_pb.App_Port{
			{
				Network:       ery_pb.App_Port_TCP,
				RequestedPort: uint32(Port),
			},
		},
	}

	app, err := e.appServiceServer.CreateApp(ctx, &ery_pb.CreateAppRequest{})
	if err != nil {
		return nil
	}
	addr := fmt.Sprintf("%s:%d", app.GetIp(), app.Ports[0].GetAssignedPort())

	return grapiserver.New(
		grapiserver.WithDefaultLogger(),
		grapiserver.WithSignalHandling(false),
		grapiserver.WithAddr("tcp", addr),
		grapiserver.WithServers(
			e.appServiceServer,
		),
	).ServeContext(ctx)
}
