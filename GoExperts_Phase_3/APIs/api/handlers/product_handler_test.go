package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/api/middleware"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/dto"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/entity"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/usecase"
	pkgentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/pkg/entity"
	"gorm.io/gorm"
)

// fakeProductRepo em memória
type fakeProductRepo struct {
	products map[pkgentity.ID]*entity.Product
}

func newFakeProductRepo() *fakeProductRepo {
	return &fakeProductRepo{
		products: make(map[pkgentity.ID]*entity.Product),
	}
}

func (r *fakeProductRepo) Create(product *entity.Product) error {
	r.products[product.ID] = product
	return nil
}

func (r *fakeProductRepo) FindByID(id pkgentity.ID) (*entity.Product, error) {
	p := r.products[id]
	if p == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return p, nil
}

func (r *fakeProductRepo) FindAll(page, limit int, sort string) ([]entity.Product, int64, error) {
	items := make([]entity.Product, 0, len(r.products))
	for _, p := range r.products {
		items = append(items, *p)
	}
	return items, int64(len(items)), nil
}

func (r *fakeProductRepo) Update(product *entity.Product) error {
	if _, ok := r.products[product.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	r.products[product.ID] = product
	return nil
}

func (r *fakeProductRepo) Delete(id pkgentity.ID) error {
	if _, ok := r.products[id]; !ok {
		return gorm.ErrRecordNotFound
	}
	delete(r.products, id)
	return nil
}

func makeBearerTokenForTests(t *testing.T, tokenAuth *jwtauth.JWTAuth, userID pkgentity.ID) string {
	t.Helper()

	exp := time.Now().Add(1 * time.Hour).Unix()
	_, tokenString, err := tokenAuth.Encode(map[string]interface{}{
		"sub": userID.String(),
		"exp": exp,
	})
	if err != nil {
		t.Fatalf("expected no error encoding token, got: %v", err)
	}
	return tokenString
}

func TestProductHandler_CreateProduct_NoAuth_401(t *testing.T) {
	repo := newFakeProductRepo()
	uc := usecase.NewProductUseCase(repo)
	h := NewProductHandler(uc)

	body, _ := json.Marshal(dto.CreateProductInput{Name: "P1", Price: 10})
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	h.CreateProduct(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestProductHandler_CreateProduct_BadJSON_400(t *testing.T) {
	repo := newFakeProductRepo()
	uc := usecase.NewProductUseCase(repo)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	h := NewProductHandler(uc)

	protected := middleware.AuthMiddleware(tokenAuth)(http.HandlerFunc(h.CreateProduct))

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString("{invalid"))
	req.Header.Set("Authorization", "Bearer "+makeBearerTokenForTests(t, tokenAuth, pkgentity.NewId()))
	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestProductHandler_CreateProduct_Success_201(t *testing.T) {
	repo := newFakeProductRepo()
	uc := usecase.NewProductUseCase(repo)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	h := NewProductHandler(uc)

	protected := middleware.AuthMiddleware(tokenAuth)(http.HandlerFunc(h.CreateProduct))

	body, _ := json.Marshal(dto.CreateProductInput{Name: "P1", Price: 10})
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+makeBearerTokenForTests(t, tokenAuth, pkgentity.NewId()))
	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d", http.StatusCreated, rec.Code)
	}
}

func TestProductHandler_GetProductByID_InvalidID_400(t *testing.T) {
	repo := newFakeProductRepo()
	uc := usecase.NewProductUseCase(repo)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	h := NewProductHandler(uc)

	protected := middleware.AuthMiddleware(tokenAuth)(http.HandlerFunc(h.GetProductByID))

	req := httptest.NewRequest(http.MethodGet, "/products/invalid", nil)
	req.Header.Set("Authorization", "Bearer "+makeBearerTokenForTests(t, tokenAuth, pkgentity.NewId()))
	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestProductHandler_GetProductByID_NotFound_404(t *testing.T) {
	repo := newFakeProductRepo()
	uc := usecase.NewProductUseCase(repo)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	h := NewProductHandler(uc)

	protected := middleware.AuthMiddleware(tokenAuth)(http.HandlerFunc(h.GetProductByID))

	id := pkgentity.NewId()
	req := httptest.NewRequest(http.MethodGet, "/products/"+id.String(), nil)
	req.Header.Set("Authorization", "Bearer "+makeBearerTokenForTests(t, tokenAuth, pkgentity.NewId()))
	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestProductHandler_ListProducts_Success_200(t *testing.T) {
	repo := newFakeProductRepo()
	uc := usecase.NewProductUseCase(repo)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	h := NewProductHandler(uc)

	p1, _ := entity.NewProduct("A", 10)
	_ = repo.Create(p1)

	protected := middleware.AuthMiddleware(tokenAuth)(http.HandlerFunc(h.ListProducts))

	req := httptest.NewRequest(http.MethodGet, "/products?page=1&limit=10&sort=asc", nil)
	req.Header.Set("Authorization", "Bearer "+makeBearerTokenForTests(t, tokenAuth, pkgentity.NewId()))
	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestProductHandler_UpdateProduct_Success_200(t *testing.T) {
	repo := newFakeProductRepo()
	uc := usecase.NewProductUseCase(repo)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	h := NewProductHandler(uc)

	p, _ := entity.NewProduct("A", 10)
	_ = repo.Create(p)

	protected := middleware.AuthMiddleware(tokenAuth)(http.HandlerFunc(h.UpdateProduct))

	newName := "B"
	body, _ := json.Marshal(dto.UpdateProductInput{Name: &newName})

	req := httptest.NewRequest(http.MethodPut, "/products/"+p.ID.String(), bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+makeBearerTokenForTests(t, tokenAuth, pkgentity.NewId()))
	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestProductHandler_DeleteProduct_NotFound_404(t *testing.T) {
	repo := newFakeProductRepo()
	uc := usecase.NewProductUseCase(repo)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	h := NewProductHandler(uc)

	protected := middleware.AuthMiddleware(tokenAuth)(http.HandlerFunc(h.DeleteProduct))

	id := pkgentity.NewId()
	req := httptest.NewRequest(http.MethodDelete, "/products/"+id.String(), nil)
	req.Header.Set("Authorization", "Bearer "+makeBearerTokenForTests(t, tokenAuth, pkgentity.NewId()))
	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestProductHandler_DeleteProduct_Success_204(t *testing.T) {
	repo := newFakeProductRepo()
	uc := usecase.NewProductUseCase(repo)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	h := NewProductHandler(uc)

	p, _ := entity.NewProduct("A", 10)
	_ = repo.Create(p)

	protected := middleware.AuthMiddleware(tokenAuth)(http.HandlerFunc(h.DeleteProduct))

	req := httptest.NewRequest(http.MethodDelete, "/products/"+p.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+makeBearerTokenForTests(t, tokenAuth, pkgentity.NewId()))
	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d", http.StatusNoContent, rec.Code)
	}
}
