<div align="center"><img src="logo/logo.jpg" width="500"/></div>

### A self-hosted Chocolatey wrapper providing the ability to inventory, manage, and update 3rd party software using Chocolatey over a large number of machines.

### Design


/api/v1/node/check
 - 200 = Node exists and is enabled
 - 401 = JWT token is invalid or expired
 - 403 = Node exists but has not been approved
 - 404 = Node does not exist

/api/v1/node/register
 - 401 = JWT token is invalid or expired
 - 201 = Created a new user based on the provided information
 - 409 = Node already exists

/api/v1/node/schedule
 - 401 = JWT token is invalid or expired
 - 403 = Node is disabled
 - 404 = Node is not registered
 - 200 = Returns the complete schedule for the node

/api/v1/node/sources
 - 401 = JWT token is invalid or expired
 - 403 = Node is disabled
 - 404 = Node is not registered
 - 200 = Returns the complete list of sources and statuses for the node

/api/v1/node/pending
 - 401 = JWT token is invalid or expired
 - 403 = Node is disabled
 - 404 = Node is not registered
 - 200 = Determines if there are any pending updates to the sources or schedules

/api/v1/node/jobs
 - 401 = JWT token is invalid or expired
 - 403 = Node is disabled
 - 404 = Node is not registered
 - 200 = Returns the complete list of pending package jobs, waits to perform them per the schedule

#### `/choco`: Chocolatey wrapper

- Bootstrap (install chocolatey with PowerShell)
- 