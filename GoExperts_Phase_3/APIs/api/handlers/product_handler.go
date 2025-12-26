package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/api/middleware"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/dto"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/entity"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/usecase"
	pkgentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/pkg/entity"
	"gorm.io/gorm"
)

// ProductHandler concentra endpoints HTTP relacionados a Product.
// Todas as rotas de Product são protegidas, então exigimos userID no contexto.
type ProductHandler struct {
	UseCase *usecase.ProductUseCase
}

func NewProductHandler(uc *usecase.ProductUseCase) *ProductHandler {
	return &ProductHandler{UseCase: uc}
}

/*
CreateProduct trata a requisição HTTP de criação de produto (POST /products), exigindo
autenticação, realizando o parse do corpo JSON, delegando a criação/validação ao caso
de uso e retornando o produto criado em JSON.

Fluxo:
 1. Verifica autenticação extraindo o userID do contexto via middleware.UserIDFromContext.
    - Se não houver userID no contexto, responde com http.StatusUnauthorized (401) e retorna.
 2. Decodifica o JSON do corpo da requisição em dto.CreateProductInput.
    - Se falhar ao decodificar, responde com http.StatusBadRequest (400) e retorna.
 3. Chama o caso de uso h.UseCase.CreateProduct(input).
    - Se o erro for entity.ErrRequired ou entity.ErrInvalidPrice, responde com 400 e retorna.
    - Para qualquer outro erro, responde com http.StatusInternalServerError (500) e retorna.
 4. Em caso de sucesso:
    - Define "Content-Type" como "application/json".
    - Responde com http.StatusCreated (201).
    - Codifica e escreve o produto no body em JSON (o erro do encode é ignorado com "_ = ...").

Parâmetros:
  - w: http.ResponseWriter para escrita do status, headers e body.
  - r: *http.Request contendo o contexto autenticado e o corpo JSON com os dados do produto.

Respostas HTTP:
  - 201 (Created): produto criado com sucesso; body contém o produto em JSON.
  - 400 (Bad Request): JSON inválido ou erro de validação (campos obrigatórios/preço inválido).
  - 401 (Unauthorized): ausência de userID no contexto (requisição não autenticada).
  - 500 (Internal Server Error): falha inesperada ao criar/persistir o produto.

Dependências:
  - middleware.UserIDFromContext para validar autenticação via contexto.
*/
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	if _, ok := middleware.UserIDFromContext(r.Context()); !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var input dto.CreateProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p, err := h.UseCase.CreateProduct(input)
	if err != nil {
		if errors.Is(err, entity.ErrRequired) || errors.Is(err, entity.ErrInvalidPrice) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(p)
}

/*
ListProducts trata a requisição HTTP de listagem paginada de produtos (GET /products),
exigindo autenticação e suportando paginação e ordenação via query string.

Fluxo:
 1. Verifica autenticação extraindo o userID do contexto via middleware.UserIDFromContext.
    - Se não houver userID no contexto, responde com http.StatusUnauthorized (401) e retorna.
 2. Lê parâmetros de query:
    - page: convertido de string para int via strconv.Atoi (erros são ignorados, resultando em 0).
    - limit: convertido de string para int via strconv.Atoi (erros são ignorados, resultando em 0).
    - sort: lido diretamente (string).
 3. Chama o caso de uso h.UseCase.ListProducts(page, limit, sort).
    - Se ocorrer erro, responde com http.StatusInternalServerError (500) e retorna.
 4. Em caso de sucesso:
    - Define "Content-Type" como "application/json".
    - Retorna um JSON com:
    - "items": lista de produtos (paginada)
    - "total": total de produtos existentes (sem paginação)
    - (o erro do encode é ignorado com "_ = ...").

Parâmetros (query string):
  - page: número da página (int). Se ausente/ inválido, será 0 e a normalização ocorre no use case/repositório.
  - limit: quantidade de itens por página (int). Se ausente/ inválido, será 0 e a normalização ocorre no use case/repositório.
  - sort: direção de ordenação (string), tipicamente "asc" ou "desc" (validação/normalização ocorre no use case/repositório).

Respostas HTTP:
  - 200 (OK): listagem retornada com sucesso; body contém {"items": [...], "total": <n>}.
  - 401 (Unauthorized): ausência de userID no contexto (requisição não autenticada).
  - 500 (Internal Server Error): falha inesperada ao listar produtos.

Dependências:
  - middleware.UserIDFromContext para validar autenticação via contexto.
  - strconv.Atoi para parsing de page e limit.
*/
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	if _, ok := middleware.UserIDFromContext(r.Context()); !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	sort := q.Get("sort")

	items, total, err := h.UseCase.ListProducts(page, limit, sort)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"items": items,
		"total": total,
	})
}

