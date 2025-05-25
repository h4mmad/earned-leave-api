 SELECT
   EXTRACT(YEAR  FROM date)         AS year,
   EXTRACT(MONTH FROM date)         AS month,
   EXTRACT(DAY FROM date)           AS day,
   type,
   entry_id
FROM entries
WHERE employee_id = $1;