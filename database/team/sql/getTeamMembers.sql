SELECT 
    u.id as user_internal_id,
    u.username,
    u.is_active
FROM teams t
JOIN team_members tm ON t.id = tm.team_id
JOIN users u ON tm.user_id = u.id
WHERE t.name = $1
ORDER BY u.id;