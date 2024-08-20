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
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// PGX (postgres) implementation of SweetTooth Core
type CorePGX struct {
	pool *pgxpool.Pool     // connection pool
	q    *database.Queries // sqlc query functions
}

func (CorePGX) ErrNotFound(err error) bool {
	return err == pgx.ErrNoRows
}

// convert a pgx org to api org
func pgxOrgToCoreOrg(dborg *database.Organization) *api.Organization {
	var org api.Organization
	org.ID = dborg.ID
	org.Name = dborg.Name
	return &org
}

// convert a pgx package job to api package job
func pgxPackageJobToCorePackageJob(dbjob *database.PackageJob) *api.PackageJob {
	var job api.PackageJob

	// base values of the job
	job.ID = dbjob.ID
	job.NodeID = dbjob.NodeID
	if dbjob.GroupID.Valid {
		gid := uuid.UUID(dbjob.GroupID.Bytes)
		job.GroupID = &gid
	}
	job.OrganizationID = dbjob.OrganizationID
	job.Attempts = int(dbjob.Attempts)

	job.Action = int(dbjob.Action)

	// set the parameters
	job.Parameters.Name = dbjob.Name
	if dbjob.Version.Valid {
		job.Parameters.Version = &dbjob.Version.String
	}
	job.Parameters.IgnoreChecksum = dbjob.IgnoreChecksum
	job.Parameters.InstallOnUpgrade = dbjob.InstallOnUpgrade
	job.Parameters.Force = dbjob.Force
	job.Parameters.VerboseOutput = dbjob.VerboseOutput
	job.Parameters.NotSilent = dbjob.NotSilent

	// set the result
	job.Result = &api.PackageJobResult{}
	job.Result.Status = int(dbjob.Status)
	if dbjob.Output.Valid {
		job.Result.Output = dbjob.Output.String
	}

	// final metadata
	job.CreatedAt = dbjob.CreatedAt.Time
	if dbjob.ExpiresAt.Valid {
		job.ExpiresAt = &dbjob.ExpiresAt.Time
	}
	if dbjob.AttemptedAt.Valid {
		job.AttemptedAt = &dbjob.AttemptedAt.Time
	}
	if dbjob.CompletedAt.Valid {
		job.CompletedAt = &dbjob.CompletedAt.Time
	}

	return &job
}

// convert a pgx node to api node
func pgxNodeToCoreNode(dbnode *database.Node) *api.Node {
	var node api.Node

	node.ID = dbnode.ID
	node.OrganizationID = &dbnode.OrganizationID
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

func (core *CorePGX) GetOrganization(ctx context.Context, orgid uuid.UUID) (*api.Organization, error) {
	org, err := core.q.GetOrganizationByID(ctx, orgid)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		log.Error().Err(err).Msg("failed to get organization")
		return nil, err
	}
	return pgxOrgToCoreOrg(&org), nil
}

func (core *CorePGX) ProcessRegistrationToken(ctx context.Context, token uuid.UUID) (*uuid.UUID, error) {
	orgid, err := core.q.GetValidRegistrationToken(ctx, token)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		log.Error().Err(err).Msg("failed to get registration token")
		return nil, err
	}

	return &orgid, nil
}

func (core *CorePGX) UpdateNodePackages(ctx context.Context, nodeid uuid.UUID, packages *api.Packages) error {
	err := core.q.UpdateNodePackages(ctx, database.UpdateNodePackagesParams{
		ID:               nodeid,
		PackagesChoco:    packages.PackagesChoco,
		PackagesSystem:   packages.PackagesSystem,
		PackagesOutdated: packages.PackagesOutdated,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to update packages")
	}
	return err
}

func (core *CorePGX) GetNodePackages(ctx context.Context, nodeid uuid.UUID) (*api.Packages, error) {
	var packages api.Packages
	pkg, err := core.q.GetNodePackages(ctx, nodeid)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		log.Error().Err(err).Send()
		return nil, err
	}

	packages.PackagesChoco = pkg.PackagesChoco
	packages.PackagesSystem = pkg.PackagesSystem
	packages.PackagesOutdated = pkg.PackagesOutdated

	return &packages, nil
}

