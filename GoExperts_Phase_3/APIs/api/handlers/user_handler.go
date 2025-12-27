package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/api/middleware"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/dto"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/entity"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/usecase"
	"gorm.io/gorm"
)

type UserHandler struct {
	UseCase      *usecase.UserUseCase
	TokenAuth    *jwtauth.JWTAuth
	JWTExpiresIn int // em segundos
}

/*
NewUserHandler constrói e retorna uma instância de UserHandler, injetando as dependências
necessárias para executar os handlers de usuário e produzir tokens JWT quando aplicável.

Fluxo:
 1. Instancia um UserHandler preenchendo:
    - UseCase: caso de uso responsável pela lógica de usuário (uc).
    - TokenAuth: configuração/instância de JWTAuth usada para gerar/validar tokens (tokenAuth).
    - JWTExpiresIn: tempo de expiração do token (em segundos) utilizado pelo handler (jwtExpiresIn).
 2. Retorna o ponteiro para o UserHandler criado.

Parâmetros:
  - uc: ponteiro para usecase.UserUseCase (camada de aplicação) usado pelos handlers.
  - tokenAuth: ponteiro para jwtauth.JWTAuth usado para operações com JWT.
  - jwtExpiresIn: duração de expiração do JWT (em segundos) utilizada pelo handler.

Retorno:
  - *UserHandler com suas dependências configuradas.
*/
func NewUserHandler(uc *usecase.UserUseCase, tokenAuth *jwtauth.JWTAuth, jwtExpiresIn int) *UserHandler {
	return &UserHandler{
		UseCase:      uc,
		TokenAuth:    tokenAuth,
		JWTExpiresIn: jwtExpiresIn,
	}
}

// userResponse é o formato seguro de retorno (sem password).
type userResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

/*
toUserResponse converte uma entidade de domínio (*entity.User) para o formato de resposta
(userResponse) exposto pela camada HTTP, selecionando apenas os campos públicos.

Fluxo:
 1. Mapeia os campos do usuário:
    - ID: converte u.ID para string via u.ID.String().
    - Name: copia u.Name.
    - Email: copia u.Email.
 2. Retorna o userResponse resultante.

Parâmetros:
  - u: ponteiro para entity.User contendo os dados do usuário.

Retorno:
  - userResponse com ID (string), Name e Email.
*/
func toUserResponse(u *entity.User) userResponse {
	return userResponse{
		ID:    u.ID.String(),
		Name:  u.Name,
		Email: u.Email,
	}
}

/*
CreateUser trata a requisição HTTP de criação de usuário (POST /users), realizando o parse do
corpo JSON, delegando a criação/validação ao caso de uso e retornando o usuário criado em JSON.

Fluxo:
 1. Decodifica o JSON do corpo da requisição em dto.CreateUserInput.
    - Se falhar ao decodificar, responde com http.StatusBadRequest (400) e retorna.
 2. Chama o caso de uso h.UseCase.RegisterUser(input) para criar e persistir o usuário.
    - Se ocorrer erro de validação (entity.ErrRequired, entity.ErrInvalidEmail, entity.ErrWeakPassword),
    responde com http.StatusBadRequest (400) e retorna.
    - Para qualquer outro erro, responde com http.StatusInternalServerError (500) e retorna.
 3. Em caso de sucesso:
    - Define o header "Content-Type" como "application/json".
    - Responde com http.StatusCreated (201).
    - Codifica e escreve a resposta JSON usando toUserResponse(u).
    (o resultado do encode é ignorado com "_ = ...").

Parâmetros:
  - w: http.ResponseWriter para escrita do status, headers e body.
  - r: *http.Request contendo o corpo JSON com os dados do usuário.

Respostas HTTP:
  - 201 (Created): usuário criado com sucesso; body contém userResponse em JSON.
  - 400 (Bad Request): JSON inválido ou erro de validação de entrada.
  - 500 (Internal Server Error): falha inesperada ao criar/persistir o usuário.

Efeitos colaterais:
  - Lê r.Body para decodificar o payload.
  - Persiste o usuário via camada de caso de uso/repositório.
*/
// CreateUser godoc
// @Summary      Cria um novo usuário
// @Description  Registra um usuário com nome, email e senha. Retorna o usuário (sem password).
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        body body dto.CreateUserInput true "Dados do usuário"
// @Success      201 {object} userResponse
// @Failure      400 {string} string
// @Failure      500 {string} string
// @Router       /users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input dto.CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u, err := h.UseCase.RegisterUser(input)
	if err != nil {
		if errors.Is(err, entity.ErrRequired) ||
			errors.Is(err, entity.ErrInvalidEmail) ||
			errors.Is(err, entity.ErrWeakPassword) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(toUserResponse(u))
}

