UPDATE pr_reviewers 
SET user_id = sub.new_user
FROM (
    SELECT u.id AS new_user
    FROM users u
    JOIN team_members tm ON u.id = tm.user_id
    JOIN team_members old_tm ON tm.team_id = old_tm.team_id
    WHERE old_tm.user_id = $2
      AND u.id != $2
      AND u.is_active = true
      AND u.id NOT IN (
          SELECT user_id 
          FROM pr_reviewers 
          WHERE pr_id = $1
      )
    LIMIT 1
) AS sub
WHERE pr_id = $1 
  AND user_id = $2
  AND sub.new_user IS NOT NULL
RETURNING sub.new_user;