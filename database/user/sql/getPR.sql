SELECT pr.id, pr.name, pr.author_id, pr.status
FROM pull_requests pr
JOIN pr_reviewers rw ON pr.id = rw.pr_id
WHERE rw.user_id = $1;