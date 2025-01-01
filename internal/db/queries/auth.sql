-- internal/db/queries/auth.sql
-- name: GetValidSession :one
SELECT s.id, s.user_id, u.email 
FROM sessions s
JOIN users u ON s.user_id = u.id
WHERE s.id = $1 
 AND s.expires_at > NOW()
 AND u.email_verified_at IS NOT NULL;
