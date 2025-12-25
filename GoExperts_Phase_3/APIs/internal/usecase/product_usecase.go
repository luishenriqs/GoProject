package usecase

import (
	"strings"

	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/infra/database"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/dto"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/entity"
	pkgentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/pkg/entity"
)

// ProductUseCase concentra regras de aplicação/orquestração de Product.
type ProductUseCase struct {
	ProductRepo database.ProductInterface
}

// NewProductUseCase cria o usecase de Product injetando a interface do repositório.
func NewProductUseCase(repo database.ProductInterface) *ProductUseCase {
	return &ProductUseCase{ProductRepo: repo}
}

/*
CreateProduct cria e persiste um novo produto a partir de um payload de criação,
delegando a validação e a inicialização de campos padrão (como ID e CreatedAt)
para a camada de entidade, e persistindo o resultado via repositório.

Fluxo:
 1. Cria a entidade de produto chamando entity.NewProduct(input.Name, input.Price).
    - Se a criação/validação falhar, retorna (nil, err).
 2. Persiste o produto no repositório via uc.ProductRepo.Create(p).
    - Se a persistência falhar, retorna (nil, err).
 3. Retorna a entidade criada em caso de sucesso.

Parâmetros:
  - input: dto.CreateProductInput contendo Name e Price.

Retorno:
  - (*entity.Product, nil) em caso de sucesso.
  - (nil, err) se a criação/validação do produto falhar ou se a persistência no repositório falhar.
*/
func (uc *ProductUseCase) CreateProduct(input dto.CreateProductInput) (*entity.Product, error) {
	p, err := entity.NewProduct(input.Name, input.Price)
	if err != nil {
		return nil, err
	}

	if err := uc.ProductRepo.Create(p); err != nil {
		return nil, err
	}

	return p, nil
}

// GetProduct carrega um produto por ID.
func (uc *ProductUseCase) GetProduct(id pkgentity.ID) (*entity.Product, error) {
	return uc.ProductRepo.FindByID(id)
}

// ListProducts retorna produtos paginados e o total.
func (uc *ProductUseCase) ListProducts(page, limit int, sort string) ([]entity.Product, int64, error) {
	return uc.ProductRepo.FindAll(page, limit, sort)
}

/*
UpdateProduct atualiza um produto existente (identificado por id) aplicando alterações
parciais (patch) em Name e/ou Price e persistindo o resultado no repositório.

Fluxo:
 1. Busca o produto no repositório por ID via uc.ProductRepo.FindByID(id).
    - Se falhar, retorna (nil, err).
 2. Se input não contém nenhum campo (Name == nil e Price == nil), não altera nada (no-op)
    e retorna o produto carregado.
 3. Se Name foi fornecido:
    - Normaliza com strings.TrimSpace.
    - Se ficar vazio, retorna entity.ErrRequired.
    - Caso contrário, atualiza p.Name.
 4. Se Price foi fornecido:
    - Se price == 0, retorna entity.ErrRequired.
    - Se price < 0, retorna entity.ErrInvalidPrice.
    - Caso contrário, atualiza p.Price.
 5. Persiste as alterações via uc.ProductRepo.Update(p).
    - Se falhar, retorna (nil, err).
 6. Retorna o produto atualizado.

Parâmetros:
  - id: ID do produto a ser atualizado.
  - input: dto.UpdateProductInput com campos opcionais (ponteiros) para atualização parcial:
  - Name: *string (opcional)
  - Price: *float64 (opcional)

Retorno:
  - (*entity.Product, nil) em caso de sucesso (incluindo no-op quando nada é enviado).
  - (nil, entity.ErrRequired) se Name for enviado e resultar em string vazia após trim,
    ou se Price for enviado e for igual a zero.
  - (nil, entity.ErrInvalidPrice) se Price for enviado e for negativo.
  - (nil, err) para falhas de busca por ID ou falhas ao persistir no repositório.
*/
func (uc *ProductUseCase) UpdateProduct(id pkgentity.ID, input dto.UpdateProductInput) (*entity.Product, error) {
	p, err := uc.ProductRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// No-op se nada foi enviado.
	if input.Name == nil && input.Price == nil {
		return p, nil
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, entity.ErrRequired
		}
		p.Name = name
	}

	if input.Price != nil {
		price := *input.Price
		if price == 0 {
			return nil, entity.ErrRequired
		}
		if price < 0 {
			return nil, entity.ErrInvalidPrice
		}
		p.Price = price
	}

	if err := uc.ProductRepo.Update(p); err != nil {
		return nil, err
	}

	return p, nil
}

// DeleteProduct remove um produto por ID.
func (uc *ProductUseCase) DeleteProduct(id pkgentity.ID) error {
	return uc.ProductRepo.Delete(id)
}
