package server

import (
	"time"

	"github.com/google/uuid"
)

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

type Core interface {
	Close()
	GetNodeByID(id uuid.UUID) (*Node, error)
}
