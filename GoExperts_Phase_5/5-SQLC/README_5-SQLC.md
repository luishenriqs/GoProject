# README — Módulo 5-SQLC

## Visão geral

Este módulo demonstra como utilizar a biblioteca **sqlc** para transformar queries SQL escritas manualmente em código Go tipado, previsível e pronto para uso.

A proposta central aqui é manter o **SQL como fonte de verdade** da camada de persistência, enquanto o `sqlc` gera automaticamente structs, parâmetros e métodos Go a partir dos arquivos `.sql`.

Além disso, o módulo também consolida uma estratégia importante de **transactions**, mostrando como reutilizar os métodos gerados pelo `sqlc` tanto em operações simples quanto em operações atômicas envolvendo múltiplas queries.

---

## Estrutura do módulo

A árvore principal do módulo está organizada assim:

```text
5-SQLC/
├─ cmd/
│  ├─ runSQLC/
│  │  └─ main.go
│  └─ runSQLCTX/
│     └─ main.go
├─ internal/
│  └─ db/
│     ├─ db.go
│     ├─ models.go
│     └─ query.sql.go
├─ sql/
│  ├─ migrations/
│  │  ├─ 000001_init.up.sql
│  │  └─ 000001_init.down.sql
│  └─ queries/
│     └─ query.sql
├─ data.db
├─ docker-compose.yaml
├─ go.mod
├─ go.sum
├─ Makefile
└─ sqlc.yaml
```

---

## O que foi implementado neste módulo

Este módulo cobre quatro blocos principais:

1. **Criação do schema do banco** com migrations.
2. **Escrita manual das queries SQL** em arquivo dedicado.
3. **Geração automática do código Go** com `sqlc`.
4. **Execução de queries com e sem transaction**.

---

## 1) Schema do banco de dados

O schema foi definido com migrations.

### Arquivo

- `sql/migrations/000001_init.up.sql`

### Estrutura criada

#### Tabela `categories`

```sql
CREATE TABLE categories (
    id varchar(36) NOT NULL PRIMARY KEY,
    name text NOT NULL,
    description text
);
```

#### Tabela `courses`

```sql
CREATE TABLE courses (
    id varchar(36) NOT NULL PRIMARY KEY,
    category_id varchar(36) NOT NULL,
    name text NOT NULL,
    description text,
    price decimal(10,2) NOT NULL,
    FOREIGN KEY (category_id) REFERENCES categories(id)
);
```

### Observações

- Cada `course` pertence a uma `category`.
- A relação entre as tabelas é garantida por **foreign key** em `category_id`.
- O campo `price` foi incluído no curso para representar o valor do curso.

### Arquivo de rollback

- `sql/migrations/000001_init.down.sql`

```sql
DROP TABLE IF EXISTS courses;
DROP TABLE IF EXISTS categories;
```

Esse arquivo permite desfazer a migration, removendo primeiro `courses` e depois `categories`, respeitando a dependência entre as tabelas.

---

## 2) Queries SQL definidas manualmente

As operações do módulo foram escritas no arquivo:

- `sql/queries/query.sql`

### Queries de categoria

```sql
-- name: ListCategories :many
SELECT * FROM categories;

-- name: GetCategory :one
SELECT * FROM categories 
WHERE id = ?;

-- name: CreateCategory :exec
INSERT INTO categories (id, name, description) 
VALUES (?,?,?);

-- name: UpdateCategory :exec
UPDATE categories SET name = ?, description = ? 
WHERE id = ?;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = ?;
```

### Queries de curso

```sql
-- name: CreateCourse :exec
INSERT INTO courses (id, name, description, category_id, price)
VALUES (?,?,?,?,?);

-- name: ListCourses :many
SELECT c.*, ca.name as category_name 
FROM courses c JOIN categories ca ON c.category_id = ca.id;
```

### O que essas anotações significam

O padrão `-- name: NomeDaQuery :tipo` é interpretado pelo `sqlc`.

Exemplos:

- `:exec` → gera método para executar comando sem retorno de linhas.
- `:one` → gera método que retorna exatamente um registro.
- `:many` → gera método que retorna uma lista de registros.

Essas anotações são o elo entre o SQL escrito manualmente e o código Go gerado automaticamente.

---

## 3) Papel do `sqlc`

O `sqlc` lê:

- o schema/migrations,
- o arquivo de queries,
- e a configuração em `sqlc.yaml`,

para gerar código Go tipado dentro de `internal/db/`.

### Comando principal

```bash
sqlc generate
```

### O que foi gerado

Após a geração, os principais arquivos foram:

- `internal/db/db.go`
- `internal/db/models.go`
- `internal/db/query.sql.go`

