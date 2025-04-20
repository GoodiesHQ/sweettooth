-- Database: sweettooth

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TYPE schedule_entry AS (
    rrule TEXT,
    time_beg SMALLINT,
    time_end SMALLINT
);

-- Users are administrators of the nodes
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- random UUID
  email CITEXT NOT NULL, -- case-insensitive email address
  password TEXT NOT NULL, -- hashed password
  mfatoken TEXT DEFAULT NULL, -- MFA secret token
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- when the user was created
  last_login TIMESTAMP DEFAULT NULL, -- when the user last logged in
  UNIQUE(email) -- all email addresses must be unique
);

-- Each node will be categorized into organizations and 0 or more groups within each organization
CREATE TABLE IF NOT EXISTS organizations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- random UUID 
  name CITEXT NOT NULL, -- case-insensitive name of the organization
  UNIQUE(name) -- all organization names must be unique
);

-- Each user can be assigned to organizations with various roles
CREATE TABLE IF NOT EXISTS user_organization_assignments (
  user_id UUID NOT NULL REFERENCES users(id), -- the user ID to add to the organization
  organization_id UUID NOT NULL REFERENCES organizations(id), -- the organization ID the user is joining
  role SMALLINT NOT NULL DEFAULT 0, -- the role of the user in the organization
  PRIMARY KEY (user_id, organization_id) -- each user can only be in an organization once
);

/*
  Roles:
    0 - Reader    = Can view nodes and jobs
    1 - Approver  = Reader + approve nodes
    2 - Operator  = Approver + run jobs
    3 - Manager   = Operator + set schedules, manage groups
    4 - Admin     = Manager + manage users
*/

-- Registration keys for an organization
CREATE TABLE IF NOT EXISTS registration_tokens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- the random ID used as the token
  organization_id UUID NOT NULL REFERENCES organizations(id), -- organization thi skey is active for
  name CITEXT NOT NULL, -- a name for the registration token
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- when this registration token was created
  expires_at TIMESTAMP DEFAULT NULL, -- when this registration token expires
  UNIQUE(organization_id, name) -- name should be unique within an organization
);

-- Nodes represent individual machines that are connecting to the server
CREATE TABLE IF NOT EXISTS nodes (
  id UUID PRIMARY KEY, -- node ID should be the UUIDv5 of the public key
  organization_id UUID NOT NULL REFERENCES organizations(id), -- each node should be associated with exactly 1 organization.
  public_key TEXT NOT NULL, -- node-generated ED25519 public key in base64 format
  label CITEXT DEFAULT NULL, -- admin-provided name that will be used in place of the hostname
  hostname CITEXT NOT NULL, -- the node's system hostname
  client_version CITEXT NOT NULL, -- the sweettooth client version
  pending_sources BOOLEAN NOT NULL DEFAULT FALSE, -- if true, client should update its sources
  pending_schedule BOOLEAN NOT NULL DEFAULT FALSE, -- if true, client should update its schedule
  os_kernel TEXT NOT NULL, -- the version of the windows NT kernel (e.g. 6.1)
  os_name TEXT NOT NULL, -- the OS distribution name (e.g. Windows 11 Enterprise)
  os_major INT NOT NULL, -- the OS major version
  os_minor INT NOT NULL, -- the OS minor version
  os_build INT NOT NULL, -- the OS build version
  packages_choco JSONB NOT NULL, -- SoftwareList managed by chocolatey
  packages_system JSONB NOT NULL, -- SoftwareList NOT managed by chocolatey
  packages_outdated JSONB NOT NULL, --  SoftwareOutdatedList managed by chocolatey
  packages_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- last time the packages were updated
  connected_on TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- when the node initially registered
  approved_on TIMESTAMP DEFAULT NULL, -- when the node was originally approvied
  last_seen TIMESTAMP DEFAULT NULL, -- when the node has most recently checked in
  approved BOOLEAN NOT NULL DEFAULT FALSE -- determines whether the device is approved or not
);

-- this is meant to keep a complete history of all package changes
CREATE TABLE IF NOT EXISTS node_package_changelog (
  id SERIAL PRIMARY KEY,
  node_id UUID NOT NULL REFERENCES nodes(id),
  organization_id UUID NOT NULL REFERENCES organizations(id),
  packages_choco JSONB NOT NULL,
  packages_system JSONB NOT NULL,
  packages_outdated JSONB NOT NULL,
  timestamp TIMESTAMP NOT NULL
);

-- Groups will be statically composed of nodes
CREATE TABLE IF NOT EXISTS groups (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- random UUID
  organization_id UUID NOT NULL REFERENCES organizations(id), -- each group exists within exactly one organization
  name CITEXT NOT NULL, -- name of the group, case-insensitive
  UNIQUE(organization_id, name) -- name must be unique within an organization
);

CREATE TABLE IF NOT EXISTS node_group_assignments (
  node_id UUID NOT NULL REFERENCES nodes(id), -- the node ID to add to the group
  group_id UUID NOT NULL REFERENCES groups(id), -- the group ID the node is joining
  organization_id UUID NOT NULL REFERENCES organizations(id), -- organization_id (redundant, should match node and group) included for sharding along org_id on all tables
  PRIMARY KEY (node_id, group_id, organization_id)
);

