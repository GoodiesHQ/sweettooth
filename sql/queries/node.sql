-- name: GetNodeByID :one
SELECT
    * -- get the full node by ID, try and keep this to a minimum as there can be a lot of data per node
FROM
    nodes
WHERE
    id=$1;

-- name: CreateNode :one
INSERT INTO
    nodes (
        id,
        organization_id,
        public_key,
        hostname,
        client_version,
        os_kernel, os_name, os_major, os_minor, os_build,
        packages_choco, packages_system, packages_outdated
    )
SELECT
    $1, -- id
    rt.organization_id, -- organization_id from registration_tokens
    $3, -- public_key
    $4, -- hostname
    $5, -- client_version
    $6, -- os_kernel
    $7, -- os_name
    $8, -- os_major
    $9, -- os_minor
    $10, -- os_build
    $11, -- packages_choco
    $12, -- packages_system
    $13 -- packages_outdated
FROM
    registration_tokens rt
WHERE
    rt.id = $2 -- registration token value
RETURNING *;

-- name: GetNodePackages :one
SELECT packages_choco, packages_system, packages_outdated FROM nodes WHERE id = $1;

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

-- name: UpdateNodePackages :exec
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
        id=(SELECT node_id FROM i) AND (packages_choco!=$2 OR packages_system!=$3 OR packages_outdated!=$4);

-- name: UpdateNodePackagesChoco :exec
UPDATE nodes SET packages_choco=$2 WHERE id=$1;

-- name: UpdateNodePackagesSystem :exec
UPDATE nodes SET packages_system=$2 WHERE id=$1;

-- name: UpdateNodePackagesOutdated :exec
UPDATE nodes SET packages_outdated=$2 WHERE id=$1;

---- name:UpdateNodePackages:exec
-- UPDATE nodes SET packages_choco=$2, packages_system=$3, packages_outdated=$4 WHERE nodes.id=$1;
