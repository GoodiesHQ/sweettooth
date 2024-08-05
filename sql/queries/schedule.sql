/*Get a list of all schedules that apply to an node */
-- name: GetNodeSchedules :many
WITH node_schedules AS (
    SELECT 
        ns.schedule_id,
        s.name,
        s.days,
        s.start_time,
        s.finish_time
    FROM 
        node_schedule_assignments ns
    JOIN 
        schedules s ON ns.schedule_id = s.id
    WHERE 
        ns.node_id = $1
), group_schedules AS (
    SELECT 
        gs.schedule_id,
        s.name,
        s.days,
        s.start_time,
        s.finish_time
    FROM 
        node_group_assignments nga
    JOIN 
        group_schedule_assignments gs ON nga.group_id = gs.group_id
    JOIN 
        schedules s ON gs.schedule_id = s.id
    WHERE 
        nga.node_id = $1
)
SELECT DISTINCT
    *
FROM 
    node_schedules
UNION
SELECT DISTINCT
    *
FROM 
    group_schedules;