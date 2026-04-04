# README — Módulo 6-UOW

## Visão geral

Este módulo aprofunda a camada de persistência construída no módulo anterior e introduz dois conceitos centrais: **Repositories** e **UOW (Unit of Work)**.

A base continua sendo o uso do **sqlc** para gerar código Go a partir de SQL escrito manualmente. O avanço aqui é que as operações deixam de ser executadas apenas de forma isolada e passam a ser coordenadas por uma unidade transacional central, capaz de garantir **commit** ou **rollback** de um conjunto de ações relacionadas.

O módulo usa MySQL, schema simples com `categories` e `courses`, geração de código com `sqlc`, repositórios específicos por agregado e uma implementação própria de `Unit of Work` em `pkg/uow/uow.go`.

## Objetivo didático

O foco principal é demonstrar:

* como encapsular acesso ao banco em repositories
* como reaproveitar o código gerado pelo `sqlc`
* como injetar transações nesses repositories
* como usar um `Unit of Work` para coordenar múltiplas operações
* como evitar inconsistência quando duas ou mais operações fazem parte da mesma regra de negócio

## Árvore do módulo

```text
6-UOW/
├─ internal/
│  ├─ db/
│  │  ├─ db.go
│  │  ├─ models.go
│  │  └─ queries.sql.go
│  ├─ entity/
│  │  └─ entity.go
│  ├─ repository/
│  │  ├─ category.go
│  │  └─ course.go
│  └─ usecase/
│     ├─ add_course.go
│     ├─ add_course_test.go
│     ├─ add_course_uow.go
│     └─ add_course_uow_test.go
├─ pkg/
│  └─ uow/
│     └─ uow.go
├─ sql/
│  ├─ queries.sql
│  └─ schema.sql
├─ docker-compose.yaml
├─ go.mod
├─ go.sum
└─ sqlc.yaml
```

## Tecnologias utilizadas

### Go + `database/sql`

A base de acesso ao banco continua sendo a biblioteca padrão `database/sql`, com `*sql.DB`, `*sql.Tx`, `BeginTx`, `Commit` e `Rollback`.

### `sqlc`

O `sqlc` lê o schema em `sql/schema.sql` e as queries em `sql/queries.sql`, gerando código tipado em `internal/db`. Isso é configurado em `sqlc.yaml`.

### MySQL

O banco do módulo é MySQL 5.7, configurado no `docker-compose.yaml`, usando a base `courses`.

### Testify

Os testes usam `github.com/stretchr/testify/assert`.

## Schema do banco

O schema definido em `sql/schema.sql` é:

```sql
CREATE TABLE IF NOT EXISTS `categories` (
  id int PRIMARY KEY AUTO_INCREMENT,
  name varchar(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS `courses` (
  id int PRIMARY KEY AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  category_id INTEGER NOT NULL,
  FOREIGN KEY (category_id) REFERENCES categories(id)
);
```

A tabela `courses` depende de `categories` pela chave estrangeira `category_id`. Isso cria o cenário ideal para estudar transações e rollback.

## Queries SQL escritas manualmente

As queries ficam em `sql/queries.sql`:

```sql
-- name: CreateCategory :exec
INSERT INTO categories (id, name) VALUES (?, ?);

-- name: CreateCourse :exec
INSERT INTO courses (id, name, category_id) VALUES (?, ?, ?);
```

O `sqlc` usa os comentários `-- name:` para gerar métodos Go, especialmente `CreateCategory` e `CreateCourse`.

## Código gerado pelo `sqlc`

A saída gerada vai para `internal/db`.

### `internal/db/db.go`

Esse arquivo define:

* a interface `DBTX`
* a struct `Queries`
* o construtor `New`
* o método `WithTx`

A interface `DBTX` permite que a mesma struct `Queries` funcione tanto com `*sql.DB` quanto com `*sql.Tx`. Esse é um dos pontos centrais do módulo.

### `internal/db/models.go`

Contém os modelos gerados a partir do schema:

* `Category`
* `Course`

São structs focadas em persistência.

### `internal/db/queries.sql.go`

Contém os métodos gerados:

* `CreateCategory(ctx, arg)`
* `CreateCourse(ctx, arg)`

Esses métodos executam `INSERT` usando `ExecContext`.

## Camada de entidade

O arquivo `internal/entity/entity.go` define duas entidades simples:

```go
type Category struct {
    ID       int
    Name     string
    CourseID []int
}

func (c *Category) AddCourse(id int) {
    c.CourseID = append(c.CourseID, id)
}

type Course struct {
    ID         int
    Name       string
    CategoryID int
}
```

Aqui já existe separação entre:

* entidade de domínio
* modelo de persistência gerado pelo `sqlc`

## Camada de repository

O módulo cria dois repositories:

* `CategoryRepository`
* `CourseRepository`

### `CategoryRepository`

Arquivo: `internal/repository/category.go`

Responsabilidades:

* receber uma entidade `entity.Category`
* converter essa entidade para o formato exigido pelo `sqlc`
* delegar a inserção para `Queries.CreateCategory`

