// internal/entity/user_test.go
package entity

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewUser_Success(t *testing.T) {
	name := "Luís"
	rawEmail := "  TEST.User+tag@Example.COM  "
	plain := "s3cr3t-Strong!"

	u, err := NewUser(name, rawEmail, plain)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if u == nil {
		t.Fatalf("expected non-nil user")
	}

	// ID deve ser gerado (uuid != uuid.Nil)
	if u.ID == uuid.Nil {
		t.Fatalf("expected non-nil uuid, got Nil")
	}

	// Email deve ser normalizado (trim + lower-case)
	wantEmail := "test.user+tag@example.com"
	if u.Email != wantEmail {
		t.Fatalf("expected email %q, got %q", wantEmail, u.Email)
	}

	// Password deve armazenar o hash (não pode ser igual ao plain) e validar com CheckPassword
	if u.Password == "" {
		t.Fatalf("expected hashed password, got empty string")
	}
	if u.Password == plain {
		t.Fatalf("expected hashed password different from plain text")
	}
	if ok := u.CheckPassword(plain); !ok {
		t.Fatalf("expected CheckPassword to succeed for the original plain password")
	}
	if ok := u.CheckPassword("wrong-password"); ok {
		t.Fatalf("expected CheckPassword to fail for a wrong password")
	}
}

func TestNewUser_InvalidEmail(t *testing.T) {
	_, err := NewUser("Any Name", "not-an-email", "secret")
	if err == nil {
		t.Fatalf("expected error for invalid email, got nil")
	}
	if err != ErrInvalidEmail {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestNewUser_WeakPassword_Empty(t *testing.T) {
	_, err := NewUser("Any Name", "user@example.com", "")
	if err == nil {
		t.Fatalf("expected error for empty password, got nil")
	}
	if err != ErrWeakPassword {
		t.Fatalf("expected ErrWeakPassword, got %v", err)
	}
}

func TestNewUser_WeakPassword_Spaces(t *testing.T) {
	_, err := NewUser("Any Name", "user@example.com", "   ")
	if err == nil {
		t.Fatalf("expected error for blank/space-only password, got nil")
	}
	if err != ErrWeakPassword {
		t.Fatalf("expected ErrWeakPassword, got %v", err)
	}
}
