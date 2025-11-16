SELECT u.id 
        FROM team_members tm
        JOIN teams t ON tm.team_id = t.id
        JOIN users u ON tm.user_id = u.id
        WHERE u.id != $1
          AND t.name = $2
          AND u.is_active = true
        LIMIT 2;