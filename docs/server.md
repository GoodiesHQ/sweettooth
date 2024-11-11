# Server

The sweettooth server consists of 2 services:

1) The SweetTooth server executable which serves the application on HTTP(S)
2) A database that contains private but not sensitive information

## Configuration

SweetTooth Server is a binary that interacts with a PostgreSQL database. It is configured using environment variables to pass credentials used to connect to the database. To make the configuration easier, it re-uses several of the same environment variables as the postgres official image with additional ones:

| Variable | Default | Purpose |
| -------- | ------- | ------- |
| `POSTGRES_USER` | *required* | PostgreSQL username |
| `POSTGRES_PASSWORD` | *required* | PostgreSQL password |
| `POSTGRES_HOST` | `"localhost"` | PostgreSQL host |
| `POSTGRES_PORT` | `5432` | PostgreSQL port |
| `POSTGRES_DB` | `"sweettooth"` | PostgreSQL database name |
| `SWEEETTOOTH_SECRET` | *required, any string* | Secret used to sign web tokens |

## API

The API is broken into two components:
- **Node API**: handles connections from the sweettooth Windows client running on nodes
  - JWT Authorization token is signed by each node with a public key registered in the database
  - Simple JSON API:
    - Success: Returns the JSON object directly
    - Error: Sends an error `{"status": "error", "message": "..."}`
- **Web API**: handles requests from web browsers handling administration
  - JWT Authorization token is signed by the server, short lived expiration
  - Uses HTMX so HTTP endpoints send rendered HTML directly, no separate web client app
  - View system information:
    - OS release and build version info
    - All install and outdated Chocolatey packages
    - All installed software not managed by Chocolatey
    - The last check-in and maintenance schedules applied
  - Provision and view jobs to install, upgrade, or uninstall packages

### Node API

At the heart of SweetTooth is the database which contains all of the information needed for administrators to make decisions on package software management. The API allows the creation of package jobs and the modification of the maintenance schedules and Chocolatey sources of nodes or groups of nodes.

These are the endpoints used by nodes to interact with the server and database:

##### Unauthorized Endpoints
- **`/api/v1/node/register`**
Does not require an authorization header JWT token, but the registration form does require a signed value to prove private key ownership. In a perfect world, each public key/fingerprint would be verified by each node, but I realize that is not a likely scenario, so deploying behind HTTPS and having proper SSL health will help ensure that registration to the server is not hijacked. Approval procedures are your own.

##### Authorized Endpoints
All other endpoints require a signed JWT token. The JWT must be signed with the private key whos matching public key was registered and approved in the database. During authorization, the public key is verified to have been the originating signer and  the node ID is calculated, then checked in the database for validity and approval (or cache if present).


- **`GET /api/v1/node/check`**
Used as a check-in for the device to inform the server that it is currently online and able to communicate. This is performed periodically and a `Last Seen` value is updated on the server each check-in.

- **`GET /api/v1/node/schedule`**
Once authorized, the node ID is used to query the database for all applied maintenance schedules. These schedules can be applied to the node directly or to any groups that the node is a member of. It returns an array of schedules that the client can use to check its local time and determine if it is in a maintenance window.

- **`GET /api/v1/node/packages`**
Once authorized, the node ID is used to query the database for all packages installed on the system with or without chocolatey. Upon startup, the software tracker is empty. It needs to acquire the current inventory of software from the server's perspective so the client can determine any differences and report on them.

- **`PUT /api/v1/node/packages`**
Once authorized, the node ID is used to update the database entry for the node and replace the software inventory held by the server. This includes a 3 categories of packages: choco-managed, choco-unmanaged, choco-managed outdated. If any of these change, the software tracker will submit all 3 to the server. The server will not only update the node's software inventory with the new values, but it will also take the old values and the last time stamp and add those to a changelog which is a table containing a snapshot of the software inventory every client loop iteration. can be viewed or queried for reports and change history.

- **`GET /api/v1/node/packages/jobs?attempts_max=5`**
Once authorized, the node ID is used to query the database for all package-related jobs (install, upgrade, uninstall). It will only be queried during a maintenance schedule and returns only a list of Job UUIDs. Returning a list of IDs does not count as an attempt since the parameters of the job are not provided. Jobs in the database are returned that meet the following criteria:
  - Is assigned to the node ID which signed the token
  - Currently has a status of 0 (no result submitted)
  - Has fewer than `attempts_max` attempts (default: 5)
  - Is not expired or has an expiration of NULL

- **`GET /api/v1/node/packages/jobs/{JobID}`**
Once authorized, the node ID and provided job ID are used to update the database and increase the attempt count of the job by 1, update the `attempted_at` attribute to the current time, and return the current job parameters if it has not yet been completed. Jobs have a timeout, so if a job with `attempted_at` is greater than the current time minus the timeout, then there's a chance it is still running and has not yes succeeded or failed.

- **`POST /api/v1/node/packages/jobs/{JobID}`**
Once authorized, the node ID and provided job ID are used to update the database and set the job response values in the database signally the completion of the job whether it be success or failure. This includes the determined completion status (various kinds of success and failure determined by Choco command output) as well as the entire output of the command itself or any Go errors associated with the command.