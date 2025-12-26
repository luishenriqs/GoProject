// infra/database/product_db.go
package database

import (
	appentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/entity"
	pkgentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/pkg/entity"
	"gorm.io/gorm"
)

// Garantia de conformidade com a interface em tempo de compilação.
var _ ProductInterface = (*ProductDB)(nil)

type ProductDB struct {
	DB *gorm.DB
}

// Injeta a conexão do GORM (db) e retorna o repositório pronto para uso.
func NewProductDB(db *gorm.DB) *ProductDB {
	return &ProductDB{DB: db}
}

/*
Create persiste um novo produto no banco de dados utilizando o GORM.

Fluxo:
 1. Executa r.DB.Create(product).
 2. Retorna diretamente o campo Error da operação.

Parâmetros:
  - product: ponteiro para appentity.Product contendo os dados do produto a ser inserido.

Retorno:
  - nil em caso de sucesso.
  - err se o GORM falhar ao inserir o registro.
*/
func (r *ProductDB) Create(product *appentity.Product) error {
	return r.DB.Create(product).Error
}

/*
FindByID busca um produto no banco de dados pelo seu ID e retorna a entidade encontrada.

Fluxo:
 1. Declara uma variável local p do tipo appentity.Product.
 2. Executa a consulta com r.DB.First(&p, "id = ?", id).
    - Se ocorrer erro, retorna (nil, err) (incluindo gorm.ErrRecordNotFound quando aplicável).
 3. Retorna (&p, nil) em caso de sucesso.

Parâmetros:
  - id: identificador do produto (pkgentity.ID) utilizado como filtro na consulta.

Retorno:
  - (*appentity.Product, nil) se o produto for encontrado.
  - (nil, err) se ocorrer erro na consulta (incluindo gorm.ErrRecordNotFound se não existir registro).
*/
func (r *ProductDB) FindByID(id pkgentity.ID) (*appentity.Product, error) {
	var p appentity.Product
	if err := r.DB.First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

/*
FindAll busca produtos de forma paginada no banco de dados, retornando a lista de itens,
o total de registros existentes (sem paginação) e um erro, quando aplicável.

Fluxo:
 1. Calcula o total de produtos existentes usando COUNT no model appentity.Product.
    - Se falhar, retorna (nil, 0, err).
 2. Normaliza parâmetros de paginação:
    - Se page <= 0, assume page = 1.
    - Se limit <= 0, assume limit = 10.
    - Calcula offset = (page - 1) * limit.
 3. Normaliza o parâmetro sort:
    - Se sort for diferente de ""/"asc"/"desc", força sort = "asc".
    - Se sort for "", mantém o comportamento atual (ordenação "created_at " + sort).
 4. Executa a consulta:
    - Ordena por "created_at" concatenado com sort.
    - Aplica Limit e Offset.
    - Carrega os itens em []appentity.Product.
    - Se falhar, retorna (nil, 0, err).
 5. Retorna (items, total, nil) em caso de sucesso.

Parâmetros:
  - page: número da página (1-based). Valores <= 0 são normalizados para 1.
  - limit: quantidade de itens por página. Valores <= 0 são normalizados para 10.
  - sort: direção de ordenação para created_at. Aceita "asc", "desc" ou "".
    Valores inválidos são normalizados para "asc".

Retorno:
  - ([]appentity.Product, total, nil) em caso de sucesso, onde total é a contagem total
    de produtos existentes (antes da paginação).
  - (nil, 0, err) se ocorrer erro ao contar ou ao consultar os itens.
*/
func (r *ProductDB) FindAll(page, limit int, sort string) ([]appentity.Product, int64, error) {
	var (
		items []appentity.Product
		total int64
	)
	if err := r.DB.Model(&appentity.Product{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	if sort != "" && sort != "asc" && sort != "desc" {
		sort = "asc"
	}

	if err := r.DB.
		Order("created_at " + sort).
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

/*
Update atualiza um produto existente no banco de dados com base no ID presente em product.ID.

Fluxo:
 1. Monta uma operação de update no model appentity.Product filtrando por "id = ?".
 2. Executa Updates(product), delegando ao GORM a atualização dos campos conforme a struct recebida.
 3. Se tx.Error estiver preenchido, retorna o erro.
 4. Se tx.RowsAffected for 0, retorna gorm.ErrRecordNotFound (nenhum registro correspondeu ao ID).
 5. Retorna nil em caso de sucesso.

Parâmetros:
  - product: ponteiro para appentity.Product contendo o ID do registro a ser atualizado e os
    campos a serem enviados para o update conforme o comportamento do GORM.

Retorno:
  - nil em caso de sucesso.
  - gorm.ErrRecordNotFound se nenhum registro foi atualizado (ID inexistente).
  - err se ocorrer falha na execução do update pelo GORM.
*/
func (r *ProductDB) Update(product *appentity.Product) error {
	tx := r.DB.Model(&appentity.Product{}).
		Where("id = ?", product.ID).
		Updates(product)

	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *ProductDB) Delete(id pkgentity.ID) error {
	return r.DB.Delete(&appentity.Product{}, "id = ?", id).Error
}
