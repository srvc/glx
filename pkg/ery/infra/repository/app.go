package repository

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"

	"go.uber.org/zap"

	ery_pb "github.com/srvc/glx/api/ery"
	"github.com/srvc/glx/pkg/ery/infra/gateway/local"
)

type App interface {
	List(context.Context) ([]*ery_pb.App, error)
	GetByHostname(context.Context, string) (*ery_pb.App, error)
	Create(context.Context, *ery_pb.App) error
	Delete(context.Context, string) error
}

type appImpl struct {
	sync.Mutex
	m          sync.Map
	byHostname sync.Map
	ipPool     local.IPPool
	portPool   local.PortPool
	log        *zap.Logger
}

func NewApp(
	ipPool local.IPPool,
	portPool local.PortPool,
) App {
	return &appImpl{
		ipPool:   ipPool,
		portPool: portPool,
		log:      zap.L().Named("mem"),
	}
}

func (r *appImpl) List(context.Context) ([]*ery_pb.App, error) {
	r.Lock()
	defer r.Unlock()

	apps := []*ery_pb.App{}
	r.m.Range(func(_, v interface{}) bool {
		if app, ok := v.(*ery_pb.App); ok {
			apps = append(apps, app)
		}
		return true
	})
	return apps, nil
}

func (r *appImpl) GetByHostname(_ context.Context, hostname string) (*ery_pb.App, error) {
	r.Lock()
	defer r.Unlock()

	v, ok := r.byHostname.Load(hostname)
	if ok {
		v, ok := r.m.Load(v.(string))
		if ok {
			if app, ok := v.(*ery_pb.App); ok {
				return app, nil
			}
		}
	}
	return nil, fmt.Errorf("%s is not found", hostname)
}

func (r *appImpl) Create(ctx context.Context, app *ery_pb.App) error {
	r.Lock()
	defer r.Unlock()

	if app.AppId == "" {
		k := make([]byte, 16)
		if _, err := rand.Read(k); err != nil {
			return err
		}
		app.AppId = fmt.Sprintf("%x", k)
	}
	if app.Ip == "" {
		ip, err := r.ipPool.Get(ctx)
		if err != nil {
			return err
		}
		app.Ip = ip.String()
	}
	for _, port := range app.GetPorts() {
		if port.GetAssignedPort() == 0 {
			p, err := r.portPool.Get(ctx)
			if err != nil {
				return err
			}
			port.AssignedPort = uint32(p)
		}
	}
	r.m.Store(app.GetAppId(), app)
	r.byHostname.Store(app.GetHostname(), app.GetAppId())

	r.log.Debug("registered a new app", zap.Any("app", app))

	return nil
}

func (r *appImpl) Delete(_ context.Context, id string) error {
	r.Lock()
	defer r.Unlock()

	v, ok := r.m.Load(id)
	if !ok {
		return fmt.Errorf("%s is not found", id)
	}
	r.m.Delete(id)
	r.byHostname.Delete(v.(*ery_pb.App).GetHostname())
	return nil
}
