package ery

import (
	"github.com/docker/docker/client"
	"github.com/google/wire"

	"github.com/srvc/glx/pkg/ery/infra/gateway/local"
	"github.com/srvc/glx/pkg/ery/infra/gateway/proxy"
	"github.com/srvc/glx/pkg/ery/infra/repository"
	"github.com/srvc/glx/pkg/ery/server"
)

var Set = wire.NewSet(
	local.NewHosts,
	local.NewIPPool,
	local.NewPortPool,
	repository.NewApp,
	proxy.NewManager,
	proxy.NewDockerServerFactory,
	wire.Bind(new(proxy.ServerFactory), new(*proxy.DockerServerFactory)),
	server.NewAppServiceServer,
	provideDockerClient,
)

func provideDockerClient() (*client.Client, func(), error) {
	c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, nil, err
	}

	return c, func() { c.Close() }, nil
}
