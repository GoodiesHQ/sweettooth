package api

import "github.com/google/uuid"

// Node Registration
type RegistrationRequest struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	PublicKey      string     `json:"public_key"`
	Label          *string    `json:"label,omitempty"`
	Hostname       string     `json:"hostname"`
	OSKernel       string     `json:"os_kernel"`
	OSName         string     `json:"os_name"`
	OSMajor        int        `json:"os_major"`
	OSMinor        int        `json:"os_minor"`
	OSBuild        int        `json:"os_build"`
}

type GroupRequest struct {
	OrganizationID uuid.UUID
}
