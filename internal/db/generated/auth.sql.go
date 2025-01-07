// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: auth.sql

package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const createMagicLink = `-- name: CreateMagicLink :one
INSERT INTO magic_links (
    token,
    expires_at,
    user_id
) VALUES (
    $1,
    $2,
    $3
) RETURNING token, created_at, expires_at, user_id, used_at
`

type CreateMagicLinkParams struct {
	Token     string
	ExpiresAt time.Time
	UserID    uuid.UUID
}

func (q *Queries) CreateMagicLink(ctx context.Context, arg CreateMagicLinkParams) (MagicLink, error) {
	row := q.db.QueryRowContext(ctx, createMagicLink, arg.Token, arg.ExpiresAt, arg.UserID)
	var i MagicLink
	err := row.Scan(
		&i.Token,
		&i.CreatedAt,
		&i.ExpiresAt,
		&i.UserID,
		&i.UsedAt,
	)
	return i, err
}

const createSession = `-- name: CreateSession :one
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
LIMIT 1
`

func (q *Queries) CreateSession(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	row := q.db.QueryRowContext(ctx, createSession, userID)
	var id uuid.UUID
	err := row.Scan(&id)
	return id, err
}

const createUser = `-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING id
`

type CreateUserParams struct {
	Email        string
	PasswordHash sql.NullString
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (uuid.UUID, error) {
	row := q.db.QueryRowContext(ctx, createUser, arg.Email, arg.PasswordHash)
	var id uuid.UUID
	err := row.Scan(&id)
	return id, err
}

const createVerification = `-- name: CreateVerification :one
INSERT INTO verifications (
    token,
    expires_at,
    user_id
) VALUES (
    $1,
    $2,
    $3
) RETURNING token, created_at, expires_at, user_id, used_at
`

type CreateVerificationParams struct {
	Token     string
	ExpiresAt time.Time
	UserID    uuid.UUID
}

func (q *Queries) CreateVerification(ctx context.Context, arg CreateVerificationParams) (Verification, error) {
	row := q.db.QueryRowContext(ctx, createVerification, arg.Token, arg.ExpiresAt, arg.UserID)
	var i Verification
	err := row.Scan(
		&i.Token,
		&i.CreatedAt,
		&i.ExpiresAt,
		&i.UserID,
		&i.UsedAt,
	)
	return i, err
}

const deleteSession = `-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1
`

func (q *Queries) DeleteSession(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, deleteSession, id)
	return err
}

const getAndUseMagicLink = `-- name: GetAndUseMagicLink :one
UPDATE magic_links 
SET used_at = CURRENT_TIMESTAMP
WHERE token = $1 
    AND used_at IS NULL 
    AND expires_at > CURRENT_TIMESTAMP
RETURNING token, created_at, expires_at, user_id, used_at
`

func (q *Queries) GetAndUseMagicLink(ctx context.Context, token string) (MagicLink, error) {
	row := q.db.QueryRowContext(ctx, getAndUseMagicLink, token)
	var i MagicLink
	err := row.Scan(
		&i.Token,
		&i.CreatedAt,
		&i.ExpiresAt,
		&i.UserID,
		&i.UsedAt,
	)
	return i, err
}

const getAndUseVerification = `-- name: GetAndUseVerification :one
UPDATE verifications 
SET used_at = CURRENT_TIMESTAMP
WHERE token = $1 
    AND used_at IS NULL 
    AND expires_at > CURRENT_TIMESTAMP
RETURNING token, created_at, expires_at, user_id, used_at
`

func (q *Queries) GetAndUseVerification(ctx context.Context, token string) (Verification, error) {
	row := q.db.QueryRowContext(ctx, getAndUseVerification, token)
	var i Verification
	err := row.Scan(
		&i.Token,
		&i.CreatedAt,
		&i.ExpiresAt,
		&i.UserID,
		&i.UsedAt,
	)
	return i, err
}

const getUser = `-- name: GetUser :one
SELECT id, password_hash, email_verified_at
  FROM users
WHERE id = $1
`

type GetUserRow struct {
	ID              uuid.UUID
	PasswordHash    sql.NullString
	EmailVerifiedAt sql.NullTime
}

func (q *Queries) GetUser(ctx context.Context, id uuid.UUID) (GetUserRow, error) {
	row := q.db.QueryRowContext(ctx, getUser, id)
	var i GetUserRow
	err := row.Scan(&i.ID, &i.PasswordHash, &i.EmailVerifiedAt)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, password_hash, email_verified_at
FROM users
WHERE email = $1
`

type GetUserByEmailRow struct {
	ID              uuid.UUID
	PasswordHash    sql.NullString
	EmailVerifiedAt sql.NullTime
}

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (GetUserByEmailRow, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmail, email)
	var i GetUserByEmailRow
	err := row.Scan(&i.ID, &i.PasswordHash, &i.EmailVerifiedAt)
	return i, err
}

const getValidSession = `-- name: GetValidSession :one

SELECT s.id, s.user_id, u.email 
FROM sessions s
JOIN users u ON s.user_id = u.id
WHERE s.id = $1 
 AND s.expires_at > NOW()
 AND u.email_verified_at IS NOT NULL
`

type GetValidSessionRow struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Email  string
}

// internal/db/queries/auth.sql
func (q *Queries) GetValidSession(ctx context.Context, id uuid.UUID) (GetValidSessionRow, error) {
	row := q.db.QueryRowContext(ctx, getValidSession, id)
	var i GetValidSessionRow
	err := row.Scan(&i.ID, &i.UserID, &i.Email)
	return i, err
}

const userEmailExists = `-- name: UserEmailExists :one
SELECT EXISTS (
    SELECT 1 FROM users
    WHERE email = $1
)
`

func (q *Queries) UserEmailExists(ctx context.Context, email string) (bool, error) {
	row := q.db.QueryRowContext(ctx, userEmailExists, email)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const userSessionExists = `-- name: UserSessionExists :one
SELECT EXISTS (
    SELECT 1 FROM sessions
    WHERE id = $1
)
`

func (q *Queries) UserSessionExists(ctx context.Context, id uuid.UUID) (bool, error) {
	row := q.db.QueryRowContext(ctx, userSessionExists, id)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const verifyUserEmail = `-- name: VerifyUserEmail :exec
UPDATE users
SET email_verified_at = NOW()
WHERE id = $1
`

func (q *Queries) VerifyUserEmail(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, verifyUserEmail, id)
	return err
}
