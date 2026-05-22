-- name: GlobalLeaderboardWeekly :many
-- Tổng steps 7 ngày gần nhất, không tính flagged.
SELECT u.id, u.handle, u.display_name, u.avatar_url,
       COALESCE(SUM(d.merged_steps), 0)::int AS total_steps
FROM users u
LEFT JOIN daily_steps d
       ON d.user_id = u.id
      AND d.day >= CURRENT_DATE - INTERVAL '7 days'
      AND d.flagged = false
WHERE u.status = 'active'
GROUP BY u.id
ORDER BY total_steps DESC
LIMIT $1;

-- name: ChallengeLeaderboard :many
SELECT u.id, u.handle, u.display_name, u.avatar_url, cp.total_steps, cp.days_completed
FROM challenge_participants cp
JOIN users u ON u.id = cp.user_id
WHERE cp.challenge_id = $1
ORDER BY cp.total_steps DESC
LIMIT $2;
