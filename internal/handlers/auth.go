package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/fivetentaylor/hotpog/internal/db"
	"github.com/fivetentaylor/hotpog/internal/templ/components"
	"github.com/google/uuid"
)

func (h *Handler) AuthPage(w http.ResponseWriter, r *http.Request) {
	if h.isLoggedIn(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Check if we should show login or register form
	formType := r.URL.Query().Get("form")
	if formType == "" {
		formType = "login" // Default to login form
	}

	if r.Header.Get("HX-Request") == "true" {
		// For HTMX requests, only render the form content
		component := components.AuthFormContent(formType)
		component.Render(r.Context(), w)
	} else {
		// For full page loads, render the entire page
		component := components.AuthPage(formType)
		component.Render(r.Context(), w)
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	email := r.FormValue("email")
	password := r.FormValue("password")

	_, err := h.DB.CreateUser(ctx, email, password)
	if err != nil {
		if errors.As(err, &db.ErrorEmailExists{}) {
			// Return HTMX response with error
			w.Header().Set("HX-Retarget", "#register-form")
			w.Header().Set("HX-Reswap", "innerHTML")
			http.Error(w, "Email already taken", http.StatusBadRequest)
			return
		}
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// For HTMX requests, return a redirect header
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", fmt.Sprintf("/verify?email=%s", url.QueryEscape(email)))
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/verify?email=%s", url.QueryEscape(email)), http.StatusSeeOther)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	email := r.FormValue("email")
	password := r.FormValue("password")

	sessionID, err := h.DB.Login(ctx, email, password)
	if err != nil {
		if errors.As(err, &db.ErrorUnverifiedEmail{}) {
			// For HTMX requests, return a redirect header
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", fmt.Sprintf("/verify?email=%s", url.QueryEscape(email)))
				w.WriteHeader(http.StatusOK)
				return
			}
			http.Redirect(w, r, fmt.Sprintf("/verify?email=%s", url.QueryEscape(email)), http.StatusSeeOther)
			return
		}

		// Return HTMX response with error
		w.Header().Set("HX-Retarget", "#login-form")
		w.Header().Set("HX-Reswap", "innerHTML")
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

	// For HTMX requests, return a redirect header
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
		return
	}

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

	exists, err := h.DB.Queries.UserSessionExists(r.Context(), sessionID)
	if err != nil {
		return false
	}

	return exists
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

func (h *Handler) VerifyUserEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get token from URL query params
	token := r.URL.Query().Get("token")
	if token == "" {
		component := components.VerifyEmailPage("test")
		component.Render(ctx, w)

		return
	}

	sessionID, err := h.DB.VerifyUserEmail(ctx, token)
	if err != nil {
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
