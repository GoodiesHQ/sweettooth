-- name: GetNodeByID :one
SELECT
    *
FROM
    nodes
WHERE
    id=$1;

-- name: CreateNode :exec
INSERT INTO
    nodes (
        id,
        organization_id,
        public_key,
        hostname,
        client_version,
        os_kernel, os_name, os_major, os_minor, os_build,
        packages_choco, packages_system
    )
VALUES
    ( $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: SetNodeApproval :exec
UPDATE
    nodes
SET
    approved=$2
WHERE
    id=$1;

-- name: CheckInNode :exec
UPDATE
    nodes
SET
    last_seen=CURRENT_TIMESTAMP
WHERE
    id=$1;

-- name: UpdatePackages :exec
WITH
    n AS (
        SELECT
            id, organization_id, packages_choco, packages_system, packages_outdated, packages_updated_at
        FROM
            nodes
        WHERE
            nodes.id=$1
        FOR UPDATE
    ),
    i AS (
        INSERT INTO
            node_package_changelog(
                node_id,
                organization_id,
                packages_choco,
                packages_system,
                packages_outdated,
                timestamp
            )
        SELECT
            id, organization_id, packages_choco, packages_system, packages_outdated, packages_updated_at
        FROM
            n
        RETURNING *
    )
    UPDATE
        nodes
    SET
        packages_updated_at=CURRENT_TIMESTAMP, packages_choco=$2, packages_system=$3, packages_outdated=$4
    WHERE
        id=(SELECT node_id FROM i);

-- name: UpdateNodePackagesChoco :exec
UPDATE nodes SET packages_choco=$2 WHERE id=$1;

-- name: UpdateNodePackagesSystem :exec
UPDATE nodes SET packages_system=$2 WHERE id=$1;

-- name: UpdateNodePackagesOutdated :exec
UPDATE nodes SET packages_outdated=$2 WHERE id=$1;

-- name: UpdateNodePackages :exec
UPDATE nodes SET packages_choco=$2, packages_system=$3, packages_outdated=$4 WHERE nodes.id=$1;
