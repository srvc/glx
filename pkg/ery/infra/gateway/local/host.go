package local

import (
	"context"
	"net"
	"sync"

	"github.com/txn2/txeh"
)

type Hosts interface {
	Add(ctx context.Context, hostname string, ip net.IP) error
	Delete(ctx context.Context, hostname string) error
}

func NewHosts() (Hosts, error) {
	h, err := txeh.NewHostsDefault()
	if err != nil {
		return nil, err
	}

	// TODO: should backup /etc/hosts file

	return &hostsImpl{hosts: h}, nil
}

type hostsImpl struct {
	hosts *txeh.Hosts
	mu    sync.Mutex
}

func (h *hostsImpl) Add(ctx context.Context, hostname string, ip net.IP) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	err := h.hosts.Reload()
	if err != nil {
		return err
	}

	h.hosts.AddHost(ip.String(), hostname)

	return h.hosts.Save()
}

func (h *hostsImpl) Delete(ctx context.Context, hostname string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	err := h.hosts.Reload()
	if err != nil {
		return err
	}

	h.hosts.RemoveHost(hostname)

	return h.hosts.Save()
}
