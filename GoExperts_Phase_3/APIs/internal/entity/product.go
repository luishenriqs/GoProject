// internal/entity/product.go
package entity

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	pkgentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/pkg/entity"
)

// Erros específicos deste agregado.
// (Reutilizamos ErrRequired já definido em user.go no mesmo pacote.)
var (
	ErrInvalidID    = errors.New("invalid id")
	ErrInvalidPrice = errors.New("invalid price")
)

// Product representa um produto do domínio.
type Product struct {
	ID        pkgentity.ID `json:"id"`
	Name      string       `json:"name"`
	Price     float64      `json:"price"`
	CreatedAt time.Time    `json:"created_at"`
}

// NewProduct cria um novo produto atribuindo ID e CreatedAt automaticamente.
// Mantido sem validação automática para não mudar comportamento sem autorização.
func NewProduct(name string, price float64) (*Product, error) {
	product := &Product{
		ID:        pkgentity.NewId(),
		Name:      name,
		Price:     price,
		CreatedAt: time.Now().UTC(),
	}

	err := product.validate()
	if err != nil {
		return nil, err
	}

	return product, nil
}

// validate executa validações de domínio para o Product.
// Regras:
// - id, name e price são required
// - id e price precisam ser válidos (id != uuid.Nil, price > 0)
func (p *Product) validate() error {
	if p == nil {
		return ErrRequired
	}
	// ID requerido e válido
	if p.ID == uuid.Nil {
		// Zero-value de uuid => inválido/ausente
		return ErrInvalidID
	}
	// Name requerido (não vazio após trim)
	if strings.TrimSpace(p.Name) == "" {
		return ErrRequired
	}
	// Price requerido e válido (> 0)
	if p.Price <= 0 {
		if p.Price == 0 {
			return ErrRequired
		}
		return ErrInvalidPrice
	}
	return nil
}
