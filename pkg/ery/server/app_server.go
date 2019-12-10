package server

import (
	"context"
	"net"
	"sync"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/izumin5210/grapi/pkg/grapiserver"
	"go.uber.org/zap"

	ery_pb "github.com/srvc/glx/api/ery"
	"github.com/srvc/glx/pkg/ery/infra/gateway/local"
	"github.com/srvc/glx/pkg/ery/infra/gateway/proxy"
	"github.com/srvc/glx/pkg/ery/infra/repository"
)

// AppServiceServer is a composite interface of ery_pb.AppServiceServer and grapiserver.Server.
type AppServiceServer interface {
	ery_pb.AppServiceServer
	grapiserver.Server
}

// NewAppServiceServer creates a new AppServiceServer instance.
func NewAppServiceServer(
	appRepo repository.App,
	hosts local.Hosts,
	proxies proxy.Manager,
) AppServiceServer {
	return &appServiceServerImpl{
		appRepo: appRepo,
		hosts:   hosts,
		proxies: proxies,
		log:     zap.L().Named("server"),
	}
}

type appServiceServerImpl struct {
	mu      sync.RWMutex
	appRepo repository.App
	hosts   local.Hosts
	proxies proxy.Manager
	log     *zap.Logger
}

func (s *appServiceServerImpl) ListApps(ctx context.Context, req *ery_pb.ListAppsRequest) (*ery_pb.ListAppsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	apps, err := s.appRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	return &ery_pb.ListAppsResponse{Apps: apps}, nil
}

func (s *appServiceServerImpl) CreateApp(ctx context.Context, req *ery_pb.CreateAppRequest) (*ery_pb.App, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	app := req.GetApp()
	err := s.appRepo.Create(ctx, app)
	if err != nil {
		return nil, err
	}

	ip := net.ParseIP(app.GetIp())
	err = s.hosts.Add(ctx, app.GetHostname(), ip)
	if err != nil {
		dErr := s.appRepo.Delete(ctx, app.GetAppId())
		if dErr != nil {
			s.log.Warn("failed to remove app", zap.String("app_id", app.GetAppId()), zap.Error(dErr))
		}
		return nil, err
	}

	if app.GetProxyRequired() {
		err = s.proxies.Add(ctx, app)
		if err != nil {
		}
	}

	return app, nil
}

func (s *appServiceServerImpl) DeleteApp(ctx context.Context, req *ery_pb.DeleteAppRequest) (*empty.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	app, err := s.appRepo.Get(ctx, req.GetAppId())
	if err != nil {
		return nil, err
	}
	err = s.appRepo.Delete(ctx, app.GetAppId())
	if err != nil {
		return nil, err
	}
	err = s.hosts.Delete(ctx, app.GetHostname())
	if err != nil {
		return nil, err
	}
	return new(empty.Empty), nil
}
