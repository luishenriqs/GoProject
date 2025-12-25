package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/jwtauth/v5"
	pkgentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/pkg/entity"
)

func TestAuthMiddleware_NoHeader_Returns401(t *testing.T) {
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := AuthMiddleware(tokenAuth)(next)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	if called {
		t.Fatalf("expected next handler not to be called")
	}
}

func TestAuthMiddleware_InvalidToken_Returns401(t *testing.T) {
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := AuthMiddleware(tokenAuth)(next)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.value")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	if called {
		t.Fatalf("expected next handler not to be called")
	}
}

func TestAuthMiddleware_ValidToken_CallsNextAndInjectsUserID(t *testing.T) {
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	expectedID := pkgentity.NewId()
	_, tokenString, err := tokenAuth.Encode(map[string]interface{}{
		"sub": expectedID.String(),
	})
	if err != nil {
		t.Fatalf("expected no error encoding token, got: %v", err)
	}

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true

		gotID, ok := UserIDFromContext(r.Context())
		if !ok {
			t.Fatalf("expected userID in context")
		}

		if gotID != expectedID {
			t.Fatalf("expected userID %v, got %v", expectedID, gotID)
		}

		w.WriteHeader(http.StatusOK)
	})

	handler := AuthMiddleware(tokenAuth)(next)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if !called {
		t.Fatalf("expected next handler to be called")
	}
}