Estrutura:

```go
type CategoryRepository struct {
    DB      *sql.DB
    Queries *db.Queries
}
```

### `CourseRepository`

Arquivo: `internal/repository/course.go`

Responsabilidades:

* receber uma entidade `entity.Course`
* converter a entidade para `db.CreateCourseParams`
* chamar o método gerado pelo `sqlc`

Estrutura:

```go
type CourseRepository struct {
    DB      *sql.DB
    Queries *db.Queries
}
```

## Por que usar repositories

O repository:

* isola o restante da aplicação do `sqlc`
* centraliza a conversão entre entidade e persistência
* facilita troca de `Queries` com ou sem transação
* reduz o acoplamento do use case com SQL e banco

Com isso, o use case depende apenas de interfaces:

* `CategoryRepositoryInterface`
* `CourseRepositoryInterface`

## Use case sem UOW

Arquivo: `internal/usecase/add_course.go`

Esse use case recebe dois repositories:

```go
type AddCourseUseCase struct {
    CourseRepository   repository.CourseRepositoryInterface
    CategoryRepository repository.CategoryRepositoryInterface
}
```

Fluxo:

1. cria uma categoria
2. insere a categoria
3. cria um curso
4. insere o curso

### Problema deste modelo

As duas operações são independentes do ponto de vista transacional. Se a primeira inserção funcionar e a segunda falhar, o banco pode ficar inconsistente.

## Teste sem UOW

Arquivo: `internal/usecase/add_course_test.go`

Esse teste monta um cenário em que o curso aponta para uma categoria inexistente:

```go
CourseCategoryID: 2
```

enquanto o comentário sugere que a categoria criada seria `ID->1`.

A intenção didática é mostrar:

* categoria pode ser inserida
* curso falha por chave estrangeira
* sem transação, a categoria pode permanecer gravada

## Conceito de Unit of Work

O **Unit of Work** coordena um conjunto de alterações que devem ser tratadas como uma única unidade transacional.

Pergunta central:

> Essas operações devem ser confirmadas juntas ou desfeitas juntas?

Neste módulo, a unidade de trabalho é:

* inserir categoria
* inserir curso

## Implementação do UOW

Arquivo: `pkg/uow/uow.go`

### Estruturas principais

```go
type RepositoryFactory func(tx *sql.Tx) interface{}
```

Cada repository é registrado como uma fábrica que sabe construir sua versão transacional a partir de um `*sql.Tx`.

```go
type Uow struct {
    Db           *sql.DB
    Tx           *sql.Tx
    Repositories map[string]RepositoryFactory
}
```

O `Uow` guarda:

* conexão principal
* transação atual
* mapa de factories registradas

### Interface do UOW

```go
type UowInterface interface {
    Register(name string, fc RepositoryFactory)
    UnRegister(name string)
    GetRepository(ctx context.Context, name string) (interface{}, error)
    Do(ctx context.Context, fn func(uow *Uow) error) error
    Rollback() error
    CommitOrRollback() error
}
```

### Responsabilidade dos métodos

* `Register`: registra uma factory
* `UnRegister`: remove uma factory
* `GetRepository`: recupera uma instância associada à transação atual
* `Do`: executa uma função dentro de um escopo transacional
* `Rollback`: desfaz a transação atual
* `CommitOrRollback`: tenta confirmar; se falhar, tenta rollback

### Fluxo interno

#### `GetRepository`

Se não existir transação ativa, o método inicia uma:

```go
tx, err := u.Db.BeginTx(ctx, nil)
```

Depois usa a factory registrada:

```go
repo := u.Repositories[name](u.Tx)
```

#### `Do`

Esse é o coração do padrão:

1. verifica se já existe transação
2. inicia nova transaction
3. executa a função recebida
4. se houver erro, faz rollback
5. se não houver erro, chama `CommitOrRollback`

## Como os repositories entram no UOW

No teste `add_course_uow_test.go`, os repositories são registrados assim:

```go
uow.Register("CategoryRepository", func(tx *sql.Tx) interface{} {
    repo := repository.NewCategoryRepository(dbt)
    repo.Queries = db.New(tx)
    return repo
})
```

```go
uow.Register("CourseRepository", func(tx *sql.Tx) interface{} {
    repo := repository.NewCourseRepository(dbt)
    repo.Queries = db.New(tx)
    return repo
})
```

### Ponto central

O repository é criado normalmente, mas o campo `Queries` é substituído por:

```go
db.New(tx)
```

Assim, ele deixa de operar com `*sql.DB` e passa a operar com `*sql.Tx`.

Esse é o elo entre:

* repositories
* `sqlc`
* transaction
* Unit of Work

## Use case com UOW

Arquivo: `internal/usecase/add_course_uow.go`

Estrutura:

```go
type AddCourseUseCaseUow struct {
    Uow uow.UowInterface
}
```

Em vez de receber repositories diretamente, ele recebe um `UowInterface`.

### Fluxo do Execute

O método executa tudo dentro de:

