UPDATE users SET is_active = $2
WHERE id = $1
RETURNING id, username, is_active;