/*
GetProductByID trata a requisição HTTP para obter um produto pelo ID (GET /products/{id}),
exigindo autenticação, validando/parsing o ID a partir do path e retornando o produto em JSON.

Fluxo:
 1. Verifica autenticação extraindo o userID do contexto via middleware.UserIDFromContext.
    - Se não houver userID no contexto, responde com http.StatusUnauthorized (401) e retorna.
 2. Extrai o ID do path removendo o prefixo "/products/" de r.URL.Path.
    - Se o ID estiver vazio ou contiver "/" (indicando path inválido), responde com http.StatusBadRequest (400) e retorna.
 3. Converte o ID de string para pkgentity.ID via pkgentity.ParseId(idStr).
    - Se falhar, responde com http.StatusBadRequest (400) e retorna.
 4. Busca o produto via h.UseCase.GetProduct(id).
    - Se o erro for gorm.ErrRecordNotFound, responde com http.StatusNotFound (404) e retorna.
    - Para qualquer outro erro, responde com http.StatusInternalServerError (500) e retorna.
 5. Em caso de sucesso:
    - Define "Content-Type" como "application/json".
    - Codifica e escreve o produto no body em JSON (o erro do encode é ignorado com "_ = ...").

Parâmetros:
  - w: http.ResponseWriter para escrita do status, headers e body.
  - r: *http.Request contendo o contexto autenticado e o path com o ID do produto.

Respostas HTTP:
  - 200 (OK): produto encontrado; body contém o produto em JSON.
  - 400 (Bad Request): path inválido ou ID inválido.
  - 401 (Unauthorized): ausência de userID no contexto (requisição não autenticada).
  - 404 (Not Found): produto não encontrado para o ID informado.
  - 500 (Internal Server Error): falha inesperada ao buscar o produto.

Observação:
  - A extração do ID é feita via manipulação direta de r.URL.Path (TrimPrefix) e validação simples
    para rejeitar paths com segmentos adicionais.
*/
func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	if _, ok := middleware.UserIDFromContext(r.Context()); !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/products/")
	if idStr == "" || strings.Contains(idStr, "/") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := pkgentity.ParseId(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p, err := h.UseCase.GetProduct(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(p)
}

/*
UpdateProduct trata a requisição HTTP de atualização parcial de um produto (PUT /products/{id}),
exigindo autenticação, validando/parsing o ID a partir do path, decodificando o payload JSON
e delegando a atualização ao caso de uso, retornando o produto atualizado em JSON.

Fluxo:
 1. Verifica autenticação extraindo o userID do contexto via middleware.UserIDFromContext.
    - Se não houver userID no contexto, responde com http.StatusUnauthorized (401) e retorna.
 2. Extrai o ID do path removendo o prefixo "/products/" de r.URL.Path.
    - Se o ID estiver vazio ou contiver "/" (indicando path inválido), responde com http.StatusBadRequest (400) e retorna.
 3. Converte o ID de string para pkgentity.ID via pkgentity.ParseId(idStr).
    - Se falhar, responde com http.StatusBadRequest (400) e retorna.
 4. Decodifica o JSON do corpo da requisição em dto.UpdateProductInput.
    - Se falhar ao decodificar, responde com http.StatusBadRequest (400) e retorna.
 5. Chama o caso de uso h.UseCase.UpdateProduct(id, input).
    - Se o erro for entity.ErrRequired ou entity.ErrInvalidPrice, responde com 400 e retorna.
    - Se o erro for gorm.ErrRecordNotFound, responde com 404 e retorna.
    - Para qualquer outro erro, responde com http.StatusInternalServerError (500) e retorna.
 6. Em caso de sucesso:
    - Define "Content-Type" como "application/json".
    - Codifica e escreve o produto atualizado no body em JSON (o erro do encode é ignorado com "_ = ...").

Parâmetros:
  - w: http.ResponseWriter para escrita do status, headers e body.
  - r: *http.Request contendo o contexto autenticado, o path com o ID do produto e o corpo JSON com campos de atualização.

Respostas HTTP:
  - 200 (OK): produto atualizado (ou no-op conforme lógica do caso de uso); body contém o produto em JSON.
  - 400 (Bad Request): path/ID inválido, JSON inválido ou erro de validação (campos obrigatórios/preço inválido).
  - 401 (Unauthorized): ausência de userID no contexto (requisição não autenticada).
  - 404 (Not Found): produto não encontrado para o ID informado.
  - 500 (Internal Server Error): falha inesperada ao atualizar/persistir o produto.

Observação:
  - A extração do ID é feita via manipulação direta de r.URL.Path (TrimPrefix) e validação simples
    para rejeitar paths com segmentos adicionais.
*/
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	if _, ok := middleware.UserIDFromContext(r.Context()); !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/products/")
	if idStr == "" || strings.Contains(idStr, "/") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := pkgentity.ParseId(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var input dto.UpdateProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p, err := h.UseCase.UpdateProduct(id, input)
	if err != nil {
		if errors.Is(err, entity.ErrRequired) || errors.Is(err, entity.ErrInvalidPrice) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(p)
}

/*
DeleteProduct trata a requisição HTTP para remover um produto pelo ID (DELETE /products/{id}),
exigindo autenticação, validando/parsing o ID a partir do path e delegando a remoção ao caso
de uso, respondendo com o status adequado conforme o resultado.

Fluxo:
 1. Verifica autenticação extraindo o userID do contexto via middleware.UserIDFromContext.
    - Se não houver userID no contexto, responde com http.StatusUnauthorized (401) e retorna.
 2. Extrai o ID do path removendo o prefixo "/products/" de r.URL.Path.
    - Se o ID estiver vazio ou contiver "/" (indicando path inválido), responde com http.StatusBadRequest (400) e retorna.
 3. Converte o ID de string para pkgentity.ID via pkgentity.ParseId(idStr).
    - Se falhar, responde com http.StatusBadRequest (400) e retorna.
 4. Chama o caso de uso h.UseCase.DeleteProduct(id) para remover o produto.
    - Se o erro for gorm.ErrRecordNotFound, responde com http.StatusNotFound (404) e retorna.
    - Para qualquer outro erro, responde com http.StatusInternalServerError (500) e retorna.
 5. Em caso de sucesso, responde com http.StatusNoContent (204) sem body.

Parâmetros:
  - w: http.ResponseWriter para escrita do status, headers e body.
  - r: *http.Request contendo o contexto autenticado e o path com o ID do produto.

Respostas HTTP:
  - 204 (No Content): produto removido com sucesso.
  - 400 (Bad Request): path inválido ou ID inválido.
  - 401 (Unauthorized): ausência de userID no contexto (requisição não autenticada).
  - 404 (Not Found): produto não encontrado para o ID informado.
  - 500 (Internal Server Error): falha inesperada ao remover o produto.

Observação:
  - A extração do ID é feita via manipulação direta de r.URL.Path (TrimPrefix) e validação simples
    para rejeitar paths com segmentos adicionais.
*/
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	if _, ok := middleware.UserIDFromContext(r.Context()); !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/products/")
	if idStr == "" || strings.Contains(idStr, "/") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := pkgentity.ParseId(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.UseCase.DeleteProduct(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
