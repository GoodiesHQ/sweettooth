package server

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/goodieshq/sweettooth/pkg/api"
	"github.com/goodieshq/sweettooth/pkg/crypto"
	"github.com/goodieshq/sweettooth/pkg/database"
	"github.com/goodieshq/sweettooth/pkg/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// PGX (postgres) implementation of SweetTooth Core
type CorePGX struct {
	pool *pgxpool.Pool     // connection pool
	q    *database.Queries // sqlc query functions
}

// convert a pgx org to api org
func pgxOrgToCoreOrg(dborg *database.Organization) *api.Organization {
	var org api.Organization
	org.ID = dborg.ID
	org.Name = dborg.Name
	return &org
}

// convert a pgx node to api node
func pgxNodeToCoreNode(dbnode *database.Node) *api.Node {
	var node api.Node

	node.ID = dbnode.ID

	if dbnode.OrganizationID.Valid {
		oid := uuid.UUID(dbnode.OrganizationID.Bytes)
		node.OrganizationID = &oid
	}

	node.PublicKey = dbnode.PublicKey

	if dbnode.Label.Valid {
		node.Label = &dbnode.Label.String
	}

	node.Hostname = dbnode.Hostname
	node.ClientVersion = dbnode.ClientVersion
	node.PendingSources = dbnode.PendingSources
	node.PendingSchedule = dbnode.PendingSchedule
	node.OSKernel = dbnode.OsKernel
	node.OSName = dbnode.OsName
	node.OSMajor = int(dbnode.OsMajor)
	node.OSMinor = int(dbnode.OsMinor)
	node.OSBuild = int(dbnode.OsBuild)

	if dbnode.ConnectedOn.Valid {
		node.ConnectedOn = dbnode.ConnectedOn.Time
	}
	if dbnode.ApprovedOn.Valid {
		node.ApprovedOn = &dbnode.ApprovedOn.Time
	}
	if dbnode.LastSeen.Valid {
		node.LastSeen = &dbnode.LastSeen.Time
	}

	node.Approved = dbnode.Approved

	return &node
}

func (core *CorePGX) Close() {
	core.pool.Close()
}

func (core *CorePGX) Seen(ctx context.Context, id uuid.UUID) error {
	return core.q.CheckInNode(ctx, id)
}

func (core *CorePGX) GetOrganization(ctx context.Context, id uuid.UUID) (*api.Organization, error) {
	org, err := core.q.GetOrganizationByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return pgxOrgToCoreOrg(&org), nil
}

func (core *CorePGX) UpdateNodePackages(ctx context.Context, nodeid uuid.UUID, packages *api.Packages) error {
	return core.q.UpdateNodePackages(ctx, database.UpdateNodePackagesParams{
		ID:               nodeid,
		PackagesChoco:    packages.PackagesChoco,
		PackagesSystem:   packages.PackagesSystem,
		PackagesOutdated: packages.PackagesOutdated,
	})
}

func (core *CorePGX) GetNode(ctx context.Context, id uuid.UUID) (*api.Node, error) {
	node, err := core.q.GetNodeByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return pgxNodeToCoreNode(&node), nil
}

func (core *CorePGX) GetNodeSchedule(ctx context.Context, id uuid.UUID) (api.Schedule, error) {
	q := database.New(core.pool)
	dbschedules, err := q.GetCombinedScheduleByNode(ctx, id)
	if err != nil {
		return nil, err
	}

	// extract all of the entries from all of the schedules assigned
	var sched api.Schedule
	for _, dbsched := range dbschedules {
		sched = append(sched, dbsched.Entries...)
	}

	if sched == nil {
		sched = make(api.Schedule, 0)
	}

	return sched, nil
}

func (core *CorePGX) CreateNode(ctx context.Context, req api.RegistrationRequest) error {
	q := database.New(core.pool)
	var params database.CreateNodeParams

	id, err := util.Base64toPubKey(req.PublicKey)
	if err != nil {
		return err
	}
	params.ID = crypto.Fingerprint(id)
	if req.OrganizationID == nil {
		params.OrganizationID.Valid = false
	} else {
		params.OrganizationID.Bytes = *req.OrganizationID
		params.OrganizationID.Valid = true
	}
	params.PublicKey = req.PublicKey
	params.Hostname = req.Hostname
	params.ClientVersion = req.ClientVersion
	params.OsKernel = req.OSKernel
	params.OsMajor = int32(req.OSMajor)
	params.OsMinor = int32(req.OSMinor)
	params.OsBuild = int32(req.OSBuild)
	params.PackagesChoco = req.PackagesChoco
	params.PackagesSystem = req.PackagesSystem

	err = q.CreateNode(ctx, params)
	if err != nil {
		log.Error().Err(err).Msg("failed to create node")
	}

	// no errors, node was created
	return nil
}

func testPool(ctx context.Context, pool *pgxpool.Pool) error {
	var target = rand.Intn(10000)
	var result int

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	err = conn.QueryRow(ctx, fmt.Sprintf("SELECT %d", target)).Scan(&result)
	if err != nil {
		return err
	}

	if result != target {
		return fmt.Errorf("unexpected value from test query")
	}

	return nil
}

func NewCorePGX(ctx context.Context, connStr string) (*CorePGX, error) {
	log.Info().Str("connstr", connStr).Msg("Setting Database Connection Parameters")
	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// test the pool to determine if it is valid
	if err := testPool(ctx, pool); err != nil {
		return nil, err
	}

	return &CorePGX{
		pool: pool,
		q:    database.New(pool),
	}, nil
}
