package local

import (
	"context"
	"net"
	"sync"

	"github.com/srvc/glx/pkg/ery/domain"
)

type PortPool interface {
	Get(context.Context) (domain.Port, error)
}

type portPoolImpl struct {
	m sync.Mutex
}

func NewPortPool() PortPool {
	return &portPoolImpl{}
}

func (p *portPoolImpl) Get(ctx context.Context) (domain.Port, error) {
	p.m.Lock()
	defer p.m.Unlock()

	port, err := getFreePort()
	if err != nil {
		return domain.Port(0), err
	}

	return domain.Port(port), nil
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}

	lis, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer lis.Close()

	return lis.Addr().(*net.TCPAddr).Port, nil
}
