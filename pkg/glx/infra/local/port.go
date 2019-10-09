package local

import (
	"context"
	"sync"

	"github.com/srvc/glx"
	"github.com/srvc/glx/pkg/glx/domain"
	netutil "github.com/srvc/glx/pkg/util/net"
)

type PortPool struct {
	m sync.Mutex
}

var _ domain.PortPool = (*PortPool)(nil)

func NewPortPool() *PortPool {
	return &PortPool{}
}

func (p *PortPool) Get(ctx context.Context) (glx.Port, error) {
	p.m.Lock()
	defer p.m.Unlock()

	port, err := netutil.GetFreePort()
	if err != nil {
		return glx.Port(0), err
	}

	return glx.Port(port), nil
}
