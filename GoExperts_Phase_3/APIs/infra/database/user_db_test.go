package database

import (
	"errors"
	"testing"

	appentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/entity"
	pkgentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/pkg/entity"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupUserTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite db: %v", err)
	}

	if err := db.AutoMigrate(&appentity.User{}); err != nil {
		t.Fatalf("failed to migrate user schema: %v", err)
	}

	return db
}

func TestUserRepository_Create_And_FindByEmail_Success(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserDb(db)

	u, err := appentity.NewUser("John Doe", "john_doe@email.com", "123")
	if err != nil {
		t.Fatalf("expected no error creating user entity, got: %v", err)
	}

	if err := repo.Create(u); err != nil {
		t.Fatalf("expected no error creating user, got: %v", err)
	}

	found, err := repo.FindByEmail("john_doe@email.com")
	if err != nil {
		t.Fatalf("expected no error finding user by email, got: %v", err)
	}

	if found.Email != "john_doe@email.com" {
		t.Fatalf("expected email %q, got %q", "john_doe@email.com", found.Email)
	}

	if found.Name != "John Doe" {
		t.Fatalf("expected name %q, got %q", "John Doe", found.Name)
	}

	if !found.CheckPassword("123") {
		t.Fatalf("expected password to match")
	}
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserDb(db)

	_, err := repo.FindByEmail("missing@email.com")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
	}
}

func TestUserRepository_FindByID_Success(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserDb(db)

	u, err := appentity.NewUser("John Doe", "john_doe@email.com", "123")
	if err != nil {
		t.Fatalf("expected no error creating user entity, got: %v", err)
	}

	if err := repo.Create(u); err != nil {
		t.Fatalf("expected no error creating user, got: %v", err)
	}

	found, err := repo.FindByID(u.ID)
	if err != nil {
		t.Fatalf("expected no error finding user by id, got: %v", err)
	}

	if found.ID != u.ID {
		t.Fatalf("expected id %v, got %v", u.ID, found.ID)
	}
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserDb(db)

	_, err := repo.FindByID(pkgentity.NewId())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
	}
}

func TestUserRepository_Update_Success(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserDb(db)

	u, err := appentity.NewUser("John Doe", "john_doe@email.com", "123")
	if err != nil {
		t.Fatalf("expected no error creating user entity, got: %v", err)
	}

	if err := repo.Create(u); err != nil {
		t.Fatalf("expected no error creating user, got: %v", err)
	}

	u.Name = "John Updated"
	if err := u.SetPassword("456"); err != nil {
		t.Fatalf("expected no error setting password, got: %v", err)
	}

	if err := repo.Update(u); err != nil {
		t.Fatalf("expected no error updating user, got: %v", err)
	}

	found, err := repo.FindByID(u.ID)
	if err != nil {
		t.Fatalf("expected no error finding updated user, got: %v", err)
	}

	if found.Name != "John Updated" {
		t.Fatalf("expected name %q, got %q", "John Updated", found.Name)
	}

	if !found.CheckPassword("456") {
		t.Fatalf("expected updated password to match")
	}
}

func TestUserRepository_Update_NotFound(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserDb(db)

	u, err := appentity.NewUser("John Doe", "john_doe@email.com", "123")
	if err != nil {
		t.Fatalf("expected no error creating user entity, got: %v", err)
	}

	err = repo.Update(u)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
	}
}

func TestUserRepository_Delete_Success(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserDb(db)

	u, err := appentity.NewUser("John Doe", "john_doe@email.com", "123")
	if err != nil {
		t.Fatalf("expected no error creating user entity, got: %v", err)
	}

	if err := repo.Create(u); err != nil {
		t.Fatalf("expected no error creating user, got: %v", err)
	}

	if err := repo.Delete(u.ID); err != nil {
		t.Fatalf("expected no error deleting user, got: %v", err)
	}

	_, err = repo.FindByID(u.ID)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
	}
}

func TestUserRepository_Delete_NotFound(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserDb(db)

	err := repo.Delete(pkgentity.NewId())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
	}
}
