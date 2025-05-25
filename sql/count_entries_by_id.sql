SELECT json_build_object(
  'WORKED', COUNT(*) FILTER (WHERE type = 'WORKED'),
  'LEAVE',  COUNT(*) FILTER (WHERE type = 'LEAVE')
) AS result
FROM employee_entries
WHERE "employeeId" = '1';