---

## 4) Arquivos gerados pelo `sqlc`

### `internal/db/db.go`

Esse arquivo define a base de execução das queries.

#### Elementos principais

##### Interface `DBTX`

```go
type DBTX interface {
    ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
    PrepareContext(context.Context, string) (*sql.Stmt, error)
    QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
    QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}
```

Essa interface é extremamente importante porque permite que a struct `Queries` funcione tanto com:

- `*sql.DB`
- `*sql.Tx`

Isso significa que os métodos gerados pelo `sqlc` podem ser reutilizados dentro e fora de transações.

##### Construtor

```go
func New(db DBTX) *Queries {
    return &Queries{db: db}
}
```

##### Struct `Queries`

```go
type Queries struct {
    db DBTX
}
```

##### Método `WithTx`

```go
func (q *Queries) WithTx(tx *sql.Tx) *Queries {
    return &Queries{
        db: tx,
    }
}
```

Esse método reforça a ideia de reaproveitamento das mesmas queries dentro de um contexto transacional.

---

### `internal/db/models.go`

Esse arquivo contém os modelos Go gerados com base nas tabelas.

#### `Category`

```go
type Category struct {
    ID          string
    Name        string
    Description sql.NullString
}
```

#### `Course`

```go
type Course struct {
    ID          string
    CategoryID  string
    Name        string
    Description sql.NullString
    Price       float64
}
```

### Observação importante

O campo `description` foi mapeado como `sql.NullString`, o que mostra como o `sqlc` respeita corretamente colunas que podem ser nulas no banco.

---

### `internal/db/query.sql.go`

Esse é o arquivo mais importante gerado pelo `sqlc`, porque ele transforma cada query SQL em código Go executável.

#### Exemplos gerados

##### `CreateCategory`

```go
type CreateCategoryParams struct {
    ID          string
    Name        string
    Description sql.NullString
}

func (q *Queries) CreateCategory(ctx context.Context, arg CreateCategoryParams) error {
    _, err := q.db.ExecContext(ctx, createCategory, arg.ID, arg.Name, arg.Description)
    return err
}
```

##### `CreateCourse`

```go
type CreateCourseParams struct {
    ID          string
    Name        string
    Description sql.NullString
    CategoryID  string
    Price       float64
}

func (q *Queries) CreateCourse(ctx context.Context, arg CreateCourseParams) error {
    _, err := q.db.ExecContext(ctx, createCourse,
        arg.ID,
        arg.Name,
        arg.Description,
        arg.CategoryID,
        arg.Price,
    )
    return err
}
```

##### `ListCategories`

Retorna `[]Category`.

##### `GetCategory`

Retorna `Category`.

##### `ListCourses`

Gera uma struct específica para o resultado do `JOIN`:

```go
type ListCoursesRow struct {
    ID           string
    CategoryID   string
    Name         string
    Description  sql.NullString
    Price        float64
    CategoryName string
}
```

### Aprendizado principal

O `sqlc` não apenas gera funções. Ele também gera:

- structs de entrada,
- structs de saída,
- mapeamento de tipos SQL para tipos Go,
- segurança de compilação ao consumir as queries.

Isso reduz erros manuais e mantém a camada de dados consistente com o SQL real.

---

## 5) Execução simples das queries

O exemplo básico de uso está em:

- `cmd/runSQLC/main.go`

### O que esse arquivo faz

1. abre conexão com o banco MySQL,
2. cria uma instância de `Queries`,
3. executa `CreateCategory`,
4. lista categorias com `ListCategories`,
5. imprime o resultado.

### Trecho conceitual

```go
dbConn, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/courses")
queries := db.New(dbConn)
```

Depois disso:

```go
err = queries.CreateCategory(ctx, db.CreateCategoryParams{ ... })
categories, err := queries.ListCategories(ctx)
```

### O que esse exemplo ensina

Esse arquivo mostra o fluxo mais simples de uso do `sqlc`:

- você escreve SQL,
- o `sqlc` gera métodos,
- você apenas consome esses métodos em Go.

Sem montar SQL manualmente no `main.go`.

---

## 6) Estratégia de transactions

A parte mais importante do módulo além da geração de código é a estratégia de transaction implementada em:

- `cmd/runSQLCTX/main.go`

Esse arquivo demonstra como garantir **atomicidade** em uma operação composta.

---

## 7) Estrutura usada para trabalhar com transaction

### `CourseDB`

```go
type CourseDB struct {
    dbConn *sql.DB
    *db.Queries
}
```

Essa struct encapsula:

- a conexão principal com o banco,
- os métodos gerados pelo `sqlc`.

### Construtor

