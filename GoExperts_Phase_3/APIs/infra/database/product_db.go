// infra/database/product_db.go
package database

import (
	appentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/entity"
	pkgentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/pkg/entity"
	"gorm.io/gorm"
)

// TO DO - CRIAR OS HEADERS DAS FUNÇÕES DESTA CAMADA
// TO DO - CRIAR OS HEADERS DAS FUNÇÕES DESTA CAMADA
// TO DO - CRIAR OS HEADERS DAS FUNÇÕES DESTA CAMADA
// TO DO - CRIAR OS HEADERS DAS FUNÇÕES DESTA CAMADA
// TO DO - CRIAR OS HEADERS DAS FUNÇÕES DESTA CAMADA
// TO DO - CRIAR OS HEADERS DAS FUNÇÕES DESTA CAMADA
// TO DO - CRIAR OS HEADERS DAS FUNÇÕES DESTA CAMADA
// TO DO - CRIAR OS HEADERS DAS FUNÇÕES DESTA CAMADA
// TO DO - CRIAR OS HEADERS DAS FUNÇÕES DESTA CAMADA
// TO DO - CRIAR OS HEADERS DAS FUNÇÕES DESTA CAMADA
// TO DO - CRIAR OS HEADERS DAS FUNÇÕES DESTA CAMADA

// Garantia de conformidade com a interface em tempo de compilação.
var _ ProductInterface = (*ProductDB)(nil)

type ProductDB struct {
	DB *gorm.DB
}

// Injeta a conexão do GORM (db) e retorna o repositório pronto para uso.
func NewProductDB(db *gorm.DB) *ProductDB {
	return &ProductDB{DB: db}
}

// Create persiste um novo produto.
func (r *ProductDB) Create(product *appentity.Product) error {
	return r.DB.Create(product).Error
}

func (r *ProductDB) FindByID(id pkgentity.ID) (*appentity.Product, error) {
	var p appentity.Product
	if err := r.DB.First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

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

// Update: NUNCA cria registro novo. Retorna gorm.ErrRecordNotFound se não existir.
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
