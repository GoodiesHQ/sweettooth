package api

import (
	"fmt"
	"time"

	"github.com/goodieshq/sweettooth/internal/schedule"
	"github.com/goodieshq/sweettooth/internal/util"
	"github.com/google/uuid"
)

// Node Registration
type RegistrationRequest struct {
	Token            uuid.UUID                 `json:"token"` // registry token to validate
	ClientVersion    string                    `json:"client_version"`
	PublicKey        string                    `json:"public_key"`
	PublicKeySig     string                    `json:"public_key_sig"` // the resulting base64-encoded signature of using the private key to sign the public key's raw bytes
	Label            *string                   `json:"label,omitempty"`
	Hostname         string                    `json:"hostname"`
	OSKernel         string                    `json:"os_kernel"`
	OSName           string                    `json:"os_name"`
	OSMajor          int                       `json:"os_major"`
	OSMinor          int                       `json:"os_minor"`
	OSBuild          int                       `json:"os_build"`
	PackagesChoco    util.SoftwareList         `json:"packages_choco"`
	PackagesSystem   util.SoftwareList         `json:"packages_system"`
	PackagesOutdated util.SoftwareOutdatedList `json:"packages_outdated"`
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

type ErrorResponse struct {
	Message    string `json:"message"`               // error message
	Status     string `json:"status"`                // error status
	StatusCode int    `json:"status_code,omitempty"` // error status code
}

func (err ErrorResponse) Error() string {
	return fmt.Sprintf("(%d) %s", err.StatusCode, err.Message)
}

type Schedule schedule.Schedule

type Organization struct {
	ID   uuid.UUID `json:"id"`   // random org ID
	Name string    `json:"name"` // org name (unique, case-insensitive)
}

type OrganizationSummary struct {
	Organization
	NodeCount int `json:"node_count"` // number of nodes in this org
}

type Node struct {
	ID              uuid.UUID  `json:"id"`               // node's ID (determined by public key)
	OrganizationID  *uuid.UUID `json:"organization_id"`  // node's org ID
	PublicKey       string     `json:"public_key"`       // node's public key
	Label           *string    `json:"label"`            // optional label to overwrite hostname
	Hostname        string     `json:"hostname"`         // the system's hostname
	ClientVersion   string     `json:"client_version"`   // sweetooth client version
	PendingSources  bool       `json:"pending_sources"`  // determines if there are source changes for this node
	PendingSchedule bool       `json:"pending_schedule"` // determines if there are schedule changes for this node
	OSKernel        string     `json:"os_kernel"`        // OS kernel version
	OSName          string     `json:"os_name"`          // OS name (e.g. "Windows 11")
	OSMajor         int        `json:"os_major"`         // OS major version
	OSMinor         int        `json:"os_minor"`         // OS minor version
	OSBuild         int        `json:"os_build"`         // OS build version
	ConnectedOn     time.Time  `json:"connected_on"`     // when the node first connected
	ApprovedOn      *time.Time `json:"approved_on"`      // when the node was approved
	LastSeen        *time.Time `json:"last_seen"`        // when the node last checked in
	Approved        bool       `json:"approved"`         // whether the node has been approved or not
}

type Packages struct {
	PackagesChoco    util.SoftwareList         `json:"packages_choco"`    // list of packages on the node managed by chocolatey
	PackagesSystem   util.SoftwareList         `json:"packages_system"`   // list of packages on the node NOT managed by chocolatey
	PackagesOutdated util.SoftwareOutdatedList `json:"packages_outdated"` // list of outdated packages on the node managed by chocolatey
}

type PackageJobParameters struct {
	Name             string  `json:"name"`               // target package name
	Version          *string `json:"string,omitempty"`   // target package version (optional)
	Timeout          int     `json:"timeout"`            // timeout for the command (sans grace period)
	IgnoreChecksum   bool    `json:"ignore_checksum"`    // ignore package checksum
	InstallOnUpgrade bool    `json:"install_on_upgrade"` // install if missing (upgrade action only)
	Force            bool    `json:"force"`              // force the action
	VerboseOutput    bool    `json:"verbose_output"`     // verbose output
	NotSilent        bool    `json:"bool"`               // disable silent install
}

type PackageJobResult struct {
	Status   int     `json:"status"`    // the choco status of the result (sweettooth specific)
	ExitCode int     `json:"exit_code"` // the choco process exit code
	Output   string  `json:"output"`    // the chocolatey command output
	Error    *string `json:"error"`
}

type PackageJob struct {
	ID             uuid.UUID            `json:"id"`                 // random package job ID
	NodeID         uuid.UUID            `json:"node_id"`            // target node ID for this job
	GroupID        *uuid.UUID           `json:"group_id,omitempty"` // the group ID the job was assigned to (if applicable)
	OrganizationID uuid.UUID            `json:"organization_id"`    // the organization this job/node is associated with
	Attempts       int                  `json:"attempts"`           // the number of attempts this task has undergone
	Action         int                  `json:"action"`             // the action that should be performed
	Parameters     PackageJobParameters `json:"parameters"`         // the parameters passed to chocolatey
	CreatedAt      time.Time            `json:"created_at"`         // when the task was created
	ExpiresAt      *time.Time           `json:"expires_at"`         // when the task expires
	AttemptedAt    *time.Time           `json:"attempted_at"`       // when the task was last attempted
	CompletedAt    *time.Time           `json:"completed_at"`       // when the task was completed
	Result         *PackageJobResult    `json:"result"`             // the result of the task if completed
}

type PackageJobList []uuid.UUID // a node receives a list of job IDs instead of the entire job
