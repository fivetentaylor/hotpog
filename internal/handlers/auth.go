package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	sqlc "github.com/fivetentaylor/hotpog/internal/db/generated"
	"github.com/fivetentaylor/hotpog/internal/templ/components"
	"github.com/google/uuid"
)

func (h *Handler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	if h.isLoggedIn(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	component := components.RegisterPage()
	component.Render(r.Context(), w)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	email := r.FormValue("email")
	password := r.FormValue("password")

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Server error", 500)
		return
	}

	userID, err := h.DB.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        email,
		PasswordHash: sql.NullString{String: string(hash), Valid: true},
	})
	if err != nil {
		http.Error(w, "Email already taken", 400)
		return
	}
	fmt.Printf("userID: %s\n", userID)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if h.isLoggedIn(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	component := components.LoginPage()
	component.Render(r.Context(), w)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.DB.Queries.GetUserByEmail(ctx, email)
	if user.PasswordHash.Valid == false {
		http.Error(w, "Invalid credentials", 401)
		return
	}

	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(password)) != nil {
		http.Error(w, "Invalid credentials", 401)
		return
	}

	sessionID, err := h.DB.Queries.CreateSession(ctx, user.ID)
	if err != nil {
		http.Error(w, "Server error", 500)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID.String(),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   60 * 60 * 24 * 30, // 30 days
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) isLoggedIn(r *http.Request) bool {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return false
	}

	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		return false
	}

	_, err = h.DB.Queries.GetValidSession(r.Context(), sessionID)
	if err != nil {
		return false
	}

	return true
}

func (h *Handler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		sessionID, err := uuid.Parse(cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		session, err := h.DB.Queries.GetValidSession(r.Context(), sessionID)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), "user_id", session.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err = h.DB.Queries.DeleteSession(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "Server error", 500)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
