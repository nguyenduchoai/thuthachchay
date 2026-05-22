-- name: InsertStepIngestEvent :one
INSERT INTO step_ingest_events
    (user_id, day, source, steps, client_nonce, sensor_hash, raw)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (user_id, client_nonce) DO NOTHING
RETURNING *;

-- name: UpsertDailySteps :one
INSERT INTO daily_steps (user_id, day, zmp_steps, strava_steps, merged_steps, cadence_variance, flagged, flag_reasons)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (user_id, day) DO UPDATE
SET zmp_steps       = EXCLUDED.zmp_steps,
    strava_steps    = EXCLUDED.strava_steps,
    merged_steps    = EXCLUDED.merged_steps,
    cadence_variance= EXCLUDED.cadence_variance,
    flagged         = daily_steps.flagged OR EXCLUDED.flagged,
    flag_reasons    = (SELECT ARRAY(SELECT DISTINCT unnest(daily_steps.flag_reasons || EXCLUDED.flag_reasons)))
RETURNING *;

-- name: GetDailyStepsRange :many
SELECT * FROM daily_steps
WHERE user_id = $1 AND day BETWEEN $2 AND $3
ORDER BY day DESC;

-- name: GetTodaySteps :one
SELECT * FROM daily_steps
WHERE user_id = $1 AND day = CURRENT_DATE;
