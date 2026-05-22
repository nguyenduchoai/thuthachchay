-- name: ListChallenges :many
SELECT * FROM challenges
WHERE ($1::text IS NULL OR status = $1)
ORDER BY start_date DESC
LIMIT $2 OFFSET $3;

-- name: GetChallenge :one
SELECT * FROM challenges WHERE id = $1;

-- name: CreateChallenge :one
INSERT INTO challenges
    (host_id, visibility, name, description, cover_url, daily_steps_target,
     duration_days, entry_points, prize_pool, max_participants, start_date, end_date, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 'open')
RETURNING *;

-- name: JoinChallenge :one
INSERT INTO challenge_participants (challenge_id, user_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING
RETURNING *;

-- name: ListMyChallenges :many
SELECT c.*, cp.total_steps, cp.days_completed, cp.won, cp.payout_points
FROM challenge_participants cp
JOIN challenges c ON c.id = cp.challenge_id
WHERE cp.user_id = $1
ORDER BY c.start_date DESC;

-- name: ChallengesToSettle :many
SELECT * FROM challenges
WHERE status = 'live' AND end_date < CURRENT_DATE
ORDER BY end_date ASC
LIMIT $1;
