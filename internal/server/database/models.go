// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package database

import (
	"github.com/goodieshq/sweettooth/internal/schedule"
	"github.com/goodieshq/sweettooth/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Group struct {
	ID             uuid.UUID `db:"id" json:"id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`
	Name           string    `db:"name" json:"name"`
}

type GroupScheduleAssignment struct {
	ScheduleID     uuid.UUID `db:"schedule_id" json:"schedule_id"`
	GroupID        uuid.UUID `db:"group_id" json:"group_id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`
}

type GroupSourceAssignment struct {
	SourceID       uuid.UUID `db:"source_id" json:"source_id"`
	GroupID        uuid.UUID `db:"group_id" json:"group_id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`
}

type Node struct {
	ID                uuid.UUID                 `db:"id" json:"id"`
	OrganizationID    uuid.UUID                 `db:"organization_id" json:"organization_id"`
	PublicKey         string                    `db:"public_key" json:"public_key"`
	Label             pgtype.Text               `db:"label" json:"label"`
	Hostname          string                    `db:"hostname" json:"hostname"`
	ClientVersion     string                    `db:"client_version" json:"client_version"`
	PendingSources    bool                      `db:"pending_sources" json:"pending_sources"`
	PendingSchedule   bool                      `db:"pending_schedule" json:"pending_schedule"`
	OsKernel          string                    `db:"os_kernel" json:"os_kernel"`
	OsName            string                    `db:"os_name" json:"os_name"`
	OsMajor           int32                     `db:"os_major" json:"os_major"`
	OsMinor           int32                     `db:"os_minor" json:"os_minor"`
	OsBuild           int32                     `db:"os_build" json:"os_build"`
	PackagesChoco     util.SoftwareList         `db:"packages_choco" json:"packages_choco"`
	PackagesSystem    util.SoftwareList         `db:"packages_system" json:"packages_system"`
	PackagesOutdated  util.SoftwareOutdatedList `db:"packages_outdated" json:"packages_outdated"`
	PackagesUpdatedAt pgtype.Timestamp          `db:"packages_updated_at" json:"packages_updated_at"`
	ConnectedOn       pgtype.Timestamp          `db:"connected_on" json:"connected_on"`
	ApprovedOn        pgtype.Timestamp          `db:"approved_on" json:"approved_on"`
	LastSeen          pgtype.Timestamp          `db:"last_seen" json:"last_seen"`
	Approved          bool                      `db:"approved" json:"approved"`
}

type NodeGroupAssignment struct {
	NodeID         uuid.UUID `db:"node_id" json:"node_id"`
	GroupID        uuid.UUID `db:"group_id" json:"group_id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`
}

type NodePackageChangelog struct {
	ID               int32                     `db:"id" json:"id"`
	NodeID           uuid.UUID                 `db:"node_id" json:"node_id"`
	OrganizationID   uuid.UUID                 `db:"organization_id" json:"organization_id"`
	PackagesChoco    util.SoftwareList         `db:"packages_choco" json:"packages_choco"`
	PackagesSystem   util.SoftwareList         `db:"packages_system" json:"packages_system"`
	PackagesOutdated util.SoftwareOutdatedList `db:"packages_outdated" json:"packages_outdated"`
	Timestamp        pgtype.Timestamp          `db:"timestamp" json:"timestamp"`
}

type NodeScheduleAssignment struct {
	ScheduleID     uuid.UUID `db:"schedule_id" json:"schedule_id"`
	NodeID         uuid.UUID `db:"node_id" json:"node_id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`
}

type NodeSourceAssignment struct {
	SourceID       uuid.UUID `db:"source_id" json:"source_id"`
	NodeID         uuid.UUID `db:"node_id" json:"node_id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`
}

type Organization struct {
	ID   uuid.UUID `db:"id" json:"id"`
	Name string    `db:"name" json:"name"`
}

type OrganizationScheduleAssignment struct {
	ScheduleID     uuid.UUID `db:"schedule_id" json:"schedule_id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`
}

type OrganizationSourceAssignment struct {
	SourceID       uuid.UUID `db:"source_id" json:"source_id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`
}

type PackageJob struct {
	ID               uuid.UUID        `db:"id" json:"id"`
	NodeID           uuid.UUID        `db:"node_id" json:"node_id"`
	GroupID          pgtype.UUID      `db:"group_id" json:"group_id"`
	OrganizationID   uuid.UUID        `db:"organization_id" json:"organization_id"`
	Attempts         int32            `db:"attempts" json:"attempts"`
	Action           int32            `db:"action" json:"action"`
	Name             string           `db:"name" json:"name"`
	Version          pgtype.Text      `db:"version" json:"version"`
	IgnoreChecksum   bool             `db:"ignore_checksum" json:"ignore_checksum"`
	InstallOnUpgrade bool             `db:"install_on_upgrade" json:"install_on_upgrade"`
	Force            bool             `db:"force" json:"force"`
	VerboseOutput    bool             `db:"verbose_output" json:"verbose_output"`
	NotSilent        bool             `db:"not_silent" json:"not_silent"`
	Timeout          int32            `db:"timeout" json:"timeout"`
	Status           int32            `db:"status" json:"status"`
	ExitCode         pgtype.Int4      `db:"exit_code" json:"exit_code"`
	Output           pgtype.Text      `db:"output" json:"output"`
	Error            pgtype.Text      `db:"error" json:"error"`
	AttemptedAt      pgtype.Timestamp `db:"attempted_at" json:"attempted_at"`
	CompletedAt      pgtype.Timestamp `db:"completed_at" json:"completed_at"`
	ExpiresAt        pgtype.Timestamp `db:"expires_at" json:"expires_at"`
	CreatedAt        pgtype.Timestamp `db:"created_at" json:"created_at"`
}

type RegistrationToken struct {
	ID             uuid.UUID        `db:"id" json:"id"`
	OrganizationID uuid.UUID        `db:"organization_id" json:"organization_id"`
	Name           string           `db:"name" json:"name"`
	CreatedAt      pgtype.Timestamp `db:"created_at" json:"created_at"`
	ExpiresAt      pgtype.Timestamp `db:"expires_at" json:"expires_at"`
}

type Schedule struct {
	ID             uuid.UUID         `db:"id" json:"id"`
	OrganizationID uuid.UUID         `db:"organization_id" json:"organization_id"`
	Name           string            `db:"name" json:"name"`
	Entries        schedule.Schedule `db:"entries" json:"entries"`
}

type Source struct {
	ID             uuid.UUID `db:"id" json:"id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`
	Name           string    `db:"name" json:"name"`
	Entries        []byte    `db:"entries" json:"entries"`
}
