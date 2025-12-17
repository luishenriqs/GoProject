// internal/infra/database/user_db.go
package database

import (
	appentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/entity"
	"gorm.io/gorm"
)

// Garantia de conformidade com a interface em tempo de compilação.
// Se algum método exigido pela interface faltar/assinar diferente, o projeto não compila.
var _ UserInterface = (*User)(nil)

type User struct {
	DB *gorm.DB
}

// Injeta a conexão do GORM (db) e retorna o repositório pronto para uso.
func NewUserDb(db *gorm.DB) *User {
	return &User{DB: db}
}

// Create persiste um novo usuário.
// Observação: espera receber appentity.User já validado/normalizado.
func (r *User) Create(user *appentity.User) error {
	return r.DB.Create(user).Error
}

// FindByEmail retorna o usuário pelo e-mail.
// Retorna erro do GORM; use errors.Is(err, gorm.ErrRecordNotFound) para testar "não encontrado".
func (r *User) FindByEmail(email string) (*appentity.User, error) {
	var u appentity.User
	if err := r.DB.Where("email = ?", email).First(&u).Error; err != nil { // Erro diferente de vazio?
		return nil, err // Retorna o erro
	}
	return &u, nil // Retorna o usuário
}