```go
func NewCourseDB(dbConn *sql.DB) *CourseDB {
    return &CourseDB{
        dbConn:  dbConn,
        Queries: db.New(dbConn),
    }
}
```

---

## 8) Método `callTx`

O coração da estratégia transacional é este método:

```go
func (c *CourseDB) callTx(ctx context.Context, fn func(*db.Queries) error) error {
    tx, err := c.dbConn.BeginTx(ctx, nil)
    if err != nil {
        return err
    }

    q := db.New(tx)
    err = fn(q)
    if err != nil {
        if errRb := tx.Rollback(); errRb != nil {
            return fmt.Errorf("error on rollback: %v, original error: %w", errRb, err)
        }
        return err
    }

    return tx.Commit()
}
```

### O que acontece aqui

1. inicia uma transaction com `BeginTx`,
2. cria uma nova instância de `Queries` usando a transaction,
3. executa a função callback recebendo `*db.Queries`,
4. faz `Rollback` se algo falhar,
5. faz `Commit` se tudo ocorrer bem.

### Por que isso funciona

Porque `db.New(...)` aceita qualquer implementação da interface `DBTX`.

Como `*sql.Tx` atende a interface `DBTX`, os métodos gerados pelo `sqlc` podem ser usados dentro da transação sem nenhuma duplicação de código.

Esse é um ponto central do aprendizado do módulo.

---

## 9) Operação atômica: criar categoria e curso juntos

O caso prático foi implementado em:

```go
func (c *CourseDB) CreateCourseAndCategory(ctx context.Context, argsCategory CategoryParams, argsCourse CourseParams) error {
    err := c.callTx(ctx, func(q *db.Queries) error {
        err := q.CreateCategory(ctx, db.CreateCategoryParams{
            ID:          argsCategory.ID,
            Name:        argsCategory.Name,
            Description: argsCategory.Description,
        })
        if err != nil {
            return err
        }

        err = q.CreateCourse(ctx, db.CreateCourseParams{
            ID:          argsCourse.ID,
            Name:        argsCourse.Name,
            Description: argsCourse.Description,
            CategoryID:  argsCategory.ID,
            Price:       argsCourse.Price,
        })
        if err != nil {
            return err
        }

        return nil
    })
    if err != nil {
        return err
    }
    return nil
}
```

### Qual é a ideia dessa operação

Esse método garante que duas etapas relacionadas sejam executadas como uma única unidade lógica:

1. cria a categoria,
2. cria o curso ligado à categoria.

### Benefício

Se a criação da categoria funcionar mas a criação do curso falhar, o `Rollback` desfaz tudo.

Ou seja:

- não fica categoria “solta” no banco,
- não fica estado inconsistente,
- a operação mantém integridade.

Esse é o valor real da transaction neste módulo.

---

## 10) Diferença entre uso normal e uso transacional

### Uso normal

No `runSQLC`:

- as operações são independentes,
- cada chamada é executada diretamente sobre `*sql.DB`.

### Uso transacional

No `runSQLCTX`:

- várias operações são agrupadas,
- todas executam sobre `*sql.Tx`,
- ou tudo persiste, ou nada persiste.

### Resumo

| Cenário | Conexão usada | Comportamento |
|---|---|---|
| Execução simples | `*sql.DB` | Cada query roda isoladamente |
| Transaction | `*sql.Tx` | Todas as queries participam do mesmo commit/rollback |

---

## 11) Conceitos aprendidos com este módulo

### 1. SQL continua sendo protagonista

Diferente de abordagens em que a camada de dados fica escondida atrás de abstrações muito genéricas, aqui o SQL é explícito e central.

Isso traz vantagens como:

- clareza do que realmente está sendo executado,
- controle fino sobre queries,
- facilidade para revisar performance e joins,
- proximidade maior com o banco real.

---

### 2. O `sqlc` gera código fortemente tipado

Com isso, você ganha:

- menos repetição manual,
- menos erro de scan,
- menos erro de ordem de parâmetros,
- melhor feedback em tempo de compilação.

---

### 3. O arquivo `.sql` vira uma API Go

Cada query nomeada vira:

- uma constante SQL,
- um método Go,
- e, quando necessário, structs auxiliares de entrada e saída.

Isso cria uma ponte elegante entre SQL puro e código Go.

---

### 4. Transactions podem reutilizar o mesmo código gerado

Essa é uma das lições mais importantes do módulo.

Graças à interface `DBTX`, não é necessário reimplementar métodos para transaction. O mesmo código gerado pelo `sqlc` serve para:

- `db.New(dbConn)`
- `db.New(tx)`

Isso reduz duplicidade e mantém a base mais limpa.

