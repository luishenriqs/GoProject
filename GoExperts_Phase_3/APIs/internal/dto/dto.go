package dto

// Este arquivo contém apenas DTOs (Data Transfer Objects) usados na camada HTTP.
// A responsabilidade aqui é representar payloads de entrada/saída da API,
// sem validações de regra de negócio (isso pertence ao domínio/usecases).

// CreateProductInput representa o payload de criação de um produto.
// (Já existia no estado atual e deve permanecer para compatibilidade.)
type CreateProductInput struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// CreateUserInput representa o payload de criação de usuário.
type CreateUserInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginInput representa o payload de autenticação.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UpdateUserInput representa o payload de atualização do usuário autenticado.
// Campos ponteiro permitem diferenciar "campo ausente" de "campo presente" no JSON.
type UpdateUserInput struct {
	Name     *string `json:"name,omitempty"`
	Password *string `json:"password,omitempty"`
}

// UpdateProductInput representa o payload de atualização de produto.
// Campos ponteiro permitem diferenciar "campo ausente" de "campo presente" no JSON.
type UpdateProductInput struct {
	Name  *string  `json:"name,omitempty"`
	Price *float64 `json:"price,omitempty"`
}