CREATE TABLE IF NOT EXISTS package_jobs(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- random job UUID
  node_id UUID NOT NULL REFERENCES nodes(id),
  group_id UUID REFERENCES groups(id) DEFAULT NULL,
  organization_id UUID NOT NULL REFERENCES organizations(id),
  attempts INTEGER NOT NULL DEFAULT 0, -- number of attempts made by the client to install
  -- PARAMETERS:
  action INTEGER NOT NULL, -- install, upgrade, uninstall
  name CITEXT NOT NULL, -- the chocolatey package name for any action
  version CITEXT DEFAULT NULL, -- the chocolatey package version (for install, upgrade)
  ignore_checksum BOOLEAN NOT NULL DEFAULT FALSE,
  install_on_upgrade BOOLEAN NOT NULL DEFAULT FALSE,
  force BOOLEAN NOT NULL DEFAULT FALSE,
  verbose_output BOOLEAN NOT NULL DEFAULT FALSE,
  not_silent BOOLEAN NOT NULL DEFAULT FALSE,
  timeout INTEGER NOT NULL DEFAULT 600, -- default of 10 minutes to perform an install/uninstall, best to set the timeout per job
  -- RESULT:
  status INTEGER NOT NULL DEFAULT 0,
  exit_code INTEGER DEFAULT NULL,
  output TEXT DEFAULT NULL,
  error TEXT DEFAULT NULL,
  -- metadata
  attempted_at TIMESTAMP DEFAULT NULL,
  completed_at TIMESTAMP DEFAULT NULL,
  expires_at TIMESTAMP DEFAULT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Schedules are iCal RRules along with start/end times
CREATE TABLE IF NOT EXISTS schedules (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- unique ID for each schedule
  organization_id UUID NOT NULL REFERENCES organizations(id), -- organization in which the schedule exists
  name CITEXT NOT NULL, -- the unique name of the schedule
  entries JSONB NOT NULL, -- a list of entries each containing an iCal RRule and start/stop times
  UNIQUE(organization_id, name) -- each schedule must have a unique name within the organization
);

-- Schedules assigned to individual nodes
CREATE TABLE IF NOT EXISTS node_schedule_assignments (
  schedule_id UUID NOT NULL REFERENCES schedules(id),
  node_id UUID NOT NULL REFERENCES nodes(id),
  organization_id UUID NOT NULL REFERENCES organizations(id),
  PRIMARY KEY(schedule_id, node_id, organization_id)
);

-- Schedules assigned to groups
CREATE TABLE IF NOT EXISTS group_schedule_assignments (
  schedule_id UUID NOT NULL REFERENCES schedules(id),
  group_id UUID NOT NULL REFERENCES groups(id),
  organization_id UUID NOT NULL REFERENCES organizations(id),
  PRIMARY KEY(schedule_id, group_id, organization_id)
);

-- Schedules assigned to the entire organization
CREATE TABLE IF NOT EXISTS organization_schedule_assignments (
  schedule_id UUID NOT NULL REFERENCES schedules(id),
  organization_id UUID NOT NULL REFERENCES organizations(id),
  PRIMARY KEY(schedule_id, organization_id)
);

-- Sources are iCal RRules along with start/end times
CREATE TABLE IF NOT EXISTS sources (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- unique ID for each source
  organization_id UUID NOT NULL REFERENCES organizations(id), -- organization in which the source exists
  name CITEXT NOT NULL, -- the unique name of the source
  entries JSONB NOT NULL, -- a list of entries each containing the chocolatey sources
  UNIQUE(organization_id, name) -- each source must have a unique name within the organization
);

-- Sources assigned to individual nodes
CREATE TABLE IF NOT EXISTS node_source_assignments (
  source_id UUID NOT NULL REFERENCES sources(id),
  node_id UUID NOT NULL REFERENCES nodes(id),
  organization_id UUID NOT NULL REFERENCES organizations(id),
  PRIMARY KEY(source_id, node_id, organization_id)
);

-- Sources assigned to groups
CREATE TABLE IF NOT EXISTS group_source_assignments (
  source_id UUID NOT NULL REFERENCES sources(id),
  group_id UUID NOT NULL REFERENCES groups(id),
  organization_id UUID NOT NULL REFERENCES organizations(id),
  PRIMARY KEY(source_id, group_id, organization_id)
);

-- Sources assigned to the entire organization
CREATE TABLE IF NOT EXISTS organization_source_assignments (
  source_id UUID NOT NULL REFERENCES sources(id),
  organization_id UUID NOT NULL REFERENCES organizations(id),
  PRIMARY KEY(source_id, organization_id)
);

/*
DROP TABLE IF EXISTS group_schedule_assignments;
DROP TABLE IF EXISTS node_schedule_assignments;
DROP TABLE IF EXISTS schedules;
DROP TABLE IF EXISTS nodes;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS node_group_assignments;
DROP TABLE IF EXISTS jobs;
*/