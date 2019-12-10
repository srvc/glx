// +build wireinject

package ery

import (
	"context"

	"github.com/google/wire"
)

// New creates a new Engine instance.
func New(context.Context) (*Engine, func(), error) {
	wire.Build(
		wire.Struct(new(Engine), "*"),
		Set,
	)
	return nil, nil, nil
}
