package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/fivetentaylor/hotpog/internal/db"
	"github.com/fivetentaylor/hotpog/internal/templ/components"
	"github.com/google/uuid"
)

func (h *Handler) PasswordLoginPage(w http.ResponseWriter, r *http.Request) {
	if h.IsLoggedIn(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	component := components.PasswordLoginPage()
	component.Render(r.Context(), w)
}

func (h *Handler) MagicLoginPage(w http.ResponseWriter, r *http.Request) {
	if h.IsLoggedIn(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	component := components.MagicLoginPage()
	component.Render(r.Context(), w)
}

func (h *Handler) LoginPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	email := r.FormValue("email")
	password := r.FormValue("password")

	sessionID, err := h.DB.LoginPassword(ctx, email, password)
	if err != nil {
		if errors.As(err, &db.ErrorUnverifiedEmail{}) {
			http.Redirect(w, r, fmt.Sprintf("/verify?email=%s", url.QueryEscape(email)), http.StatusSeeOther)
			return
		}

		slog.Error("Error logging in", "error", err)
		http.Error(w, "Invalid credentials", http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   60 * 60 * 24 * 30, // 30 days
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) IsLoggedIn(r *http.Request) bool {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		slog.Error("Error getting session cookie", "error", err)
		return false
	}

	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		slog.Error("Error parsing session ID", "error", err)
		return false
	}

	_, err = h.DB.Queries.GetValidSession(r.Context(), sessionID)

	return err == nil
}

func (h *Handler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			slog.Error("Error getting session cookie", "error", err)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		sessionID, err := uuid.Parse(cookie.Value)
		if err != nil {
			slog.Error("Error parsing session ID", "error", err)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		session, err := h.DB.Queries.GetValidSession(r.Context(), sessionID)
		if err != nil {
			slog.Error("Error getting valid session", "error", err)
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
		slog.Error("Error getting session cookie", "error", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		slog.Error("Error parsing session ID", "error", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err = h.DB.Queries.DeleteSession(r.Context(), sessionID)
	if err != nil {
		slog.Error("Error deleting session", "error", err)
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

func (h *Handler) SendMagicLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	email := r.FormValue("email")

	url, err := h.DB.LoginMagicLink(ctx, email)
	if err != nil {
		slog.Error("Error sending magic link", "error", err)
		http.Error(w, "Invalid email", http.StatusBadRequest)
		return
	}

	fmt.Println("Magic link URL:", url)

	// Show success message to user
	component := components.VerifyEmailPage(email)
	component.Render(ctx, w)
}

func (h *Handler) UseMagicLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get token from URL query params
	token := r.URL.Query().Get("token")
	if token == "" {
		email := r.URL.Query().Get("email")
		component := components.VerifyEmailPage(email)
		component.Render(ctx, w)

		return
	}

	sessionID, err := h.DB.UseMagicLink(ctx, token)
	if err != nil {
		slog.Error("Error using magic link", "error", err)
		http.Error(w, "Invalid or expired magic link", http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   60 * 60 * 24 * 30, // 30 days
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) VerifyUserEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get token from URL query params
	token := r.URL.Query().Get("token")
	if token == "" {
		email := r.URL.Query().Get("email")
		component := components.VerifyEmailPage(email)
		component.Render(ctx, w)

		return
	}

	sessionID, err := h.DB.VerifyUserEmail(ctx, token)
	if err != nil {
		slog.Error("Error verifying email", "error", err)
		http.Error(w, "Invalid or expired verification token", http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   60 * 60 * 24 * 30, // 30 days
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
