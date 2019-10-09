package domain

import (
	"context"

	api_pb "github.com/srvc/glx/api"
)

type AppRepository interface {
	List(context.Context) ([]*api_pb.App, error)
	GetByHostname(context.Context, string) (*api_pb.App, error)
	Create(context.Context, *api_pb.App) error
	Delete(context.Context, string) error
}
