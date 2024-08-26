// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: jobs.sql

package database

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const attemptPackageJob = `-- name: AttemptPackageJob :one
UPDATE
    package_jobs
SET
    attempts = attempts + 1
WHERE
    id = $1 AND node_id=$2 AND attempts < $3
RETURNING id, node_id, group_id, organization_id, attempts, action, name, version, ignore_checksum, install_on_upgrade, force, verbose_output, not_silent, timeout, status, exit_code, output, error, attempted_at, completed_at, expires_at, created_at
`

type AttemptPackageJobParams struct {
	ID       uuid.UUID `db:"id" json:"id"`
	NodeID   uuid.UUID `db:"node_id" json:"node_id"`
	Attempts int32     `db:"attempts" json:"attempts"`
}

func (q *Queries) AttemptPackageJob(ctx context.Context, arg AttemptPackageJobParams) (PackageJob, error) {
	row := q.db.QueryRow(ctx, attemptPackageJob, arg.ID, arg.NodeID, arg.Attempts)
	var i PackageJob
	err := row.Scan(
		&i.ID,
		&i.NodeID,
		&i.GroupID,
		&i.OrganizationID,
		&i.Attempts,
		&i.Action,
		&i.Name,
		&i.Version,
		&i.IgnoreChecksum,
		&i.InstallOnUpgrade,
		&i.Force,
		&i.VerboseOutput,
		&i.NotSilent,
		&i.Timeout,
		&i.Status,
		&i.ExitCode,
		&i.Output,
		&i.Error,
		&i.AttemptedAt,
		&i.CompletedAt,
		&i.ExpiresAt,
		&i.CreatedAt,
	)
	return i, err
}

const completePackageJob = `-- name: CompletePackageJob :one
UPDATE
    package_jobs
SET
    status=$3,
    exit_code=$4,
    output=$5,
    error=$6,
    completed_at=CURRENT_TIMESTAMP
WHERE
    id=$1 AND status=0 AND node_id=$2
RETURNING id, node_id, group_id, organization_id, attempts, action, name, version, ignore_checksum, install_on_upgrade, force, verbose_output, not_silent, timeout, status, exit_code, output, error, attempted_at, completed_at, expires_at, created_at
`

type CompletePackageJobParams struct {
	ID       uuid.UUID   `db:"id" json:"id"`
	NodeID   uuid.UUID   `db:"node_id" json:"node_id"`
	Status   int32       `db:"status" json:"status"`
	ExitCode pgtype.Int4 `db:"exit_code" json:"exit_code"`
	Output   pgtype.Text `db:"output" json:"output"`
	Error    pgtype.Text `db:"error" json:"error"`
}

func (q *Queries) CompletePackageJob(ctx context.Context, arg CompletePackageJobParams) (PackageJob, error) {
	row := q.db.QueryRow(ctx, completePackageJob,
		arg.ID,
		arg.NodeID,
		arg.Status,
		arg.ExitCode,
		arg.Output,
		arg.Error,
	)
	var i PackageJob
	err := row.Scan(
		&i.ID,
		&i.NodeID,
		&i.GroupID,
		&i.OrganizationID,
		&i.Attempts,
		&i.Action,
		&i.Name,
		&i.Version,
		&i.IgnoreChecksum,
		&i.InstallOnUpgrade,
		&i.Force,
		&i.VerboseOutput,
		&i.NotSilent,
		&i.Timeout,
		&i.Status,
		&i.ExitCode,
		&i.Output,
		&i.Error,
		&i.AttemptedAt,
		&i.CompletedAt,
		&i.ExpiresAt,
		&i.CreatedAt,
	)
	return i, err
}

const createPackageJob = `-- name: CreatePackageJob :one
INSERT INTO
    package_jobs(
        node_id,
        group_id,
        organization_id,
        action,
        name,
        version,
        ignore_checksum,
        install_on_upgrade,
        force,
        not_silent
    )
VALUES
    (
        $1, -- Node ID
        $2, -- Group ID (if the job was applied to a group)
        (SELECT organization_id FROM nodes WHERE id=$1), 
        $3, -- action (INSTALL, UPGRADE, UNINSTALL)
        $4, -- package name
        $5, -- verion
        $6, -- ignore checksum
        $7, -- install on upgrade
        $8, -- force
        $9 -- not_silent
    )
RETURNING id, node_id, group_id, organization_id, attempts, action, name, version, ignore_checksum, install_on_upgrade, force, verbose_output, not_silent, timeout, status, exit_code, output, error, attempted_at, completed_at, expires_at, created_at
`

