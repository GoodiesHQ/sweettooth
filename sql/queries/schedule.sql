/*Get a list of all schedules that apply to an node */

-- name: GetSchedulesByGroup :many
SELECT
    schedules.*
FROM
    schedules
JOIN
    group_schedule_assignments gsa on gsa.schedule_id = schedules.id
WHERE
    gsa.group_id=$1;


-- name: GetAllSchedulesByNode :many
SELECT DISTINCT
    schedules.*,
    'node' as entity_type,
    nsa.node_id as entity_id
FROM
    schedules
JOIN
    node_schedule_assignments nsa on nsa.schedule_id = schedules.id
WHERE
    nsa.node_id = $1
UNION
SELECT DISTINCT
    schedules.*,
    'group' as entity_type,
    gsa.group_id as entity_id
FROM
    schedules
JOIN
    group_schedule_assignments gsa ON schedules.id = gsa.schedule_id
JOIN
    groups grp ON gsa.group_id = grp.id
JOIN
    node_group_assignments nga ON grp.id = nga.group_id
WHERE
    nga.node_id = $1
UNION
SELECT DISTINCT
    schedules.*,
    'org' as entity_type,
    osa.organization_id as entity_id
FROM
    schedules
JOIN 
    organization_schedule_assignments osa ON schedules.id = osa.schedule_id
JOIN
    nodes on nodes.organization_id = osa.organization_id
WHERE
    nodes.id = $1;

-- name: GetCombinedScheduleByNode :many
SELECT DISTINCT
    schedules.*
FROM
    schedules
JOIN
    node_schedule_assignments nsa on nsa.schedule_id = schedules.id
WHERE
    nsa.node_id = $1
UNION
SELECT DISTINCT
    schedules.*
FROM
    schedules
JOIN
    group_schedule_assignments gsa ON schedules.id = gsa.schedule_id
JOIN
    groups grp ON gsa.group_id = grp.id
JOIN
    node_group_assignments nga ON grp.id = nga.group_id
WHERE
    nga.node_id = $1
UNION
SELECT DISTINCT
    schedules.*
FROM
    schedules
JOIN 
    organization_schedule_assignments osa ON schedules.id = osa.schedule_id
JOIN
    nodes on nodes.organization_id = osa.organization_id
WHERE
    nodes.id = $1;

-- -- name: GetSchedulesByNodeID :many
-- -- type: schedules
-- WITH node_schedules AS (
--     SELECT 
--         sched.*,
--         'node' as assignment
--     FROM
--         node_schedule_assignments nsa
--     JOIN
--         schedules sched ON nsa.schedule_id = sched.id
--     WHERE
--         nsa.node_id = $1
-- ), group_schedules AS (
--     SELECT 
--         sched.*,
--         'group' as assignment
--     FROM
--         node_group_assignments nga
--     JOIN 
--         group_schedule_assignments gsa ON nga.group_id = gsa.group_id
--     JOIN 
--         schedules sched ON gsa.schedule_id = sched.id
--     WHERE
--         nga.node_id = $1
-- 
-- ), organization_schedules AS (
--     SELECT
--         sched.*,
--         'organization' as assignment
--     FROM
--         organization_schedule_assignments osa
--     JOIN
--         schedules sched ON osa.id = sched.organization_id
--     WHERE
--         osa.node_id = $1
-- )
-- SELECT DISTINCT
--     *
-- FROM 
--     node_schedules
-- UNION
-- SELECT DISTINCT
--     *
-- FROM 
--     group_schedules
-- UNION
-- SELECT DISTINCT
--     *
-- FROM 
--     organization_schedules;
-- 

-- -- name: GetNodeSchedules :many
-- WITH node_schedules AS (
--     SELECT 
--         ns.schedule_id,
--         s.name,
--         s.days,
--         s.start_time,
--         s.finish_time
--     FROM 
--         node_schedule_assignments ns
--     JOIN 
--         schedules s ON ns.schedule_id = s.id
--     WHERE 
--         ns.node_id = $1
-- ), group_schedules AS (
--     SELECT 
--         gs.schedule_id,
--         s.name,
--         s.days,
--         s.start_time,
--         s.finish_time
--     FROM 
--         node_group_assignments nga
--     JOIN 
--         group_schedule_assignments gs ON nga.group_id = gs.group_id
--     JOIN 
--         schedules s ON gs.schedule_id = s.id
--     WHERE 
--         nga.node_id = $1
-- )
-- SELECT DISTINCT
--     *
-- FROM 
--     node_schedules
-- UNION
-- SELECT DISTINCT
--     *
-- FROM 
--     group_schedules;