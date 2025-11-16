INSERT INTO pull_requests (id, name, author_id, status)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO NOTHING;