func (core *CorePGX) GetNode(ctx context.Context, nodeid uuid.UUID) (*api.Node, error) {
	node, err := core.q.GetNodeByID(ctx, nodeid)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return pgxNodeToCoreNode(&node), nil
}

func (core *CorePGX) GetNodeSchedule(ctx context.Context, nodeid uuid.UUID) (api.Schedule, error) {
	q := database.New(core.pool)
	dbschedules, err := q.GetCombinedScheduleByNode(ctx, nodeid)
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

func (core *CorePGX) CreateNode(ctx context.Context, req api.RegistrationRequest) (*api.Node, error) {
	q := database.New(core.pool)
	var params database.CreateNodeParams

	pubkey, err := util.Base64toPubKey(req.PublicKey)
	if err != nil {
		return nil, err
	}

	// get the fingerprint of the public key
	params.ID = crypto.Fingerprint(pubkey)
	params.ID_2 = req.Token

	// Client information
	params.PublicKey = req.PublicKey
	params.Hostname = req.Hostname
	params.ClientVersion = req.ClientVersion

	// OS Info
	params.OsName = req.OSName
	params.OsKernel = req.OSKernel
	params.OsMajor = int32(req.OSMajor)
	params.OsMinor = int32(req.OSMinor)
	params.OsBuild = int32(req.OSBuild)

	// Package info
	params.PackagesChoco = req.PackagesChoco
	params.PackagesSystem = req.PackagesSystem
	params.PackagesOutdated = req.PackagesOutdated

	node, err := q.CreateNode(ctx, params)
	log.Info().Msg("Creating Node:")
	if err != nil {
		log.Error().Err(err).Msg("failed to create node")
		return nil, err
	}

	// no errors, node was created
	return pgxNodeToCoreNode(&node), nil
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

func (core *CorePGX) GetPackageJobList(ctx context.Context, nodeid uuid.UUID, attemptsMax int) (api.PackageJobList, error) {
	joblist, err := core.q.GetPackageJobListByNodeID(ctx, database.GetPackageJobListByNodeIDParams{
		NodeID:   nodeid,
		Attempts: int32(attemptsMax),
	})

	if err != nil {
		log.Error().Err(err).Msg("failed to query db GetPackageJobListByNodeID")
		return nil, err
	}

	if joblist == nil {
		joblist = []uuid.UUID{}
	}

	return api.PackageJobList(joblist), nil
}

func (core *CorePGX) GetPackageJob(ctx context.Context, jobid uuid.UUID) (*api.PackageJob, error) {
	log.Info().Msg("OKAY")
	log.Info().Msg("1")
	job, err := core.q.GetPackageJobByID(ctx, jobid)
	log.Info().Msg("2")

	if err != nil {
		log.Info().Msg("3")
		if err == pgx.ErrNoRows {
			log.Info().Msg("4")
			return nil, nil
		}

		log.Error().Err(err).Msg("failed to get package job by ID")
		return nil, err
	}
	log.Info().Msg("5")

	return pgxPackageJobToCorePackageJob(&job), nil
}

func (core *CorePGX) AttemptPackageJob(ctx context.Context, jobid, nodeid uuid.UUID, attemptsMax int) (*api.PackageJob, error) {
	job, err := core.q.AttemptPackageJob(ctx, database.AttemptPackageJobParams{
		ID:       jobid,
		NodeID:   nodeid,
		Attempts: int32(attemptsMax),
	})

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return pgxPackageJobToCorePackageJob(&job), nil
}

func (core *CorePGX) CompletePackageJob(ctx context.Context, jobid, nodeid uuid.UUID, result *api.PackageJobResult) error {
	dberr := pgtype.Text{}

	if result.Error != nil {
		dberr.Valid = true
		dberr.String = *result.Error
	}

	_, err := core.q.CompletePackageJob(ctx, database.CompletePackageJobParams{
		ID:     jobid,
		NodeID: nodeid,
		Status: int32(result.Status),
		ExitCode: pgtype.Int4{
			Int32: int32(result.ExitCode),
			Valid: true,
		},
		Output: pgtype.Text{
			String: result.Output,
			Valid:  true,
		},
		Error: dberr,
	})
	return err
}
