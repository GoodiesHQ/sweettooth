// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: schedule.sql

package database

import (
	"context"

	"github.com/goodieshq/sweettooth/internal/schedule"
	"github.com/google/uuid"
)

const getAllSchedulesByNode = `-- name: GetAllSchedulesByNode :many
SELECT DISTINCT
    schedules.id, schedules.organization_id, schedules.name, schedules.entries,
    'node' as entity_type,
    nsa.node_id as entity_id
FROM
    schedules
JOIN
    node_schedule_assignments nsa on nsa.schedule_id = schedules.id
WHERE
    nsa.node_id = $1
UNION
SELECT DISTINCT
    schedules.id, schedules.organization_id, schedules.name, schedules.entries,
    'group' as entity_type,
    gsa.group_id as entity_id
FROM
    schedules
JOIN
    group_schedule_assignments gsa ON schedules.id = gsa.schedule_id
JOIN
    groups grp ON gsa.group_id = grp.id
JOIN
    node_group_assignments nga ON grp.id = nga.group_id
WHERE
    nga.node_id = $1
UNION
SELECT DISTINCT
    schedules.id, schedules.organization_id, schedules.name, schedules.entries,
    'org' as entity_type,
    osa.organization_id as entity_id
FROM
    schedules
JOIN 
    organization_schedule_assignments osa ON schedules.id = osa.schedule_id
JOIN
    nodes on nodes.organization_id = osa.organization_id
WHERE
    nodes.id = $1
`

type GetAllSchedulesByNodeRow struct {
	ID             uuid.UUID         `db:"id" json:"id"`
	OrganizationID uuid.UUID         `db:"organization_id" json:"organization_id"`
	Name           string            `db:"name" json:"name"`
	Entries        schedule.Schedule `db:"entries" json:"entries"`
	EntityType     string            `db:"entity_type" json:"entity_type"`
	EntityID       uuid.UUID         `db:"entity_id" json:"entity_id"`
}

func (q *Queries) GetAllSchedulesByNode(ctx context.Context, nodeID uuid.UUID) ([]GetAllSchedulesByNodeRow, error) {
	rows, err := q.db.Query(ctx, getAllSchedulesByNode, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAllSchedulesByNodeRow
	for rows.Next() {
		var i GetAllSchedulesByNodeRow
		if err := rows.Scan(
			&i.ID,
			&i.OrganizationID,
			&i.Name,
			&i.Entries,
			&i.EntityType,
			&i.EntityID,
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

const getCombinedScheduleByNode = `-- name: GetCombinedScheduleByNode :many
SELECT DISTINCT
    schedules.id, schedules.organization_id, schedules.name, schedules.entries
FROM
    schedules
JOIN
    node_schedule_assignments nsa on nsa.schedule_id = schedules.id
WHERE
    nsa.node_id = $1
UNION
SELECT DISTINCT
    schedules.id, schedules.organization_id, schedules.name, schedules.entries
FROM
    schedules
JOIN
    group_schedule_assignments gsa ON schedules.id = gsa.schedule_id
JOIN
    groups grp ON gsa.group_id = grp.id
JOIN
    node_group_assignments nga ON grp.id = nga.group_id
WHERE
    nga.node_id = $1
UNION
SELECT DISTINCT
    schedules.id, schedules.organization_id, schedules.name, schedules.entries
FROM
    schedules
JOIN 
    organization_schedule_assignments osa ON schedules.id = osa.schedule_id
JOIN
    nodes on nodes.organization_id = osa.organization_id
WHERE
    nodes.id = $1
`

func (q *Queries) GetCombinedScheduleByNode(ctx context.Context, nodeID uuid.UUID) ([]Schedule, error) {
	rows, err := q.db.Query(ctx, getCombinedScheduleByNode, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Schedule
	for rows.Next() {
		var i Schedule
		if err := rows.Scan(
			&i.ID,
			&i.OrganizationID,
			&i.Name,
			&i.Entries,
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

const getSchedulesByGroup = `-- name: GetSchedulesByGroup :many

SELECT
    schedules.id, schedules.organization_id, schedules.name, schedules.entries
FROM
    schedules
JOIN
    group_schedule_assignments gsa on gsa.schedule_id = schedules.id
WHERE
    gsa.group_id=$1
`

// Get a list of all schedules that apply to an node
func (q *Queries) GetSchedulesByGroup(ctx context.Context, groupID uuid.UUID) ([]Schedule, error) {
	rows, err := q.db.Query(ctx, getSchedulesByGroup, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Schedule
	for rows.Next() {
		var i Schedule
		if err := rows.Scan(
			&i.ID,
			&i.OrganizationID,
			&i.Name,
			&i.Entries,
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