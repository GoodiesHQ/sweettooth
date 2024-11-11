package core_pgx

import (
	"github.com/goodieshq/sweettooth/internal/server/database"
	"github.com/goodieshq/sweettooth/pkg/api"
	"github.com/google/uuid"
)

// handles conversions between the standard and PGX implementations of objects

// convert a pgx org to api org
func pgxOrgToCoreOrg(dborg *database.Organization) *api.Organization {
	var org api.Organization
	org.ID = dborg.ID
	org.Name = dborg.Name
	return &org
}

// convert a pgx package job to api package job
func pgxPackageJobToCorePackageJob(dbjob *database.PackageJob) *api.PackageJob {
	var job api.PackageJob

	// base values of the job
	job.ID = dbjob.ID
	job.NodeID = dbjob.NodeID
	if dbjob.GroupID.Valid {
		gid := uuid.UUID(dbjob.GroupID.Bytes)
		job.GroupID = &gid
	}
	job.OrganizationID = dbjob.OrganizationID
	job.Attempts = int(dbjob.Attempts)

	job.Action = int(dbjob.Action)

	// set the parameters
	job.Parameters.Name = dbjob.Name
	if dbjob.Version.Valid {
		job.Parameters.Version = &dbjob.Version.String
	}
	job.Parameters.IgnoreChecksum = dbjob.IgnoreChecksum
	job.Parameters.InstallOnUpgrade = dbjob.InstallOnUpgrade
	job.Parameters.Force = dbjob.Force
	job.Parameters.VerboseOutput = dbjob.VerboseOutput
	job.Parameters.NotSilent = dbjob.NotSilent

	// set the result
	job.Result = &api.PackageJobResult{}
	job.Result.Status = int(dbjob.Status)
	if dbjob.Output.Valid {
		job.Result.Output = dbjob.Output.String
	}

	// final metadata
	job.CreatedAt = dbjob.CreatedAt.Time
	if dbjob.ExpiresAt.Valid {
		job.ExpiresAt = &dbjob.ExpiresAt.Time
	}
	if dbjob.AttemptedAt.Valid {
		job.AttemptedAt = &dbjob.AttemptedAt.Time
	}
	if dbjob.CompletedAt.Valid {
		job.CompletedAt = &dbjob.CompletedAt.Time
	}

	return &job
}

// convert a pgx node to api node
func pgxNodeToCoreNode(dbnode *database.Node) *api.Node {
	var node api.Node

	node.ID = dbnode.ID
	node.OrganizationID = &dbnode.OrganizationID
	node.PublicKey = dbnode.PublicKey

	if dbnode.Label.Valid {
		node.Label = &dbnode.Label.String
	}

	node.Hostname = dbnode.Hostname
	node.ClientVersion = dbnode.ClientVersion
	node.PendingSources = dbnode.PendingSources
	node.PendingSchedule = dbnode.PendingSchedule
	node.OSKernel = dbnode.OsKernel
	node.OSName = dbnode.OsName
	node.OSMajor = int(dbnode.OsMajor)
	node.OSMinor = int(dbnode.OsMinor)
	node.OSBuild = int(dbnode.OsBuild)

	if dbnode.ConnectedOn.Valid {
		node.ConnectedOn = dbnode.ConnectedOn.Time
	}
	if dbnode.ApprovedOn.Valid {
		node.ApprovedOn = &dbnode.ApprovedOn.Time
	}
	if dbnode.LastSeen.Valid {
		node.LastSeen = &dbnode.LastSeen.Time
	}

	node.Approved = dbnode.Approved

	return &node
}
