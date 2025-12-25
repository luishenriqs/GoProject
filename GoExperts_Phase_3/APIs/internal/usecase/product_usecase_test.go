package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/dto"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/entity"
	pkgentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/pkg/entity"
	"gorm.io/gorm"
)

// fakeProductRepo é um repositório em memória para testar o usecase sem GORM/DB.
type fakeProductRepo struct {
	products map[pkgentity.ID]*entity.Product

	createErr error
	findErr   error
	updateErr error
	deleteErr error

	findAllItems []entity.Product
	findAllTotal int64
	findAllErr   error
}

func newFakeProductRepo() *fakeProductRepo {
	return &fakeProductRepo{
		products: make(map[pkgentity.ID]*entity.Product),
	}
}

func (r *fakeProductRepo) Create(product *entity.Product) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.products[product.ID] = product
	return nil
}

func (r *fakeProductRepo) FindByID(id pkgentity.ID) (*entity.Product, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	p, ok := r.products[id]
	if !ok || p == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return p, nil
}

func (r *fakeProductRepo) FindAll(page, limit int, sort string) ([]entity.Product, int64, error) {
	if r.findAllErr != nil {
		return nil, 0, r.findAllErr
	}
	return r.findAllItems, r.findAllTotal, nil
}

func (r *fakeProductRepo) Update(product *entity.Product) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	_, ok := r.products[product.ID]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	r.products[product.ID] = product
	return nil
}

func (r *fakeProductRepo) Delete(id pkgentity.ID) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	_, ok := r.products[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	delete(r.products, id)
	return nil
}

func TestProductUseCase_CreateProduct_Success(t *testing.T) {
	repo := newFakeProductRepo()
	uc := NewProductUseCase(repo)

	p, err := uc.CreateProduct(dto.CreateProductInput{
		Name:  "Product 1",
		Price: 10,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if p.ID == (pkgentity.ID{}) {
		t.Fatalf("expected non-empty id")
	}

	if p.CreatedAt.IsZero() {
		t.Fatalf("expected CreatedAt to be set")
	}

	if p.CreatedAt.Location() != time.UTC {
		t.Fatalf("expected CreatedAt in UTC")
	}
}

func TestProductUseCase_CreateProduct_InvalidName(t *testing.T) {
	repo := newFakeProductRepo()
	uc := NewProductUseCase(repo)

	_, err := uc.CreateProduct(dto.CreateProductInput{
		Name:  "   ",
		Price: 10,
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, entity.ErrRequired) {
		t.Fatalf("expected ErrRequired, got: %v", err)
	}
}

func TestProductUseCase_CreateProduct_InvalidPrice(t *testing.T) {
	repo := newFakeProductRepo()
	uc := NewProductUseCase(repo)

	_, err := uc.CreateProduct(dto.CreateProductInput{
		Name:  "Product",
		Price: -1,
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, entity.ErrInvalidPrice) {
		t.Fatalf("expected ErrInvalidPrice, got: %v", err)
	}
}

func TestProductUseCase_GetProduct_NotFound(t *testing.T) {
	repo := newFakeProductRepo()
	uc := NewProductUseCase(repo)

	_, err := uc.GetProduct(pkgentity.NewId())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
	}
}

func TestProductUseCase_ListProducts_Success(t *testing.T) {
	repo := newFakeProductRepo()
	uc := NewProductUseCase(repo)

	repo.findAllItems = []entity.Product{
		{ID: pkgentity.NewId(), Name: "A", Price: 10},
		{ID: pkgentity.NewId(), Name: "B", Price: 20},
	}
	repo.findAllTotal = 2

	items, total, err := uc.ListProducts(1, 10, "asc")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestProductUseCase_UpdateProduct_Success(t *testing.T) {
	repo := newFakeProductRepo()
	uc := NewProductUseCase(repo)

	p, err := entity.NewProduct("Product", 10)
	if err != nil {
		t.Fatalf("expected no error creating entity, got: %v", err)
	}
	_ = repo.Create(p)

	newName := "Product Updated"
	newPrice := 25.0

	updated, err := uc.UpdateProduct(p.ID, dto.UpdateProductInput{
		Name:  &newName,
		Price: &newPrice,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if updated.Name != "Product Updated" {
		t.Fatalf("expected name %q, got %q", "Product Updated", updated.Name)
	}

	if updated.Price != 25.0 {
		t.Fatalf("expected price %v, got %v", 25.0, updated.Price)
	}
}

func TestProductUseCase_UpdateProduct_InvalidName(t *testing.T) {
	repo := newFakeProductRepo()
	uc := NewProductUseCase(repo)

	p, _ := entity.NewProduct("Product", 10)
	_ = repo.Create(p)

	empty := "   "
	_, err := uc.UpdateProduct(p.ID, dto.UpdateProductInput{Name: &empty})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, entity.ErrRequired) {
		t.Fatalf("expected ErrRequired, got: %v", err)
	}
}

func TestProductUseCase_UpdateProduct_InvalidPrice_Negative(t *testing.T) {
	repo := newFakeProductRepo()
	uc := NewProductUseCase(repo)

	p, _ := entity.NewProduct("Product", 10)
	_ = repo.Create(p)

	neg := -1.0
	_, err := uc.UpdateProduct(p.ID, dto.UpdateProductInput{Price: &neg})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, entity.ErrInvalidPrice) {
		t.Fatalf("expected ErrInvalidPrice, got: %v", err)
	}
}

func TestProductUseCase_UpdateProduct_NotFound(t *testing.T) {
	repo := newFakeProductRepo()
	uc := NewProductUseCase(repo)

	newName := "X"
	_, err := uc.UpdateProduct(pkgentity.NewId(), dto.UpdateProductInput{Name: &newName})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
	}
}

func TestProductUseCase_DeleteProduct_Success(t *testing.T) {
	repo := newFakeProductRepo()
	uc := NewProductUseCase(repo)

	p, _ := entity.NewProduct("Product", 10)
	_ = repo.Create(p)

	if err := uc.DeleteProduct(p.ID); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	_, err := uc.GetProduct(p.ID)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
	}
}
