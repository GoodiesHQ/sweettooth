package server

import (
	"context"
	"fmt"

	"github.com/goodieshq/sweettooth/pkg/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type CorePGX struct {
	pool *pgxpool.Pool
}

func pgxNodeToCoreNode(dbnode *database.Node) *Node {
	var node Node

	node.ID = uuid.UUID(dbnode.ID.Bytes)

	if dbnode.OrganizationID.Valid {
		oid := new(uuid.UUID)
		copy(oid[:], dbnode.OrganizationID.Bytes[:])
		node.OrganizationID = oid
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

func (core *CorePGX) GetNodeByID(id uuid.UUID) (*Node, error) {
	q := database.New(core.pool)
	node, err := q.GetNodeByID(context.Background(), pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return pgxNodeToCoreNode(&node), nil
}

func testPool(ctx context.Context, pool *pgxpool.Pool) error {
	const target = 1

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	var result int
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
	}, nil
}
