package db

import (
	"context"
	"database/sql"
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

func (db *DB) CreateUser(ctx context.Context, email, password string) (user sqlc.User, err error) {
	userExists, err := db.Queries.UserEmailExists(ctx, email)
	if err != nil {
		return user, stackerr.Wrap(err)
	}

	if userExists {
		return user, stackerr.Wrap(ErrorEmailExists{})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return user, stackerr.Wrap(err)
	}

	user, err = db.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        email,
		PasswordHash: sql.NullString{String: string(hash), Valid: true},
	})
	if err != nil {
		return user, stackerr.Wrap(err)
	}

	url, err := db.SendMagicLink(ctx, user)
	if err != nil {
		return user, stackerr.Wrap(err)
	}

	// Send the email or text message here, probably async?
	fmt.Printf("Verification URL: %s\n", url)

	return user, nil
}

type ErrorUnverifiedEmail struct{}

func (e ErrorUnverifiedEmail) Error() string {
	return "Email is unverified"
}

func (db *DB) LoginPassword(ctx context.Context, email, password string) (sessionID string, err error) {
	userExists, err := db.Queries.UserEmailExists(ctx, email)
	if err != nil {
		return "", stackerr.Wrap(err)
	}

	var user sqlc.User
	if !userExists {
		user, err = db.CreateUser(ctx, email, password)
		if err != nil {
			return "", stackerr.Wrap(err)
		}
	} else {
		user, err = db.Queries.GetUserByEmail(ctx, email)
		if err != nil {
			return "", stackerr.Wrap(err)
		}

		if !user.PasswordHash.Valid {
			return "", stackerr.Errorf("Invalid credentials")
		}

		if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(password)) != nil {
			return "", stackerr.Errorf("Invalid credentials")
		}
	}

	if !user.EmailVerifiedAt.Valid {
		url, err := db.SendMagicLink(ctx, user)
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

func (db *DB) LoginMagicLink(ctx context.Context, email string) (url string, err error) {
	userExists, err := db.Queries.UserEmailExists(ctx, email)
	if err != nil {
		return "", stackerr.Wrap(err)
	}

	var user sqlc.User
	if userExists {
		user, err = db.Queries.GetUserByEmail(ctx, email)
		if err != nil {
			return "", stackerr.Wrap(err)
		}
	} else {
		user, err = db.CreateUser(ctx, email, "")
		if err != nil {
			return "", stackerr.Wrap(err)
		}
	}

	url, err = db.SendMagicLink(ctx, user)
	if err != nil {
		return "", stackerr.Wrap(err)
	}

	return url, nil
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

	verification, err := qtx.GetAndUseMagicLink(ctx, token)
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

	return sid.String(), nil
}

func (db *DB) SendMagicLink(ctx context.Context, user sqlc.User) (string, error) {
	// Generate a magic link token
	token := uuid.New().String()
	expiresAt := time.Now().Add(15 * time.Minute) // Links expire after 15 minutes

	_, err := db.Queries.CreateMagicLink(ctx, sqlc.CreateMagicLinkParams{
		Token:     token,
		ExpiresAt: expiresAt,
		UserID:    user.ID,
	})
	if err != nil {
		return "", stackerr.Wrap(err)
	}

	// Print the magic link to console (for now)
	var magicLink string
	if user.EmailVerifiedAt.Valid {
		magicLink = fmt.Sprintf("https://%s%s/login/magic?token=%s", config.Get().Domain, config.Get().Port, token)
	} else {
		magicLink = fmt.Sprintf("https://%s%s/verify?token=%s", config.Get().Domain, config.Get().Port, token)
	}

	return magicLink, nil
}

func (db *DB) UseMagicLink(ctx context.Context, token string) (sessionID string, err error) {
	tx, err := db.Raw.BeginTx(ctx, nil)
	if err != nil {
		return "", stackerr.Wrap(err)
	}
	defer tx.Rollback()

	qtx := db.Queries.WithTx(tx)

	result, err := qtx.GetAndUseMagicLink(ctx, token)
	if err != nil {
		return "", stackerr.Wrap(err)
	}

	sid, err := qtx.CreateSession(ctx, result.UserID)
	if err != nil {
		return "", stackerr.Wrap(err)
	}

	if err = tx.Commit(); err != nil {
		return "", stackerr.Wrap(err)
	}

	return sid.String(), nil
}
