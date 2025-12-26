package database

import (
	appentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/entity"
	pkgentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/pkg/entity"
	"gorm.io/gorm"
)

// User é o repositório GORM de User.
// (Mantido exatamente o nome/estrutura do estado atual.)
type User struct {
	DB *gorm.DB
}

func NewUserDb(db *gorm.DB) *User {
	return &User{DB: db}
}

var _ UserInterface = (*User)(nil)

/*
Create persiste um novo usuário no banco de dados utilizando o GORM.

Fluxo:
 1. Executa u.DB.Create(user).
 2. Retorna diretamente o campo Error da operação.

Parâmetros:
  - user: ponteiro para appentity.User contendo os dados do usuário a ser inserido.

Retorno:
  - nil em caso de sucesso.
  - err se o GORM falhar ao inserir o registro.
*/
func (u *User) Create(user *appentity.User) error {
	return u.DB.Create(user).Error
}

/*
FindByEmail busca um usuário no banco de dados pelo e-mail e retorna a entidade encontrada.

Fluxo:
 1. Declara uma variável local user do tipo appentity.User.
 2. Executa a consulta filtrando por "email = ?" via u.DB.Where(...).First(&user).
    - Se ocorrer erro, retorna (nil, err) (incluindo gorm.ErrRecordNotFound quando aplicável).
 3. Retorna (&user, nil) em caso de sucesso.

Parâmetros:
  - email: e-mail utilizado como filtro na consulta.

Retorno:
  - (*appentity.User, nil) se o usuário for encontrado.
  - (nil, err) se ocorrer erro na consulta (incluindo gorm.ErrRecordNotFound se não existir registro).
*/
func (u *User) FindByEmail(email string) (*appentity.User, error) {
	var user appentity.User
	err := u.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

/*
FindByID busca um usuário no banco de dados pelo seu ID e retorna a entidade encontrada.

Fluxo:
 1. Declara uma variável local user do tipo appentity.User.
 2. Executa a consulta com u.DB.First(&user, "id = ?", id).
    - Se ocorrer erro, retorna (nil, err) (incluindo gorm.ErrRecordNotFound quando aplicável).
 3. Retorna (&user, nil) em caso de sucesso.

Parâmetros:
  - id: identificador do usuário (pkgentity.ID) utilizado como filtro na consulta.

Retorno:
  - (*appentity.User, nil) se o usuário for encontrado.
  - (nil, err) se ocorrer erro na consulta (incluindo gorm.ErrRecordNotFound se não existir registro).
*/
func (u *User) FindByID(id pkgentity.ID) (*appentity.User, error) {
	var user appentity.User
	err := u.DB.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

/*
Update atualiza um usuário existente no banco de dados com base no ID presente em user.ID.

Fluxo:
 1. Monta uma operação de update no model appentity.User filtrando por "id = ?".
 2. Executa Updates(user), delegando ao GORM a atualização dos campos conforme a struct recebida.
 3. Se result.Error estiver preenchido, retorna o erro.
 4. Se result.RowsAffected for 0, retorna gorm.ErrRecordNotFound (nenhum registro correspondeu ao ID).
 5. Retorna nil em caso de sucesso.

Parâmetros:
  - user: ponteiro para appentity.User contendo o ID do registro a ser atualizado e os
    campos a serem enviados para o update conforme o comportamento do GORM.

Retorno:
  - nil em caso de sucesso.
  - gorm.ErrRecordNotFound se nenhum registro foi atualizado (ID inexistente).
  - err se ocorrer falha na execução do update pelo GORM.
*/
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

/*
Delete remove um usuário do banco de dados pelo seu ID.

Fluxo:
 1. Executa a operação de delete via u.DB.Delete(&appentity.User{}, "id = ?", id).
 2. Se result.Error estiver preenchido, retorna o erro.
 3. Se result.RowsAffected for 0, retorna gorm.ErrRecordNotFound (nenhum registro correspondeu ao ID).
 4. Retorna nil em caso de sucesso.

Parâmetros:
  - id: identificador do usuário (pkgentity.ID) do registro a ser removido.

Retorno:
  - nil em caso de sucesso.
  - gorm.ErrRecordNotFound se nenhum registro foi removido (ID inexistente).
  - err se ocorrer falha na execução do delete pelo GORM.
*/
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