/*
Login trata a requisição HTTP de autenticação (POST /login), validando as credenciais via
caso de uso e, em caso de sucesso, emitindo um JWT e retornando-o no formato JSON.

Fluxo:
 1. Decodifica o JSON do corpo da requisição em dto.LoginInput.
    - Se falhar ao decodificar, responde com http.StatusBadRequest (400) e retorna.
 2. Chama o caso de uso h.UseCase.Login(input.Email, input.Password) para autenticar.
    - Se retornar usecase.ErrInvalidCredentials ou gorm.ErrRecordNotFound, responde com
    http.StatusUnauthorized (401) e retorna (evita vazar existência de usuário).
    - Para qualquer outro erro, responde com http.StatusInternalServerError (500) e retorna.
 3. Em caso de sucesso, calcula o timestamp de expiração:
    - exp = agora + JWTExpiresIn (em segundos), convertido para Unix().
 4. Gera o JWT via h.TokenAuth.Encode, definindo claims:
    - "sub": userID.String()
    - "exp": exp
    - Se falhar ao gerar o token, responde com http.StatusInternalServerError (500) e retorna.
 5. Retorna a resposta:
    - Define "Content-Type" como "application/json".
    - Codifica {"access_token": tokenString} no body (o erro do encode é ignorado com "_ = ...").

Parâmetros:
  - w: http.ResponseWriter para escrita do status, headers e body.
  - r: *http.Request contendo o corpo JSON com email e senha.

Respostas HTTP:
  - 200 (OK): credenciais válidas; body contém {"access_token": "<jwt>"}.
  - 400 (Bad Request): JSON inválido no corpo da requisição.
  - 401 (Unauthorized): credenciais inválidas ou usuário inexistente (tratado como inválido).
  - 500 (Internal Server Error): erro inesperado ao autenticar ou ao emitir o token.

Efeitos colaterais:
  - Lê r.Body para decodificar o payload.
  - Emite um JWT usando TokenAuth, com expiração baseada em JWTExpiresIn.
*/
// Login godoc
// @Summary      Autentica e retorna um JWT
// @Description  Valida credenciais e retorna um access_token JWT.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body dto.LoginInput true "Credenciais"
// @Success      200 {object} map[string]string
// @Failure      400 {string} string
// @Failure      401 {string} string
// @Failure      500 {string} string
// @Router       /login [post]
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input dto.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := h.UseCase.Login(input.Email, input.Password)
	if err != nil {
		// Tratamos not found como 401 para não vazar existência de usuário.
		if errors.Is(err, usecase.ErrInvalidCredentials) || errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	exp := time.Now().Add(time.Duration(h.JWTExpiresIn) * time.Second).Unix()
	_, tokenString, err := h.TokenAuth.Encode(map[string]interface{}{
		"sub": userID.String(),
		"exp": exp,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"access_token": tokenString,
	})
}

/*
GetMe trata a requisição HTTP para obter os dados do usuário autenticado (GET /users/me).
Ele extrai o userID do contexto (preenchido pelo middleware de autenticação), busca o usuário
via caso de uso e retorna os dados públicos em JSON.

Fluxo:
 1. Extrai o userID do contexto da requisição via middleware.UserIDFromContext(r.Context()).
    - Se não houver userID no contexto, responde com http.StatusUnauthorized (401) e retorna.
 2. Busca o usuário via h.UseCase.GetMe(userID).
    - Se o erro for gorm.ErrRecordNotFound, responde com http.StatusNotFound (404) e retorna.
    - Para qualquer outro erro, responde com http.StatusInternalServerError (500) e retorna.
 3. Em caso de sucesso:
    - Define "Content-Type" como "application/json".
    - Codifica e escreve a resposta JSON usando toUserResponse(u).
    (o resultado do encode é ignorado com "_ = ...").

Parâmetros:
  - w: http.ResponseWriter para escrita do status, headers e body.
  - r: *http.Request contendo o contexto autenticado.

Respostas HTTP:
  - 200 (OK): usuário encontrado; body contém userResponse em JSON.
  - 401 (Unauthorized): ausência de userID no contexto (requisição não autenticada).
  - 404 (Not Found): usuário não encontrado para o userID informado.
  - 500 (Internal Server Error): falha inesperada ao buscar o usuário.

Dependências:
  - middleware.UserIDFromContext para extrair o ID do usuário autenticado do contexto.
*/
// GetMe godoc
// @Summary      Retorna o usuário autenticado
// @Description  Obtém os dados públicos do usuário autenticado com base no token JWT (claim "sub").
// @Tags         Users
// @Produce      json
// @Success      200 {object} userResponse
// @Failure      401 {string} string
// @Failure      404 {string} string
// @Failure      500 {string} string
// @Security     BearerAuth
// @Router       /users/me [get]
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	u, err := h.UseCase.GetMe(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toUserResponse(u))
}

