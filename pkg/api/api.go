package api

import (
	"fmt"
	"time"

	"github.com/goodieshq/sweettooth/pkg/schedule"
	"github.com/goodieshq/sweettooth/pkg/util"
	"github.com/google/uuid"
)

// Node Registration
type RegistrationRequest struct {
	// ID             uuid.UUID       `json:"id"`
	OrganizationID *uuid.UUID        `json:"organization_id,omitempty"`
	ClientVersion  string            `json:"client_version"`
	PublicKey      string            `json:"public_key"`
	PublicKeySig   string            `json:"public_key_sig"` // the resulting base64-encoded signature of using the private key to sign the public key's raw bytes
	Label          *string           `json:"label,omitempty"`
	Hostname       string            `json:"hostname"`
	OSKernel       string            `json:"os_kernel"`
	OSName         string            `json:"os_name"`
	OSMajor        int               `json:"os_major"`
	OSMinor        int               `json:"os_minor"`
	OSBuild        int               `json:"os_build"`
	PackagesChoco  util.SoftwareList `json:"packages_choco"`
	PackagesSystem util.SoftwareList `json:"packages_system"`
}

type SoftwareInventoryRequest struct {
	PackagesChoco  util.SoftwareList `json:"packages_choco"`
	PackagesSystem util.SoftwareList `json:"packages_system"`
}

type CheckResponse struct {
	PendingSources  bool `json:"pending_sources"`  // client should update its sources
	PendingSchedule bool `json:"pending_schedule"` // client should update its schedule
	PendingJobs     bool `json:"pending_jobs"`     // client should perform pending jobs (schedule permitting)
}

type GroupRequest struct {
	OrganizationID uuid.UUID
}

type ErrorResponse struct {
	Status     string `json:"status"`
	StatusCode int    `json:"status_code,omitempty"`
	Message    string `json:"message"`
}

func (err ErrorResponse) Error() string {
	return fmt.Sprintf("(%d) %s", err.StatusCode, err.Message)
}

type Schedule schedule.Schedule

type Organization struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type Node struct {
	ID              uuid.UUID  `json:"id"`
	OrganizationID  *uuid.UUID `json:"organization_id"`
	PublicKey       string     `json:"public_key"`
	Label           *string    `json:"label"`
	Hostname        string     `json:"hostname"`
	ClientVersion   string     `json:"client_version"`
	PendingSources  bool       `json:"pending_sources"`
	PendingSchedule bool       `json:"pending_schedule"`
	OSKernel        string     `json:"os_kernel"`
	OSName          string     `json:"os_name"`
	OSMajor         int        `json:"os_major"`
	OSMinor         int        `json:"os_minor"`
	OSBuild         int        `json:"os_build"`
	ConnectedOn     time.Time  `json:"connected_on"`
	ApprovedOn      *time.Time `json:"approved_on"`
	LastSeen        *time.Time `json:"last_seen"`
	Approved        bool       `json:"approved"`
}

type Packages struct {
	PackagesChoco    util.SoftwareList         `json:"packages_choco,omitempty"`
	PackagesSystem   util.SoftwareList         `json:"packages_system,omitempty"`
	PackagesOutdated util.SoftwareOutdatedList `json:"packages_outdated,omitempty"`
}
