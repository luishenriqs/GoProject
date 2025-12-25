package database

import (
	appentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/entity"
	pkgentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/pkg/entity"
)

// UserInterface define o contrato de persistência para User.
// Nesta etapa, expandimos apenas o necessário para suportar /users/me:
// - FindByID
// - Update (sem criar se não existir)
// - Delete
type UserInterface interface {
	Create(user *appentity.User) error
	FindByEmail(email string) (*appentity.User, error)

	FindByID(id pkgentity.ID) (*appentity.User, error)
	Update(user *appentity.User) error
	Delete(id pkgentity.ID) error
}

// ProductInterface permanece inalterada nesta etapa.
type ProductInterface interface {
	Create(product *appentity.Product) error
	FindByID(id pkgentity.ID) (*appentity.Product, error)
	FindAll(page, limit int, sort string) ([]appentity.Product, int64, error)
	Update(product *appentity.Product) error
	Delete(id pkgentity.ID) error
}
