-- name: GetOrganizationByID :one
SELECT
    *
FROM
    organizations
WHERE
    id=$1;


-- name: GetOrganizations :many
SELECT
    *
FROM
    organizations
ORDER BY name ASC;

-- name: GetOrganizationSummaries :many
SELECT
  o.*,
  COUNT(n.id) AS node_count
FROM organizations o
LEFT JOIN nodes n ON n.organization_id = o.id
GROUP BY o.id ORDER BY o.name ASC;

-- name: GetValidRegistrationToken :one
SELECT
    organization_id
FROM
    registration_tokens
WHERE
    id=$1 AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP);


-- name: GetOrganizationIDFromRegistrationToken :one
SELECT
    organization_id
FROM
    registration_tokens
WHERE
    id=$1;


-- name: CreateOrganization :one
INSERT INTO
organizations (
    name
) VALUES ( 
    $1
) RETURNING *;
