UPDATE pull_requests
SET status = 'MERGED', merged_at = $2
WHERE id = $1;