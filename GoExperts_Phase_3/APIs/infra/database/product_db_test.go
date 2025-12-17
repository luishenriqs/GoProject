// infra/database/product_db_test.go
package database

import (
	"errors"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	appentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

/*
- abre um SQLite :memory: com GORM (ou seja, banco em RAM, isolado por teste)
- configura o logger como Silent para não poluir o output
- executa AutoMigrate(&appentity.Product{}) para criar a tabela/estrutura necessária antes de rodar os testes
- retorna o *gorm.DB pronto pra uso
*/
func newProductTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // silencia logs
	})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}
	if err := db.AutoMigrate(&appentity.Product{}); err != nil {
		t.Fatalf("failed to automigrate: %v", err)
	}
	return db
}

func TestProductDB_Create_And_FindByID(t *testing.T) {
	db := newProductTestDB(t) // → cria o banco
	repo := NewProductDB(db)  // → cria o repositório apontando para esse banco (injeção de dependência)

	// Cria um objeto do novo produto
	p, err := appentity.NewProduct("Notebook", 4999.90)
	if err != nil {
		t.Fatalf("NewProduct failed: %v", err)
	}

	// Persiste o novo produto no banco
	if err := repo.Create(p); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := repo.FindByID(p.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if got.ID != p.ID || got.Name != "Notebook" || got.Price != 4999.90 || got.CreatedAt.IsZero() {
		t.Fatalf("mismatch fields after create/find")
	}
}

func TestProductDB_FindAll_Pagination(t *testing.T) {
	db := newProductTestDB(t)
	repo := NewProductDB(db)

	names := []string{"A", "B", "C", "D", "E"}
	for i, n := range names {
		p, err := appentity.NewProduct(n, float64(10*(i+1)))
		if err != nil {
			t.Fatalf("NewProduct(%s) failed: %v", n, err)
		}
		if err := repo.Create(p); err != nil {
			t.Fatalf("Create(%s) failed: %v", n, err)
		}
		time.Sleep(2 * time.Millisecond)
	}

	items, total, err := repo.FindAll(2, 2, "asc")
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	if total != 5 || len(items) != 2 || items[0].Name != "C" || items[1].Name != "D" {
		t.Fatalf("unexpected pagination result: total=%d len=%d first=%s second=%s",
			total, len(items), items[0].Name, items[1].Name)
	}

	items, total, err = repo.FindAll(3, 2, "asc")
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	if total != 5 || len(items) != 1 || items[0].Name != "E" {
		t.Fatalf("unexpected pagination page3")
	}
}

func TestProductDB_Update(t *testing.T) {
	db := newProductTestDB(t)
	repo := NewProductDB(db)

	p, err := appentity.NewProduct("Mouse", 99.9)
	if err != nil {
		t.Fatalf("NewProduct failed: %v", err)
	}
	if err := repo.Create(p); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	p.Name = "Mouse Pro"
	p.Price = 149.9
	if err := repo.Update(p); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, err := repo.FindByID(p.ID)
	if err != nil {
		t.Fatalf("FindByID after update failed: %v", err)
	}
	if got.Name != "Mouse Pro" || got.Price != 149.9 || got.CreatedAt.IsZero() {
		t.Fatalf("update mismatch")
	}
}

func TestProductDB_Update_NotFound(t *testing.T) {
	db := newProductTestDB(t)
	repo := NewProductDB(db)

	p, err := appentity.NewProduct("Ghost", 1.0)
	if err != nil {
		t.Fatalf("NewProduct failed: %v", err)
	}
	// NÃO persiste p — Update deve falhar com ErrRecordNotFound
	if err := repo.Update(p); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestProductDB_Delete(t *testing.T) {
	db := newProductTestDB(t)
	repo := NewProductDB(db)

	p, err := appentity.NewProduct("Teclado", 199.0)
	if err != nil {
		t.Fatalf("NewProduct failed: %v", err)
	}
	if err := repo.Create(p); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := repo.Delete(p.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.FindByID(p.ID)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound after delete, got %v", err)
	}
}
