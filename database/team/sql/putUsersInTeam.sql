INSERT INTO team_members (team_id, user_id) 
VALUES ($1, $2) 
ON CONFLICT (team_id, user_id) DO NOTHING;