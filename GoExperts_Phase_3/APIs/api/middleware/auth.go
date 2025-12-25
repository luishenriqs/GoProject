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

// UserIDFromContext recupera o userID injetado pelo middleware.
// Retorna (id, true) se existir, ou (zero, false) caso contrário.
func UserIDFromContext(ctx context.Context) (pkgentity.ID, bool) {
	v := ctx.Value(userIDContextKey{})
	if v == nil {
		return pkgentity.ID{}, false
	}

	id, ok := v.(pkgentity.ID)
	return id, ok
}

// AuthMiddleware valida o JWT (Authorization: Bearer <token>) e injeta o userID no contexto.
// - tokenAuth deve ser o mesmo inicializado em configs (cfg.TokenAuth).
// - Em caso de token ausente ou inválido: responde 401 e NÃO chama o próximo handler.
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