/*
UpdateMe trata a requisição HTTP de atualização parcial do usuário autenticado (PUT /users/me).
Ele extrai o userID do contexto, decodifica o payload JSON com campos opcionais e delega a
atualização ao caso de uso, retornando os dados públicos do usuário em JSON.

Fluxo:
 1. Extrai o userID do contexto da requisição via middleware.UserIDFromContext(r.Context()).
    - Se não houver userID no contexto, responde com http.StatusUnauthorized (401) e retorna.
 2. Decodifica o JSON do corpo da requisição em dto.UpdateUserInput.
    - Se falhar ao decodificar, responde com http.StatusBadRequest (400) e retorna.
 3. Chama o caso de uso h.UseCase.UpdateMe(userID, input).
    - Se o erro for entity.ErrRequired ou entity.ErrWeakPassword, responde com 400.
    - Se o erro for gorm.ErrRecordNotFound, responde com 404.
    - Para qualquer outro erro, responde com 500.
 4. Em caso de sucesso:
    - Define "Content-Type" como "application/json".
    - Codifica e escreve a resposta JSON usando toUserResponse(u).
    (o resultado do encode é ignorado com "_ = ...").

Parâmetros:
  - w: http.ResponseWriter para escrita do status, headers e body.
  - r: *http.Request contendo o contexto autenticado e o corpo JSON com campos de atualização.

Respostas HTTP:
  - 200 (OK): usuário atualizado (ou no-op conforme lógica do caso de uso); body contém userResponse em JSON.
  - 400 (Bad Request): JSON inválido ou validação falhou (ex.: name vazio, senha fraca/vazia).
  - 401 (Unauthorized): ausência de userID no contexto (requisição não autenticada).
  - 404 (Not Found): usuário não encontrado para o userID informado.
  - 500 (Internal Server Error): falha inesperada ao atualizar o usuário.

Dependências:
  - middleware.UserIDFromContext para extrair o ID do usuário autenticado do contexto.
*/
// UpdateMe godoc
// @Summary      Atualiza o usuário autenticado
// @Description  Atualiza parcialmente os dados do usuário autenticado (rota protegida).
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        body body dto.UpdateUserInput true "Campos para atualização"
// @Success      200 {object} userResponse
// @Failure      400 {string} string
// @Failure      401 {string} string
// @Failure      404 {string} string
// @Failure      500 {string} string
// @Security     BearerAuth
// @Router       /users/me [put]
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var input dto.UpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u, err := h.UseCase.UpdateMe(userID, input)
	if err != nil {
		if errors.Is(err, entity.ErrRequired) || errors.Is(err, entity.ErrWeakPassword) {
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
	_ = json.NewEncoder(w).Encode(toUserResponse(u))
}

/*
DeleteMe trata a requisição HTTP para remover o usuário autenticado (DELETE /users/me).
Ele extrai o userID do contexto e delega a remoção ao caso de uso, respondendo com o status
adequado conforme o resultado.

Fluxo:
 1. Extrai o userID do contexto da requisição via middleware.UserIDFromContext(r.Context()).
    - Se não houver userID no contexto, responde com http.StatusUnauthorized (401) e retorna.
 2. Chama o caso de uso h.UseCase.DeleteMe(userID) para remover o usuário.
    - Se o erro for gorm.ErrRecordNotFound, responde com http.StatusNotFound (404) e retorna.
    - Para qualquer outro erro, responde com http.StatusInternalServerError (500) e retorna.
 3. Em caso de sucesso, responde com http.StatusNoContent (204) sem body.

Parâmetros:
  - w: http.ResponseWriter para escrita do status, headers e body.
  - r: *http.Request contendo o contexto autenticado.

Respostas HTTP:
  - 204 (No Content): usuário removido com sucesso.
  - 401 (Unauthorized): ausência de userID no contexto (requisição não autenticada).
  - 404 (Not Found): usuário não encontrado para o userID informado.
  - 500 (Internal Server Error): falha inesperada ao remover o usuário.

Dependências:
  - middleware.UserIDFromContext para extrair o ID do usuário autenticado do contexto.
*/
// DeleteMe godoc
// @Summary      Remove o usuário autenticado
// @Description  Remove o usuário autenticado (rota protegida).
// @Tags         Users
// @Success      204 {string} string
// @Failure      401 {string} string
// @Failure      404 {string} string
// @Failure      500 {string} string
// @Security     BearerAuth
// @Router       /users/me [delete]
func (h *UserHandler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := h.UseCase.DeleteMe(userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
