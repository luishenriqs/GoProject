package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/api/middleware"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/dto"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/entity"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/usecase"
	pkgentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/pkg/entity"
	"gorm.io/gorm"
)

// ProductHandler concentra endpoints HTTP relacionados a Product.
// Todas as rotas de Product são protegidas, então exigimos userID no contexto.
type ProductHandler struct {
	UseCase *usecase.ProductUseCase
}

func NewProductHandler(uc *usecase.ProductUseCase) *ProductHandler {
	return &ProductHandler{UseCase: uc}
}

// CreateProduct atende POST /products (protegido).
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	if _, ok := middleware.UserIDFromContext(r.Context()); !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var input dto.CreateProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p, err := h.UseCase.CreateProduct(input)
	if err != nil {
		if errors.Is(err, entity.ErrRequired) || errors.Is(err, entity.ErrInvalidPrice) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(p)
}

// ListProducts atende GET /products?page=&limit=&sort=asc|desc (protegido).
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	if _, ok := middleware.UserIDFromContext(r.Context()); !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	sort := q.Get("sort")

	items, total, err := h.UseCase.ListProducts(page, limit, sort)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"items": items,
		"total": total,
	})
}

// GetProductByID atende GET /products/{id} (protegido).
func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	if _, ok := middleware.UserIDFromContext(r.Context()); !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/products/")
	if idStr == "" || strings.Contains(idStr, "/") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := pkgentity.ParseId(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p, err := h.UseCase.GetProduct(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(p)
}

// UpdateProduct atende PUT /products/{id} (protegido).
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	if _, ok := middleware.UserIDFromContext(r.Context()); !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/products/")
	if idStr == "" || strings.Contains(idStr, "/") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := pkgentity.ParseId(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var input dto.UpdateProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p, err := h.UseCase.UpdateProduct(id, input)
	if err != nil {
		if errors.Is(err, entity.ErrRequired) || errors.Is(err, entity.ErrInvalidPrice) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(p)
}

// DeleteProduct atende DELETE /products/{id} (protegido).
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	if _, ok := middleware.UserIDFromContext(r.Context()); !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/products/")
	if idStr == "" || strings.Contains(idStr, "/") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := pkgentity.ParseId(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.UseCase.DeleteProduct(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
