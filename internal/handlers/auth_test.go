package handlers_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/fivetentaylor/hotpog/internal/db"
	"github.com/fivetentaylor/hotpog/internal/handlers"
	"github.com/fivetentaylor/hotpog/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestLogin(t *testing.T) {
	testutils.RunWithTestDB(t, func(db *db.DB) {
		h := handlers.NewHandler(db)

		_, err := db.CreateUser(context.Background(), "test@example.com", "test")
		require.NoError(t, err)

		// Create form data
		formData := url.Values{}
		formData.Set("email", "test@example.com")
		formData.Set("password", "testpassword123")

		// Create a new HTTP POST request with form data
		req := httptest.NewRequest(
			http.MethodPost,
			"/login",
			strings.NewReader(formData.Encode()),
		)

		// Set the content type header for form data
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// Create a response recorder to record the response
		w := httptest.NewRecorder()

		// Call the handler
		h.Login(w, req)

		// Get the result
		res := w.Result()
		defer res.Body.Close()

		// Assert the status code
		if res.StatusCode != http.StatusOK {
			t.Errorf("expected status OK; got %v", res.StatusCode)
		}

		// Optionally read and check the response body
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("could not read response: %v", err)
		}

		fmt.Printf("Response body: %s\n", string(body))

		// Add your assertions here based on the expected response
		// Example:
		// if !strings.Contains(string(body), "Login successful") {
		//     t.Errorf("unexpected response body: %s", string(body))
		// }
	})
}
