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

// User é o repositório GORM de User.
// (Mantido exatamente o nome/estrutura do estado atual.)
type User struct {
	DB *gorm.DB
}

func NewUserDb(db *gorm.DB) *User {
	return &User{DB: db}
}

var _ UserInterface = (*User)(nil)

func (u *User) Create(user *appentity.User) error {
	return u.DB.Create(user).Error
}

func (u *User) FindByEmail(email string) (*appentity.User, error) {
	var user appentity.User
	err := u.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID busca um usuário por ID.
func (u *User) FindByID(id pkgentity.ID) (*appentity.User, error) {
	var user appentity.User
	err := u.DB.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update atualiza um usuário existente.
// Semântica: NÃO cria registro se não existir (igual ao ProductDB.Update).
func (u *User) Update(user *appentity.User) error {
	result := u.DB.Model(&appentity.User{}).
		Where("id = ?", user.ID).
		Updates(user)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// Delete remove um usuário por ID.
// Para manter previsibilidade de CRUD, retorna ErrRecordNotFound quando não existir.
func (u *User) Delete(id pkgentity.ID) error {
	result := u.DB.Delete(&appentity.User{}, "id = ?", id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
