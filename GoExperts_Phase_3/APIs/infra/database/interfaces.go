// internal/infra/database/interfaces.go
package database

import (
	appentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/entity"
	pkgentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/pkg/entity"
)

// UserInterface define o contrato mínimo para persistência/consulta de usuários.
type UserInterface interface {
	Create(user *appentity.User) error
	FindByEmail(email string) (*appentity.User, error)
}

// ProductInterface define operações de CRUD e listagem paginada.
type ProductInterface interface {
	Create(product *appentity.Product) error
	FindByID(id pkgentity.ID) (*appentity.Product, error)
	// FindAll retorna a lista paginada e o total de registros (para paginação).
	FindAll(page, limit int) ([]appentity.Product, int64, error)
	Update(product *appentity.Product) error
	Delete(id pkgentity.ID) error
}
