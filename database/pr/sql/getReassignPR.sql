
SELECT pr.name, pr.author_id, pr.status, rv.user_id
FROM pull_requests pr
JOIN pr_reviewers rv ON pr.id = rv.pr_id
WHERE pr.id = $1;