type CreatePackageJobParams struct {
	NodeID           uuid.UUID   `db:"node_id" json:"node_id"`
	GroupID          pgtype.UUID `db:"group_id" json:"group_id"`
	Action           int32       `db:"action" json:"action"`
	Name             string      `db:"name" json:"name"`
	Version          pgtype.Text `db:"version" json:"version"`
	IgnoreChecksum   bool        `db:"ignore_checksum" json:"ignore_checksum"`
	InstallOnUpgrade bool        `db:"install_on_upgrade" json:"install_on_upgrade"`
	Force            bool        `db:"force" json:"force"`
	NotSilent        bool        `db:"not_silent" json:"not_silent"`
}

func (q *Queries) CreatePackageJob(ctx context.Context, arg CreatePackageJobParams) (PackageJob, error) {
	row := q.db.QueryRow(ctx, createPackageJob,
		arg.NodeID,
		arg.GroupID,
		arg.Action,
		arg.Name,
		arg.Version,
		arg.IgnoreChecksum,
		arg.InstallOnUpgrade,
		arg.Force,
		arg.NotSilent,
	)
	var i PackageJob
	err := row.Scan(
		&i.ID,
		&i.NodeID,
		&i.GroupID,
		&i.OrganizationID,
		&i.Attempts,
		&i.Action,
		&i.Name,
		&i.Version,
		&i.IgnoreChecksum,
		&i.InstallOnUpgrade,
		&i.Force,
		&i.VerboseOutput,
		&i.NotSilent,
		&i.Timeout,
		&i.Status,
		&i.ExitCode,
		&i.Output,
		&i.Error,
		&i.AttemptedAt,
		&i.CompletedAt,
		&i.ExpiresAt,
		&i.CreatedAt,
	)
	return i, err
}

const getPackageJobByID = `-- name: GetPackageJobByID :one
SELECT
    id, node_id, group_id, organization_id, attempts, action, name, version, ignore_checksum, install_on_upgrade, force, verbose_output, not_silent, timeout, status, exit_code, output, error, attempted_at, completed_at, expires_at, created_at
FROM
    package_jobs
WHERE
    id=$1
LIMIT 1
`

func (q *Queries) GetPackageJobByID(ctx context.Context, id uuid.UUID) (PackageJob, error) {
	row := q.db.QueryRow(ctx, getPackageJobByID, id)
	var i PackageJob
	err := row.Scan(
		&i.ID,
		&i.NodeID,
		&i.GroupID,
		&i.OrganizationID,
		&i.Attempts,
		&i.Action,
		&i.Name,
		&i.Version,
		&i.IgnoreChecksum,
		&i.InstallOnUpgrade,
		&i.Force,
		&i.VerboseOutput,
		&i.NotSilent,
		&i.Timeout,
		&i.Status,
		&i.ExitCode,
		&i.Output,
		&i.Error,
		&i.AttemptedAt,
		&i.CompletedAt,
		&i.ExpiresAt,
		&i.CreatedAt,
	)
	return i, err
}

const getPackageJobListByNodeID = `-- name: GetPackageJobListByNodeID :many
SELECT
    id
FROM
    package_jobs
WHERE
    node_id=$1 AND status = 0 AND (attempts < $2 OR $2 = 0)
ORDER BY created_at ASC
`

type GetPackageJobListByNodeIDParams struct {
	NodeID   uuid.UUID `db:"node_id" json:"node_id"`
	Attempts int32     `db:"attempts" json:"attempts"`
}

func (q *Queries) GetPackageJobListByNodeID(ctx context.Context, arg GetPackageJobListByNodeIDParams) ([]uuid.UUID, error) {
	rows, err := q.db.Query(ctx, getPackageJobListByNodeID, arg.NodeID, arg.Attempts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPackageJobsByNodeID = `-- name: GetPackageJobsByNodeID :many
SELECT
    id, node_id, group_id, organization_id, attempts, action, name, version, ignore_checksum, install_on_upgrade, force, verbose_output, not_silent, timeout, status, exit_code, output, error, attempted_at, completed_at, expires_at, created_at
FROM
    package_jobs
WHERE
    node_id=$1
ORDER BY created_at ASC
`

func (q *Queries) GetPackageJobsByNodeID(ctx context.Context, nodeID uuid.UUID) ([]PackageJob, error) {
	rows, err := q.db.Query(ctx, getPackageJobsByNodeID, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []PackageJob
	for rows.Next() {
		var i PackageJob
		if err := rows.Scan(
			&i.ID,
			&i.NodeID,
			&i.GroupID,
			&i.OrganizationID,
			&i.Attempts,
			&i.Action,
			&i.Name,
			&i.Version,
			&i.IgnoreChecksum,
			&i.InstallOnUpgrade,
			&i.Force,
			&i.VerboseOutput,
			&i.NotSilent,
			&i.Timeout,
			&i.Status,
			&i.ExitCode,
			&i.Output,
			&i.Error,
			&i.AttemptedAt,
			&i.CompletedAt,
			&i.ExpiresAt,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
