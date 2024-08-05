-- Database: sweettooth

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

-- Each node will be categorized into organizations and 0 or more groups within each organization
CREATE TABLE IF NOT EXISTS organizations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- random UUID 
  name CITEXT NOT NULL, -- case-insensitive name of the organization
  UNIQUE(name) -- all organization names must be unique
);

-- Nodes represent individual machines that are connecting to the server
CREATE TABLE IF NOT EXISTS nodes (
  id UUID PRIMARY KEY, -- node ID should be the UUIDv5 of the public key
  organization_id UUID REFERENCES organizations(id) DEFAULT NULL, -- each node should be associated with exactly 1 organization.
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
  connected_on TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- when the node initially registered
  approved_on TIMESTAMP DEFAULT NULL, -- when the node was originally approvied
  last_seen TIMESTAMP DEFAULT NULL, -- when the node has most recently checked in
  approved BOOLEAN NOT NULL DEFAULT FALSE -- determines whether the device is approved or not
);

-- Groups will be statically composed of  
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
  id BIGSERIAL PRIMARY KEY,
  node_id UUID NOT NULL REFERENCES nodes(id),
  group_id UUID REFERENCES groups(id) DEFAULT NULL,
  organization_id UUID NOT NULL REFERENCES organizations(id),
  --
  action INTEGER NOT NULL, -- install, upgrade, uninstall
  name CITEXT NOT NULL, -- the chocolatey package name for any action
  version CITEXT DEFAULT NULL, -- the chocolatey package version (for install, upgrade)
  ignore_checksum BOOLEAN NOT NULL DEFAULT FALSE,
  install_on_upgrade BOOLEAN NOT NULL DEFAULT FALSE,
  force BOOLEAN NOT NULL DEFAULT FALSE,
  verbose_output BOOLEAN NOT NULL DEFAULT FALSE,
  not_silent BOOLEAN NOT NULL DEFAULT FALSE,
  --
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  completed_at TIMESTAMP DEFAULT NULL,
  status INTEGER DEFAULT 0,
  output TEXT DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS schedule_jobs(
  id BIGSERIAL PRIMARY KEY,
  node_id UUID NOT NULL REFERENCES nodes(id),
  group_id UUID REFERENCES groups(id) DEFAULT NULL,
  organization_id UUID NOT NULL REFERENCES organizations(id),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  completed_at TIMESTAMP DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS schedules (
  id SERIAL PRIMARY KEY,
  name CITEXT NOT NULL,
  organization_id UUID NOT NULL REFERENCES organizations(id),
  days CITEXT NOT NULL,
  start_time TIME NOT NULL,
  finish_time TIME NOT NULL,
  UNIQUE(organization_id, name)
);

CREATE TABLE IF NOT EXISTS node_schedule_assignments (
  node_id UUID NOT NULL REFERENCES nodes(id),
  schedule_id INTEGER NOT NULL REFERENCES schedules(id),
  organization_id UUID NOT NULL REFERENCES organizations(id),
  PRIMARY KEY(node_id, schedule_id, organization_id)
);

CREATE TABLE IF NOT EXISTS group_schedule_assignments (
  group_id UUID NOT NULL REFERENCES groups(id),
  schedule_id INTEGER NOT NULL REFERENCES schedules(id),
  organization_id UUID NOT NULL REFERENCES organizations(id),
  PRIMARY KEY(group_id, schedule_id, organization_id)
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