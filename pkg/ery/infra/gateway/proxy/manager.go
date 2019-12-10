package proxy

import (
	"context"
	"sync"

	"go.uber.org/zap"

	ery_pb "github.com/srvc/glx/api/ery"
)

type Manager interface {
	Serve(context.Context) error
	Add(context.Context, *ery_pb.App) error
	Delete(ctx context.Context, appID string) error
}

type managerImpl struct {
	factory       ServerFactory
	shutdownFuncs map[string]context.CancelFunc
	log           *zap.Logger
	mu            sync.Mutex
	wg            sync.WaitGroup
}

func NewManager(
	factory ServerFactory,
) Manager {
	return &managerImpl{
		factory:       factory,
		shutdownFuncs: map[string]context.CancelFunc{},
		log:           zap.L().Named("proxy").Named("manager"),
	}
}

func (m *managerImpl) Serve(ctx context.Context) error {
	<-ctx.Done()

	m.mu.Lock()

	for _, f := range m.shutdownFuncs {
		f()
	}

	m.shutdownFuncs = map[string]context.CancelFunc{}
	m.mu.Unlock()

	m.wg.Wait()

	return nil
}

func (m *managerImpl) Add(ctx context.Context, app *ery_pb.App) error {
	s, err := m.factory.Create(ctx, app)
	if err != nil {
		return err
	}

	m.wg.Add(1)

	go func() {
		defer m.wg.Done()

		ctx, cancel := context.WithCancel(context.Background())

		m.mu.Lock()
		m.shutdownFuncs[app.GetAppId()] = cancel
		m.mu.Unlock()

		defer func() {
			m.mu.Lock()
			defer m.mu.Unlock()
			delete(m.shutdownFuncs, app.GetAppId())
		}()

		log := m.log.With(zap.String("app_id", app.GetAppId()))

		log.Debug("proxy servers will start")
		err := s.Serve(ctx)
		m.log.Debug("proxy servers will shutdown")
		if err != nil {
			m.log.Warn("shutdown proxy servers")
		}
	}()

	return nil
}

func (m *managerImpl) Delete(ctx context.Context, appID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if shutdown, ok := m.shutdownFuncs[appID]; ok {
		shutdown()
		delete(m.shutdownFuncs, appID)
	}

	return nil
}
