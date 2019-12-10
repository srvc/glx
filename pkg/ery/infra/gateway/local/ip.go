package local

import (
	"context"
	"errors"
	"fmt"
	"net"
	"runtime"
	"sync"

	"github.com/izumin5210/execx"
)

type IPPool interface {
	Get(context.Context) (net.IP, error)
}

type ipPoolImpl struct {
	loopback net.Interface
	lastByte byte
	m        sync.Mutex
}

func NewIPPool() (IPPool, error) {
	lb, err := getLoopbackInterface()
	if err != nil {
		return nil, err
	}

	return &ipPoolImpl{
		loopback: lb,
		lastByte: 2,
	}, nil
}

func (p *ipPoolImpl) Get(ctx context.Context) (net.IP, error) {
	p.m.Lock()
	defer p.m.Unlock()

	ip := net.IPv4(127, 0, 3, p.lastByte)
	p.lastByte++

	var args []string

	switch runtime.GOOS {
	case "darwin":
		args = []string{p.loopback.Name, "alias", ip.String(), "up"}

	case "linux":
		args = []string{p.loopback.Name, ip.String(), "up"}

	default:
		return net.IP{}, fmt.Errorf("unsupported os: %s", runtime.GOOS)
	}

	err := execx.CommandContext(ctx, "ifconfig", args...).Run()
	if err != nil {
		return net.IP{}, err
	}

	return ip, nil
}

func getLoopbackInterface() (net.Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return net.Interface{}, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			return iface, nil
		}
	}
	return net.Interface{}, errors.New("failed to find loopback interface")
}
