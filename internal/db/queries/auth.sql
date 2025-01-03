-- internal/db/queries/auth.sql

-- name: GetValidSession :one
SELECT s.id, s.user_id, u.email 
FROM sessions s
JOIN users u ON s.user_id = u.id
WHERE s.id = $1 
 AND s.expires_at > NOW()
 AND u.email_verified_at IS NOT NULL;

-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING id;

-- name: GetUserByEmail :one
SELECT id, password_hash, email_verified_at
FROM users
WHERE email = $1;

-- name: CreateSession :one
WITH new_session AS (
  INSERT INTO sessions (user_id, expires_at)
  SELECT $1, NOW() + interval '30 days'
  WHERE NOT EXISTS (
    SELECT 1 FROM sessions 
    WHERE user_id = $1 
    AND expires_at > NOW()
  )
  RETURNING id
)
SELECT id FROM new_session
UNION ALL
SELECT id FROM sessions 
WHERE user_id = $1 
AND expires_at > NOW()
LIMIT 1;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1;
