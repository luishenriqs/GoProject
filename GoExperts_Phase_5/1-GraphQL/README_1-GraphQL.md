# GoExperts_Phase_5/1-GraphQL (Go + gqlgen + SQLite)

Projeto de exemplo para estudar **GraphQL em Go** usando **gqlgen** e **SQLite**, com duas entidades:

- **Category** (categoria)
- **Course** (curso)

O foco do projeto é mostrar como o **gqlgen** gera *resolvers* específicos para **campos relacionais** e como esses campos são resolvidos **sob demanda** (só quando são solicitados na query), evitando consultas desnecessárias ao banco.

---

## Requisitos

- Go (projeto foi desenvolvido/testado com Go 1.19+)
- SQLite
- Dependências Go baixadas via `go mod`

---

## Como executar

1. Instale dependências:

```bash
go mod tidy
```

2. Rode o servidor (ajuste o comando conforme sua estrutura de pastas):

```bash
go run ./cmd/server
```

3. Acesse o GraphQL Playground:

- `http://localhost:<PORT>/`

> Se `PORT` não estiver definido, o servidor usa a porta padrão definida no código.

---

## Schema GraphQL (visão geral)

O schema expõe:

- Queries
  - `categories`: lista todas as categorias
  - `courses`: lista todos os cursos
- Mutations
  - `createCategory`: cria categoria
  - `createCourse`: cria curso

Além disso, existem **campos relacionais**:

- `Category.courses`: retorna os cursos daquela categoria
- `Course.category`: retorna a categoria do curso

Schema (resumo):

```graphql
type Category {
  id: ID!
  name: String!
  description: String!
  courses: [Course!]!
}

input NewCategory {
  name: String!
  description: String
}

type Course {
  id: ID!
  name: String!
  description: String!
  category: Category!
}

input NewCourse {
  name: String!
  description: String
  categoryId: ID!
}

type Mutation {
  createCategory(input: NewCategory!): Category!
  createCourse(input: NewCourse!): Course!
}

type Query {
  categories: [Category!]!
  courses: [Course!]!
}
```

---

## O que foi alterado e por quê (relacionamento Category ↔ Course)

Originalmente, `Category` tinha um array de `Courses` dentro do model, mas foi removido.  
Em seguida:

1. **Os models foram separados** (um arquivo para `Category` e outro para `Course`).
2. O relacionamento foi descrito no `schema.graphqls`:
   - `Category.courses: [Course!]!`
   - `Course.category: Category!`
3. Ao rodar o `gqlgen generate`, o gqlgen:
   - gerou resolvers dedicados (ex.: `categoryResolver`, `courseResolver`)
   - adicionou métodos *stub* “not implemented” para os campos relacionais

Isso permite resolver relacionamentos de forma **lazy**.

### Por que isso é importante (lazy loading)

No GraphQL, o resolver de um campo só é executado se aquele campo for pedido na query.

Exemplo: se você faz a query abaixo, o servidor **não** precisa buscar cursos no banco:

```graphql
query {
  categories {
    id
    name
    description
  }
}
```

Mas se você pedir `courses`, aí sim o resolver `Category.courses` roda e faz a consulta adicional:

```graphql
query {
  categories {
    id
    name
    courses {
      id
      name
    }
  }
}
```

O mesmo vale para `Course.category`: só será consultado quando você solicitar esse campo.

> Observação: esse padrão já ajuda bastante a evitar consultas desnecessárias, mas ao solicitar relacionamentos para muitos itens pode surgir o padrão **N+1**. Em produção, é comum introduzir *dataloaders* ou cache por request. Neste projeto, o objetivo é didático.

---

## Como as consultas ao banco são feitas (alto nível)

- `Query.categories` → `CategoryDB.FindAll()`
- `Query.courses` → `CourseDB.FindAll()`

Relacionamentos:

- `Category.courses` → `CourseDB.FindByCategoryID(categoryID)`
- `Course.category` → `CategoryDB.FindByCourseID(courseID)`

---

## Exemplos no Playground

### 1) Criar uma categoria

```graphql
mutation {
  createCategory(input: { name: "Backend", description: "Cursos de backend" }) {
    id
    name
    description
  }
}
```

### 2) Criar um curso (vinculado a uma categoria)

```graphql
mutation {
  createCourse(input: { name: "Go + GraphQL", description: "gqlgen na prática", categoryId: "<CATEGORY_ID>" }) {
    id
    name
    description
  }
}
```

### 3) Listar categorias (sem cursos)

```graphql
query {
  categories {
    id
    name
    description
  }
}
```

### 4) Listar categorias com cursos (executa o resolver `Category.courses`)

```graphql
query {
  categories {
    id
    name
    courses {
      id
      name
      description
    }
  }
}
```

### 5) Listar cursos (sem categoria)

```graphql
query {
  courses {
    id
    name
    description
  }
}
```

### 6) Listar cursos com categoria (executa o resolver `Course.category`)

```graphql
query {
  courses {
    id
    name
    description
    category {
      id
      name
      description
    }
  }
}
```

---

## Estrutura (arquivos principais)

- `schema.graphqls`: schema do GraphQL
- `gqlgen.yml`: configuração do gqlgen (mapeamento de models, etc.)
- `graph/generated/*`: código gerado pelo gqlgen
- `graph/schema.resolvers.go`: resolvers implementados
- `internal/database/*`: camada de acesso ao SQLite (`CategoryDB`, `CourseDB`)
- `cmd/server/*`: inicialização do servidor, abertura do banco e handlers HTTP

---

## Comandos úteis

Gerar código do gqlgen (quando o schema ou `gqlgen.yml` mudarem):

```bash
go run github.com/99designs/gqlgen generate
```
