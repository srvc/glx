package domain

import (
	"context"
	"net"

	"github.com/srvc/glx"
)

type IPPool interface {
	Get(context.Context) (net.IP, error)
}

type PortPool interface {
	Get(context.Context) (glx.Port, error)
}
