// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type MagicLink struct {
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
	UserID    uuid.UUID
	UsedAt    sql.NullTime
}

type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type User struct {
	ID                  uuid.UUID
	Email               string
	PasswordHash        sql.NullString
	EmailVerifiedAt     sql.NullTime
	FailedLoginAttempts sql.NullInt32
	LastLoginAt         sql.NullTime
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Phone               sql.NullString
	PhoneVerifiedAt     sql.NullTime
}

type Verification struct {
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
	UserID    uuid.UUID
	UsedAt    sql.NullTime
}