```go
a.Uow.Do(ctx, func(uow *uow.Uow) error {
    ...
})
```

Dentro desse bloco:

1. monta a categoria
2. recupera `CategoryRepository` pelo UOW
3. insere categoria
4. monta o curso
5. recupera `CourseRepository` pelo UOW
6. insere curso
7. retorna erro ou sucesso

## Teste com UOW

Arquivo: `internal/usecase/add_course_uow_test.go`

Esse teste repete o mesmo cenário problemático:

```go
CourseCategoryID: 2
```

Agora o fluxo passa por `Uow.Do(...)`.

A intenção didática é demonstrar que:

* a categoria é inserida dentro da mesma transaction
* a tentativa de inserir o curso falha
* o UOW executa rollback
* o banco não fica com persistência parcial daquela unidade de negócio

## Relação entre `sqlc` e UOW

Este módulo mostra algo importante:

> O `sqlc` se integra muito bem com transações.

Isso acontece porque `db.New(...)` aceita um `DBTX`, e `DBTX` pode ser satisfeito por:

* `*sql.DB`
* `*sql.Tx`

Com isso, o mesmo código gerado pelo `sqlc` pode ser executado:

* fora de transação
* dentro de transação

sem duplicar queries nem criar outra camada de acesso a dados.

## Estratégia de transaction consolidada

### Sem UOW

* cada repository usa conexão normal
* as operações não estão agrupadas transacionalmente
* uma falha intermediária pode deixar persistência parcial

### Com UOW

* a transação é aberta antes da execução do caso de uso
* os repositories são resolvidos já vinculados ao `tx`
* todas as operações compartilham o mesmo contexto transacional
* qualquer erro dispara rollback
* só há persistência definitiva se tudo funcionar

## Vantagens do padrão Unit of Work

1. **Atomicidade**
   Garante que várias operações sejam persistidas juntas ou desfeitas juntas.

2. **Consistência**
   Evita estados intermediários inválidos no banco.

3. **Centralização de transação**
   `BeginTx`, `Commit` e `Rollback` não ficam espalhados pelos use cases.

4. **Reaproveitamento de repositories**
   Os mesmos repositories continuam sendo usados, apenas mudando a origem de `Queries`.

5. **Separação de responsabilidades**

   * use case define a regra de negócio
   * repository executa persistência
   * UOW coordena a transação

## Limitações e observações do estado atual

Como material de estudo, o módulo cumpre bem seu papel, mas alguns pontos ainda estão simplificados:

* `GetRepository` retorna `interface{}` e depende de type assertion
* os nomes dos repositories são strings literais
* os testes não validam explicitamente o estado final do banco após rollback
* há um detalhe didático entre `AUTO_INCREMENT` no schema e `INSERT` explicitando `id` nas queries
* existe `InputUseCaseUow`, mas o método `Execute` do use case UOW recebe `InputUseCase`

## Como rodar o ambiente

### Subir o MySQL com Docker

```bash
docker compose up -d
```

### Validar se o container está rodando

```bash
docker compose ps
```

### Banco disponível

* host: `localhost`
* porta: `3306`
* usuário: `root`
* senha: `root`
* database: `courses`

## Gerar código com `sqlc`

Sempre que `sql/schema.sql` ou `sql/queries.sql` forem alterados:

```bash
sqlc generate
```

Isso regenera os arquivos em:

```text
internal/db/
```

## Rodar os testes

```bash
go test ./...
```

Para rodar apenas os testes do use case:

```bash
go test ./internal/usecase -v
```

## O que foi aprendido neste módulo

### Sobre repositories

* como encapsular operações de persistência
* como desacoplar use cases do `sqlc`
* como trabalhar com interfaces

### Sobre UOW

* como representar uma unidade transacional explícita
* como abrir e fechar transaction em um único ponto
* como compartilhar a mesma transaction entre múltiplos repositories
* como garantir rollback em caso de falha

### Sobre `sqlc`

* como continuar usando SQL explícito
* como gerar código tipado
* como integrar esse código com `*sql.Tx` sem duplicação

### Sobre arquitetura

* domínio, repository, use case e infraestrutura ficam mais bem definidos
* a transação passa a ser um detalhe coordenado por uma camada específica
* o fluxo fica mais próximo de uma aplicação real com regras compostas

## Síntese final

Este módulo mostra a evolução natural da camada de persistência:

1. escrever queries SQL e gerar código com `sqlc`
2. encapsular essas operações em repositories
3. coordenar várias operações em uma única transaction com Unit of Work

O aprendizado principal é que **repositories isolam o acesso a dados** e **UOW garante integridade transacional quando uma regra de negócio envolve múltiplos passos**.

Em termos práticos, o módulo demonstra que:

* inserir categoria e curso de forma independente pode gerar inconsistência
* inserir ambos dentro de uma única unidade de trabalho evita persistência parcial
* `sqlc` se integra bem com esse modelo por meio da abstração `DBTX`

Esse é o valor central do módulo: mostrar como sair de operações isoladas e chegar a um fluxo transacional organizado, reutilizável e mais seguro.