---

### 5. O padrão `callTx` é reutilizável

Esse padrão é muito útil porque centraliza:

- abertura de transaction,
- tratamento de erro,
- rollback,
- commit.

No futuro, a mesma ideia pode ser reaproveitada para outras operações compostas, mantendo o código mais organizado.

---

## 12) Fluxo prático para trabalhar neste módulo

### Gerar código com `sqlc`

```bash
cd ~/GoProject/GoExperts_Phase_5/5-SQLC
sqlc generate
```

### Compilar o projeto

```bash
go build ./...
```

### Rodar testes

```bash
go test ./...
```

### Visualizar o banco SQLite

```bash
sqlite3 data.db
```

Dentro do SQLite:

```sql
.tables
.schema categories
.schema courses
SELECT * FROM categories;
SELECT * FROM courses;
.quit
```

---

## 13) Fluxo mental para adicionar novas queries

Quando você quiser expandir esse módulo, o fluxo correto é:

1. atualizar o schema, se necessário,
2. escrever a nova query em `sql/queries/query.sql`,
3. rodar `sqlc generate`,
4. verificar o código gerado em `internal/db/`,
5. consumir o novo método no Go.

### Exemplo mental

Se você criar uma nova query como:

```sql
-- name: GetCourse :one
SELECT * FROM courses WHERE id = ?;
```

após rodar `sqlc generate`, o `sqlc` deverá gerar um método Go equivalente para buscar um curso por ID.

---

## 14) Benefícios práticos da abordagem adotada

### Clareza

As queries ficam centralizadas e legíveis.

### Segurança

O compilador ajuda a identificar incompatibilidades de tipos e assinaturas.

### Reuso

O mesmo conjunto de queries funciona com conexão normal e com transaction.

### Integridade

Operações compostas podem ser agrupadas com commit/rollback.

### Produtividade

Você escreve SQL uma vez e deixa o `sqlc` cuidar da camada repetitiva de acesso aos dados.

---

## 15) Limites atuais do módulo

Este módulo é focado em fundamentos. Pelo que foi implementado até aqui, ele ainda não introduz:

- camada de service/usecase separada,
- testes automatizados do fluxo SQLC,
- tratamento avançado de erros,
- retry policy,
- abstrações maiores de domínio,
- configuração por arquivo `.env` ou Viper neste módulo específico.

Isso é compatível com o objetivo didático da etapa atual: entender bem a geração de código com `sqlc` e a integração com transactions.

---

## 16) Resumo executivo do módulo

Este módulo ensina como:

- modelar tabelas com migrations,
- escrever queries SQL reais,
- gerar código Go tipado com `sqlc`,
- consumir esse código em aplicações Go,
- e aplicar transactions corretamente em operações compostas.

Em síntese:

- **migrations** definem o banco,
- **queries SQL** definem a persistência,
- **sqlc** gera a camada Go,
- **DBTX** torna possível reutilizar a mesma camada com `*sql.DB` e `*sql.Tx`,
- **callTx** garante atomicidade,
- **CreateCourseAndCategory** demonstra o uso prático de transaction.

---

## 17) Próximos passos naturais de estudo

Depois deste módulo, evoluções naturais seriam:

1. adicionar novas queries de leitura e atualização de cursos,
2. criar testes automatizados para os métodos gerados,
3. encapsular melhor regras de negócio fora do `main.go`,
4. explorar paginação e filtros com `sqlc`,
5. integrar esse padrão em uma API HTTP ou gRPC.

---

## 18) Conclusão

O principal valor deste módulo está em mostrar que é possível trabalhar com **SQL explícito**, **código Go tipado** e **transactions seguras** sem cair em excesso de boilerplate.

O `sqlc` entra exatamente nesse ponto: ele preserva o controle do SQL, mas elimina a parte repetitiva e frágil da implementação manual.

Já a estratégia de transaction mostra como manter integridade de dados de forma simples e reaproveitável.

Esse conjunto forma uma base muito sólida para aplicações Go que precisam de acesso a banco com clareza, segurança e previsibilidade.




## Comandos de execução

cd ~/GoProject/GoExperts_Phase_5/5-SQLC
docker compose up -d
docker compose exec -T mysql mysql -uroot -proot courses < sql/migrations/000001_init.down.sql
docker compose exec -T mysql mysql -uroot -proot courses < sql/migrations/000001_init.up.sql
go run ./cmd/runSQLC
go run ./cmd/runSQLCTX
docker compose exec mysql mysql -uroot -proot -D courses -e "SELECT * FROM categories;"
docker compose exec mysql mysql -uroot -proot -D courses -e "SELECT * FROM courses;"