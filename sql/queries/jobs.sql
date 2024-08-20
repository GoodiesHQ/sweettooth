-- name: GetPackageJobsByNodeID :many
SELECT
    *
FROM
    package_jobs
WHERE
    node_id=$1
ORDER BY created_at ASC;

-- name: GetPackageJobListByNodeID :many
SELECT
    id
FROM
    package_jobs
WHERE
    node_id=$1 AND status = 0 AND (attempts < $2 OR $2 = 0)
ORDER BY created_at ASC;

-- name: GetPackageJobByID :one
SELECT
    *
FROM
    package_jobs
WHERE
    id=$1
LIMIT 1;


-- name: AttemptPackageJob :one
UPDATE
    package_jobs
SET
    attempts = attempts + 1
WHERE
    id = $1 AND node_id=$2 AND attempts < $3
RETURNING *;

-- name: CompletePackageJob :one
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
RETURNING *;

-- name: CreatePackageJob :one
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
RETURNING *;