package graph

import "github.com/luishenriqs/GoProject/GoExperts_Phase_5/1-GraphQL/internal/database"

/*
Este arquivo define o ponto central de injeção de dependências da camada GraphQL.

Responsabilidade:
- Declarar a struct `Resolver`, que concentra as dependências necessárias para os resolvers
  (queries e mutations) executarem operações na aplicação.
- Expor as dependências (ex.: `CategoryDB`) para que o código gerado pelo gqlgen e os
  resolvers concretos consigam acessar a camada de persistência/serviços sem acoplamento
  direto à criação dessas instâncias.

Uso:
- A struct `Resolver` é instanciada no bootstrap da aplicação (ex.: `server.go`) e passada
  para o handler GraphQL, permitindo configurar quais implementações serão usadas em runtime.
*/

type Resolver struct {
	CategoryDB *database.Category
}
