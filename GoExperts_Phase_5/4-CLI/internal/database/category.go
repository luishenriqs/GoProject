package database

import (
	"database/sql"

	"github.com/google/uuid"
)

/*
Este arquivo define a camada de acesso a dados para a entidade Category.

Responsabilidade:
- Centralizar a estrutura `Category` que encapsula a dependência do banco (`*sql.DB`) e os
  campos do registro.
- Expor funções/métodos para criar e persistir categorias na tabela `categories`, isolando
  detalhes de SQL do restante da aplicação.
*/

type Category struct {
	db          *sql.DB
	ID          string
	Name        string
	Description string
}

func NewCategory(db *sql.DB) *Category {
	return &Category{db: db}
}

/*
Create cria e persiste uma nova categoria na tabela `categories`.

Responsabilidade:
- Gerar um ID único para a categoria usando UUID.
- Inserir o registro no banco de dados através de `c.db.Exec`.
- Retornar a categoria criada com os dados persistidos.

Parâmetros:
- name: nome da categoria.
- description: descrição da categoria.

Retorno:
- Category: struct preenchida com `ID`, `Name` e `Description` da categoria criada.
- error: erro retornado pela operação de INSERT caso ocorra; em caso de sucesso, nil.
*/
func (c *Category) Create(name string, description string) (Category, error) {
	id := uuid.New().String()

	_, err := c.db.Exec("INSERT INTO categories (id, name, description) VALUES ($1, $2, $3)",
		id, name, description)
	if err != nil {
		return Category{}, err
	}

	return Category{ID: id, Name: name, Description: description}, nil
}

func (c *Category) FindAll() ([]Category, error) {
	rows, err := c.db.Query("SELECT id, name, description FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []Category{}
	for rows.Next() {
		var id, name, description string
		if err := rows.Scan(&id, &name, &description); err != nil {
			return nil, err
		}

		categories = append(categories, Category{ID: id, Name: name, Description: description})
	}

	return categories, nil
}

func (c *Category) FindByCourseID(courseID string) (Category, error) {
	var id, name, description string

	err := c.db.QueryRow(
		"SELECT c.id, c.name, c.description FROM categories c JOIN courses co ON c.id = co.category_id WHERE co.id = $1",
		courseID,
	).Scan(&id, &name, &description)

	if err != nil {
		return Category{}, err
	}

	return Category{ID: id, Name: name, Description: description}, nil
}

func (c *Category) Find(id string) (Category, error) {
	var name, description string

	err := c.db.QueryRow(
		"SELECT name, description FROM categories WHERE id = $1",
		id,
	).Scan(&name, &description)

	if err != nil {
		return Category{}, err
	}

	return Category{
		ID:          id,
		Name:        name,
		Description: description,
	}, nil
}
