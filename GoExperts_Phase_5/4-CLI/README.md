# Go Experts -- CLI com Cobra

## Visão Geral

Este módulo implementa uma **Command Line Interface (CLI)** em Go
utilizando a biblioteca **Cobra**.\
O objetivo é demonstrar como construir ferramentas de linha de comando
profissionais aplicando boas práticas de arquitetura.

Durante o desenvolvimento deste módulo foram explorados diversos
conceitos importantes:

-   Construção de aplicações CLI em Go
-   Estruturação de comandos utilizando **Cobra**
-   Uso de **flags locais e flags globais**
-   Uso de **shorthand flags**
-   **Hooks de execução** (`PreRun`, `PostRun`)
-   Diferença entre `Run` e `RunE`
-   **Injeção de dependência**
-   **Inversão de controle**
-   Persistência com **SQLite**
-   Organização de código em **camadas**

------------------------------------------------------------------------

# Estrutura do Projeto

    4-CLI
    │
    ├─ cmd
    │  ├─ root.go
    │  ├─ category.go
    │  ├─ create.go
    │  ├─ list.go
    │  └─ teste.go
    │
    ├─ internal
    │  └─ database
    │      ├─ category.go
    │      └─ course.go
    │
    ├─ data.db
    ├─ go.mod
    ├─ go.sum
    └─ main.go

------------------------------------------------------------------------

# Entry Point da Aplicação

Arquivo:

    main.go

Responsável por iniciar a CLI.

``` go
func main() {
    cmd.Execute()
}
```

A função `Execute()` dispara toda a árvore de comandos da aplicação.

------------------------------------------------------------------------

# Comando Raiz

Arquivo:

    cmd/root.go

Define o comando principal da CLI.

``` go
var rootCmd = &cobra.Command{
    Use:   "4-CLI",
    Short: "A brief description of your application",
}
```

Executar:

    go run main.go

Resultado:

    Usage:
      4-CLI [command]

    Available Commands:
      category
      teste
      help

------------------------------------------------------------------------

# Hierarquia de Comandos

    4-CLI
    │
    ├─ category
    │   ├─ create
    │   └─ list
    │
    └─ teste

Exemplos:

    go run main.go category create
    go run main.go category list
    go run main.go teste

------------------------------------------------------------------------

# Flags no Cobra

## Flags Locais

Flags locais pertencem apenas ao comando em que foram definidas.

Exemplo:

``` go
createCmd.Flags().StringP("name", "n", "", "Name of the category")
createCmd.Flags().StringP("description", "d", "", "Description of the category")
```

Uso:

    go run main.go category create --name "Livros" --description "Categoria de livros"

ou

    go run main.go category create -n "Livros" -d "Categoria de livros"

------------------------------------------------------------------------

## Flags Globais (Persistent Flags)

Flags globais são herdadas pelos subcomandos.

Exemplo:

``` go
categoryCmd.PersistentFlags().StringP("name", "n", "DefaultValue", "Name of the category")
categoryCmd.PersistentFlags().BoolP("exists", "e", false, "Check if category exists")
categoryCmd.PersistentFlags().Int16P("id", "i", 0, "ID of the category")
```

Essas flags ficam disponíveis em todos os subcomandos do comando
`category`.

------------------------------------------------------------------------

# Diferença entre String e StringP

Cobra permite duas formas de definir flags.

### String

``` go
Flags().String("name", "", "Name")
```

Uso:

    --name

### StringP

Permite usar **shorthand flag**.

``` go
Flags().StringP("name", "n", "", "Name")
```

Uso:

    --name
    -n

------------------------------------------------------------------------

# Hooks de Execução

Cobra permite executar código antes e depois do comando.

Exemplo:

``` go
PreRun: func(cmd *cobra.Command, args []string) {
    fmt.Println("Chamado antes do Run")
},

PostRun: func(cmd *cobra.Command, args []string) {
    fmt.Println("Chamado depois do Run")
},
```

Fluxo de execução:

    PreRun → Run → PostRun

Também existem variações que retornam erro:

    PreRunE
    RunE
    PostRunE

------------------------------------------------------------------------

# Run vs RunE

### Run

Não retorna erro.

``` go
Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("Executando comando")
}
```

### RunE

Permite retornar erro.

``` go
RunE: func(cmd *cobra.Command, args []string) error {
    return nil
}
```

Se ocorrer erro, o Cobra trata automaticamente.

------------------------------------------------------------------------

# Injeção de Dependência

No comando `create`, o repositório de categoria é **injetado** no
comando.

``` go
func newCreateCmd(categoryDb database.Category) *cobra.Command {
    return &cobra.Command{
        RunE: runCreate(categoryDb),
    }
}
```

O comando não cria sua dependência, ele apenas **recebe o que precisa
para funcionar**.

Isso reduz acoplamento e facilita testes.

------------------------------------------------------------------------

# Inversão de Controle

A criação da dependência ocorre fora do comando.

``` go
createCmd := newCreateCmd(GetCategoryDB(GetDb()))
```

O comando não controla mais a criação do banco ou do repositório.

Esse padrão caracteriza **Inversão de Controle (IoC)**.

------------------------------------------------------------------------

# Camada de Persistência

Localizada em:

    internal/database

Exemplo da entidade Category.

``` go
type Category struct {
    db          *sql.DB
    ID          string
    Name        string
    Description string
}
```

------------------------------------------------------------------------

# Banco de Dados

SQLite é utilizado como banco local.

Arquivo:

    data.db

Conexão:

``` go
func GetDb() *sql.DB {
    db, err := sql.Open("sqlite3", "./data.db")
}
```

Tabela utilizada:

``` sql
CREATE TABLE categories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL
);
```

------------------------------------------------------------------------

# Criando Categorias

Exemplo:

    go run main.go category create -n "Livros" -d "Categoria de livros"

------------------------------------------------------------------------

# Listando Categorias

Comando disponível:

    go run main.go category list

------------------------------------------------------------------------

# Comando de Teste

Arquivo:

    cmd/teste.go

Permite testar flags.

    go run main.go teste -c ping

Resultado:

    ping

------------------------------------------------------------------------

# Validação

Comandos úteis durante desenvolvimento:

    go fmt ./...
    go mod tidy
    go run main.go

------------------------------------------------------------------------

# Conceitos Aprendidos

-   CLI em Go
-   Cobra CLI framework
-   Estrutura de comandos
-   Flags locais
-   Flags globais
-   Shorthand flags
-   Hooks (`PreRun`, `PostRun`)
-   `Run` vs `RunE`
-   Injeção de dependência
-   Inversão de controle
-   Persistência SQLite
-   Arquitetura em camadas

------------------------------------------------------------------------

# Conclusão

Este módulo demonstra como construir uma CLI robusta em Go utilizando
Cobra, aplicando conceitos fundamentais de arquitetura de software como
**injeção de dependência**, **inversão de controle** e **organização
modular do código**.
