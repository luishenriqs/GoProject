package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/jwtauth/v5"
	pkgentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/pkg/entity"
)

// userIDContextKey é a chave interna para armazenar o userID no contexto.
// Usamos um tipo privado para evitar colisões com outras chaves.
type userIDContextKey struct{}

/*
UserIDFromContext recupera o ID do usuário autenticado armazenado no contexto da requisição
pelo middleware de autenticação, retornando o ID e um booleano indicando sucesso.

Fluxo:
 1. Lê o valor do contexto usando a chave tipada userIDContextKey{}.
 2. Se o valor não existir (nil), retorna (pkgentity.ID{}, false).
 3. Faz type assertion para pkgentity.ID.
    - Retorna (id, true) se o tipo for compatível.
    - Caso contrário, retorna (pkgentity.ID{}, false).

Parâmetros:
  - ctx: contexto de onde o userID será extraído.

Retorno:
  - (pkgentity.ID, true) quando um pkgentity.ID válido estiver presente no contexto.
  - (pkgentity.ID{}, false) quando não houver valor para a chave ou o tipo não for pkgentity.ID.
*/
func UserIDFromContext(ctx context.Context) (pkgentity.ID, bool) {
	v := ctx.Value(userIDContextKey{})
	if v == nil {
		return pkgentity.ID{}, false
	}

	id, ok := v.(pkgentity.ID)
	return id, ok
}

/*
AuthMiddleware cria um middleware HTTP responsável por autenticar requisições via JWT no header
Authorization, validar o token usando tokenAuth e injetar o userID (claim "sub") no contexto
da requisição para uso pelos handlers.

Fluxo:
 1. Retorna uma função middleware com assinatura func(next http.Handler) http.Handler.
 2. Para cada requisição:
    a) Lê o header "Authorization".
    - Se estiver vazio, responde com http.StatusUnauthorized (401) e retorna.
    b) Valida o formato esperado "Bearer <token>":
    - Divide em 2 partes com strings.SplitN(authHeader, " ", 2).
    - Exige exatamente 2 partes, prefixo "Bearer" (case-insensitive) e token não vazio.
    - Se inválido, responde com 401 e retorna.
    c) Extrai e normaliza o token (trim).
    d) Verifica assinatura/validade do token via jwtauth.VerifyToken(tokenAuth, rawToken).
    - Se err != nil ou token == nil, responde com 401 e retorna.
    e) Obtém a claim "sub" via token.Get("sub").
    - Se inexistente, responde com 401 e retorna.
    f) Valida o tipo e conteúdo do "sub":
    - Deve ser string não vazia após trim.
    - Se inválido, responde com 401 e retorna.
    g) Converte o "sub" para pkgentity.ID via pkgentity.ParseId(sub).
    - Se falhar, responde com 401 e retorna.
    h) Insere o userID no contexto da requisição via context.WithValue,
    usando a chave tipada userIDContextKey{}.
    i) Chama o próximo handler com next.ServeHTTP(w, r.WithContext(ctx)).

Parâmetros:
  - tokenAuth: instância de jwtauth.JWTAuth usada para verificar tokens JWT.

Retorno:
  - Um middleware (func(next http.Handler) http.Handler) que:
  - Bloqueia requisições não autenticadas com 401.
  - Em requisições autenticadas, injeta o userID no contexto.

Respostas HTTP:
  - 401 (Unauthorized) para:
  - Authorization ausente.
  - Formato diferente de "Bearer <token>".
  - Token inválido/não verificável.
  - Claim "sub" ausente, com tipo inválido ou vazia.
  - "sub" que não pode ser convertido para pkgentity.ID.

Observações:
  - Esta implementação obtém claims via token.Get("sub"), pois o token da versão usada
    não expõe .Claims/.Valid diretamente.
  - O userID é armazenado no contexto para ser recuperado pelos handlers (ex.: UserIDFromContext).
*/
func AuthMiddleware(tokenAuth *jwtauth.JWTAuth) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Espera formato: "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			rawToken := strings.TrimSpace(parts[1])

			// Verifica assinatura/validade do token com o tokenAuth vindo do configs.
			token, err := jwtauth.VerifyToken(tokenAuth, rawToken)
			if err != nil || token == nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Nesta implementação do jwtauth/v5, o token não expõe .Claims/.Valid.
			// As claims são obtidas via token.Get("sub").
			subVal, ok := token.Get("sub")
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			sub, ok := subVal.(string)
			if !ok || strings.TrimSpace(sub) == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			userID, err := pkgentity.ParseId(sub)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey{}, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
