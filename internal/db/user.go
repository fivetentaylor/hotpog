package db

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/fivetentaylor/hotpog/internal/config"
	sqlc "github.com/fivetentaylor/hotpog/internal/db/generated"
	"github.com/fivetentaylor/hotpog/internal/stackerr"
	"github.com/google/uuid"
)

type ErrorEmailExists struct{}

func (e ErrorEmailExists) Error() string {
	return "User email already exists"
}

func (db *DB) CreateUser(ctx context.Context, email, password string) (userID string, err error) {
	userExists, err := db.Queries.UserEmailExists(ctx, email)
	if err != nil {
		return "", stackerr.Wrap(err)
	}

	if userExists {
		return "", stackerr.Wrap(ErrorEmailExists{})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", stackerr.Wrap(err)
	}

	uid, err := db.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        email,
		PasswordHash: sql.NullString{String: string(hash), Valid: true},
	})
	if err != nil {
		return "", stackerr.Wrap(err)
	}

	url, err := db._createVerificationURL(ctx, uid)
	if err != nil {
		return "", stackerr.Wrap(err)
	}

	// Send the email or text message here, probably async?
	fmt.Printf("Verification URL: %s\n", url)

	return uid.String(), nil
}

type ErrorUnverifiedEmail struct{}

func (e ErrorUnverifiedEmail) Error() string {
	return "Email is unverified"
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

	if user.EmailVerifiedAt.Valid == false {
		url, err := db._createVerificationURL(ctx, user.ID)
		if err != nil {
			return "", stackerr.Wrap(err)
		}

		// Send the email or text message here, probably async?
		fmt.Printf("Verification URL: %s\n", url)

		return "", ErrorUnverifiedEmail{}
	}

	sid, err := db.Queries.CreateSession(ctx, user.ID)
	if err != nil {
		stackerr.Wrap(err)
	}

	sessionID = sid.String()
	return sessionID, nil
}

func (db *DB) Register(ctx context.Context, email, password string) (userID string, err error) {
	userID, err = db.CreateUser(ctx, email, password)
	if err != nil {
		return "", stackerr.Wrap(err)
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return "", stackerr.Wrap(err)
	}

	url, err := db._createVerificationURL(ctx, uid)
	if err != nil {
		return "", stackerr.Wrap(err)
	}

	// Send the email or text message here, probably async?
	fmt.Printf("Verification URL: %s\n", url)

	return userID, nil
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

func (db *DB) VerifyUserEmail(ctx context.Context, token string) (sessionID string, err error) {
	tx, err := db.Raw.BeginTx(ctx, nil)
	if err != nil {
		return "", stackerr.Wrap(err)
	}
	defer tx.Rollback()

	qtx := db.Queries.WithTx(tx)

	verification, err := qtx.GetAndUseVerification(ctx, token)
	if err != nil {
		return "", stackerr.Wrap(err)
	}

	err = qtx.VerifyUserEmail(ctx, verification.UserID)
	if err != nil {
		return "", stackerr.Wrap(err)
	}

	sid, err := qtx.CreateSession(ctx, verification.UserID)
	if err != nil {
		return "", stackerr.Wrap(err)
	}

	if err = tx.Commit(); err != nil {
		return "", stackerr.Wrap(err)
	}

	sessionID = sid.String()

	return sessionID, nil
}

func (db *DB) _createVerificationURL(ctx context.Context, userID uuid.UUID) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	token := hex.EncodeToString(bytes)

	_, err := db.Queries.CreateVerification(ctx, sqlc.CreateVerificationParams{
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UserID:    userID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create verification record: %w", err)
	}

	return fmt.Sprintf("https://%s%s/verify?token=%s", config.Get().Domain, config.Get().Port, token), nil
}
