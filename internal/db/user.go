package db

import (
	"context"
	"database/sql"

	"golang.org/x/crypto/bcrypt"

	sqlc "github.com/fivetentaylor/hotpog/internal/db/generated"
	"github.com/fivetentaylor/hotpog/internal/stackerr"
	"github.com/google/uuid"
)

func (db *DB) CreateUser(ctx context.Context, email, password string) (userID string, err error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	uid, err := db.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        email,
		PasswordHash: sql.NullString{String: string(hash), Valid: true},
	})
	if err != nil {
		return "", err
	}

	return uid.String(), nil
}

func (db *DB) Login(ctx context.Context, email, password string) (sessionID string, err error) {
	user, err := db.Queries.GetUserByEmail(ctx, email)
	if user.PasswordHash.Valid == false {
		return "", stackerr.Errorf("Invalid credentials")
	}

	if err != nil {
		return "", stackerr.Wrap(err)
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(password)) != nil {
		return "", stackerr.Errorf("Invalid credentials")
	}

	sid, err := db.Queries.CreateSession(ctx, user.ID)
	if err != nil {
		stackerr.Wrap(err)
	}

	sessionID = sid.String()
	return sessionID, nil
}

func (db *DB) DeleteSession(ctx context.Context, sessionID string) error {
	sid, err := uuid.Parse(sessionID)
	if err != nil {
		return stackerr.Wrap(err)
	}

	err = db.Queries.DeleteSession(ctx, sid)
	if err != nil {
		return stackerr.Wrap(err)
	}

	return nil
}

func (db *DB) VerifyUserEmail(ctx context.Context, userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return stackerr.Wrap(err)
	}

	err = db.Queries.VerifyUserEmail(ctx, uid)
	if err != nil {
		return stackerr.Wrap(err)
	}

	return nil
}
