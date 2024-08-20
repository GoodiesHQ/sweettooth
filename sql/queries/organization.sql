-- name: GetOrganizationByID :one
SELECT
    *
FROM
    organizations
WHERE
    id=$1;


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
