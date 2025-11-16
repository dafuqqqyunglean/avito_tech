INSERT INTO users (id, username, is_active) 
VALUES ($1, $2, $3) 
ON CONFLICT (id) DO UPDATE SET 
username = EXCLUDED.username,
is_active = EXCLUDED.is_active;