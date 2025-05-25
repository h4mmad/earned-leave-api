WITH can_insert AS (
    SELECT
        COUNT(*) FILTER (WHERE type = 'WORKED') AS worked,
        COUNT(*) FILTER (WHERE type = 'LEAVE')  AS leave,
        COUNT(*) FILTER (WHERE type = 'WORKED') -
        COUNT(*) FILTER (WHERE type = 'LEAVE')  AS balance
    FROM employee_entries
    WHERE "employeeId" = $1
),
inserted AS (
    INSERT INTO employee_entries (id, "employeeId", type, date)
    SELECT $2, $1, $3, $4
    FROM can_insert
    WHERE
        $3 = 'WORKED'
        OR ($3 = 'LEAVE' AND balance > 0)
    RETURNING id, "employeeId", type, date
),
stats AS (
    SELECT
        ci.worked + (CASE WHEN ii.type = 'WORKED' THEN 1 ELSE 0 END)   AS worked,
        ci.leave  + (CASE WHEN ii.type = 'LEAVE'  THEN 1 ELSE 0 END)   AS leave,
        ci.balance + (CASE WHEN ii.type = 'WORKED' THEN 1 ELSE -1 END) AS balance
    FROM inserted ii
    CROSS JOIN can_insert ci
)
SELECT
    ii.id,
    ii."employeeId",
    ii.type,
    ii.date,
    st.worked,
    st.leave,
    st.balance
FROM inserted ii
CROSS JOIN stats st;