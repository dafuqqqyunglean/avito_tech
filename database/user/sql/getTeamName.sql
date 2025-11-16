SELECT t.name 
FROM teams t
JOIN team_members tm ON t.id = tm.team_id
WHERE tm.user_id = $1;