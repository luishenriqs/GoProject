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

/*
NewProduct cria uma nova instância de Product preenchendo os campos essenciais,
definindo CreatedAt em UTC e executando a validação do objeto antes de retorná-lo.

Fluxo:
 1. Instancia um Product com:
    - ID gerado via pkgentity.NewId()
    - Name e Price conforme parâmetros recebidos
    - CreatedAt definido como time.Now().UTC()
 2. Executa product.validate().
    - Se houver erro, retorna (nil, err).
 3. Retorna o product criado em caso de sucesso.

Parâmetros:
  - name: nome do produto.
  - price: preço do produto.

Retorno:
  - (*Product, nil) em caso de sucesso.
  - (nil, err) se a validação do produto falhar.
*/
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

/*
validate valida a integridade mínima de um Product, garantindo presença e consistência
dos campos essenciais (ID, Name e Price).

Fluxo:
 1. Se o receiver for nil, retorna ErrRequired.
 2. Valida o ID:
    - Se p.ID for uuid.Nil (zero-value), retorna ErrInvalidID.
 3. Valida o Name:
    - Se strings.TrimSpace(p.Name) resultar em string vazia, retorna ErrRequired.
 4. Valida o Price:
    - Se p.Price <= 0:
    - Se p.Price == 0, retorna ErrRequired.
    - Se p.Price < 0, retorna ErrInvalidPrice.
 5. Retorna nil se todas as validações passarem.

Retorno:
  - nil se o Product estiver válido.
  - ErrRequired se o Product for nil, se Name estiver vazio, ou se Price for zero.
  - ErrInvalidID se o ID estiver ausente/inválido (uuid.Nil).
  - ErrInvalidPrice se o Price for negativo.
*/
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
