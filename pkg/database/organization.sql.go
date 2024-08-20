// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: organization.sql

package database

import (
	"context"

	"github.com/google/uuid"
)

const createOrganization = `-- name: CreateOrganization :one
INSERT INTO
organizations (
    name
) VALUES ( 
    $1
) RETURNING id, name
`

func (q *Queries) CreateOrganization(ctx context.Context, name string) (Organization, error) {
	row := q.db.QueryRow(ctx, createOrganization, name)
	var i Organization
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const getOrganizationByID = `-- name: GetOrganizationByID :one
SELECT
    id, name
FROM
    organizations
WHERE
    id=$1
`

func (q *Queries) GetOrganizationByID(ctx context.Context, id uuid.UUID) (Organization, error) {
	row := q.db.QueryRow(ctx, getOrganizationByID, id)
	var i Organization
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const getOrganizationIDFromRegistrationToken = `-- name: GetOrganizationIDFromRegistrationToken :one
SELECT
    organization_id
FROM
    registration_tokens
WHERE
    id=$1
`

func (q *Queries) GetOrganizationIDFromRegistrationToken(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	row := q.db.QueryRow(ctx, getOrganizationIDFromRegistrationToken, id)
	var organization_id uuid.UUID
	err := row.Scan(&organization_id)
	return organization_id, err
}

const getValidRegistrationToken = `-- name: GetValidRegistrationToken :one
SELECT
    organization_id
FROM
    registration_tokens
WHERE
    id=$1 AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
`

func (q *Queries) GetValidRegistrationToken(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	row := q.db.QueryRow(ctx, getValidRegistrationToken, id)
	var organization_id uuid.UUID
	err := row.Scan(&organization_id)
	return organization_id, err
}
