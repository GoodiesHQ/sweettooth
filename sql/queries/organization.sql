-- name: GetOrganizationByID :one
SELECT
    *
FROM
    organizations
WHERE
    id=$1;

-- name: CreateOrganization :one
INSERT INTO
organizations (
    name
)
VALUES
( 
    $1
) RETURNING *;
