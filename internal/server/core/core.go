package core

import (
	"context"

	"github.com/goodieshq/sweettooth/pkg/api"
	"github.com/google/uuid"
)

type Core interface {
	Close()
	ErrNotFound(err error) bool                                                        // determines if the err is the equivalent of no SQL rows being found
	Seen(ctx context.Context, nodeid uuid.UUID) error                                  // update last seen attribute of a node
	GetOrganizations(ctx context.Context) ([]*api.Organization, error)                 // get a list of all organizations
	GetOrganizationSummaries(ctx context.Context) ([]*api.OrganizationSummary, error)  // get a list of all organizations
	GetOrganization(ctx context.Context, orgid uuid.UUID) (*api.Organization, error)   // get an organization by ID
	ProcessRegistrationToken(ctx context.Context, token uuid.UUID) (*uuid.UUID, error) // get the organization from a registration token

	// nodes
	GetNode(ctx context.Context, nodeid uuid.UUID) (*api.Node, error)
	CreateNode(ctx context.Context, req api.RegistrationRequest) (*api.Node, error)
	// nodes.approval
	// UpdateNodeApproval(ctx context.Context, nodeid uuid.UUID, approved bool)

	// nodes.packages
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
