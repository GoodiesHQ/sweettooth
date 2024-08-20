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
	ErrNotFound(err error) bool

	Close()

	// update last seen attribute of a node
	Seen(ctx context.Context, nodeid uuid.UUID) error

	// get an organization by ID
	GetOrganization(ctx context.Context, orgid uuid.UUID) (*api.Organization, error)

	// get the organization from a registration token
	ProcessRegistrationToken(ctx context.Context, token uuid.UUID) (*uuid.UUID, error)

	// nodes
	GetNode(ctx context.Context, nodeidid uuid.UUID) (*api.Node, error)
	CreateNode(ctx context.Context, req api.RegistrationRequest) (*api.Node, error)

	// packages
	UpdateNodePackages(ctx context.Context, nodeid uuid.UUID, packages *api.Packages) error
	GetNodePackages(ctx context.Context, nodeid uuid.UUID) (*api.Packages, error)

	// schedule
	GetNodeSchedule(ctx context.Context, nodeid uuid.UUID) (api.Schedule, error)

	// jobs
	GetPackageJobList(ctx context.Context, nodeid uuid.UUID, attemptsMax int) (api.PackageJobList, error)
	GetPackageJob(ctx context.Context, jobid uuid.UUID) (*api.PackageJob, error)
	AttemptPackageJob(ctx context.Context, jobid, nodeid uuid.UUID, attemptsMax int) (*api.PackageJob, error)
	CompletePackageJob(ctx context.Context, jobid, nodeid uuid.UUID, result *api.PackageJobResult) error
}
