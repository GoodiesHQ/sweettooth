-- name: GetAllOrganizations :many
SELECT * FROM organizations;

-- name: GetNodeByID :one
SELECT * FROM nodes WHERE id=$1 LIMIT 1;