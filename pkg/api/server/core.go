package server

import (
	"context"
	"errors"

	"github.com/goodieshq/sweettooth/pkg/api"
	"github.com/google/uuid"
)

var (
	ErrAlreadyRegistered = errors.New("node is already registered")
)

type Core interface {
	Close()

	// orgs
	GetOrganization(ctx context.Context, id uuid.UUID) (*api.Organization, error)

	// nodes
	GetNode(ctx context.Context, id uuid.UUID) (*api.Node, error)
	CreateNode(ctx context.Context, req api.RegistrationRequest) error
	UpdateNodePackages(ctx context.Context, nodeid uuid.UUID, packages *api.Packages) error

	Seen(ctx context.Context, id uuid.UUID) error

	GetNodeSchedule(ctx context.Context, id uuid.UUID) (api.Schedule, error)
}
