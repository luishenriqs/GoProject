// infra/database/user_db_test.go
package database

import (
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	appentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // silencia logs
	})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}

	if err := db.AutoMigrate(&appentity.User{}); err != nil {
		t.Fatalf("failed to automigrate: %v", err)
	}
	return db
}

func TestUserRepository_Create_And_FindByEmail_Success(t *testing.T) {
	db := newTestDB(t)
	repo := NewUser(db)

	u, err := appentity.NewUser("Luís", "  TEST.User+tag@Example.COM ", "s3cr3t-Strong!")
	if err != nil {
		t.Fatalf("NewUser failed: %v", err)
	}

	if err := repo.Create(u); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByEmail("test.user+tag@example.com")
	if err != nil {
		t.Fatalf("FindByEmail failed: %v", err)
	}
	if found == nil {
		t.Fatalf("expected non-nil user")
	}
	if found.ID != u.ID || found.Name != u.Name || found.Email != u.Email {
		t.Fatalf("mismatch found vs created")
	}
	if found.Password == "" || !found.CheckPassword("s3cr3t-Strong!") || found.CheckPassword("wrong") {
		t.Fatalf("password/hash checks failed")
	}
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewUser(db)

	_, err := repo.FindByEmail("nobody@example.com")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound, got %v", err)
	}
}
