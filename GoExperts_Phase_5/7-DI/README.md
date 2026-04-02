# 7-DI - Dependency Injection

## Objetivo do módulo

Este módulo demonstra o conceito de **Injeção de Dependências (Dependency Injection - DI)** em Go, partindo de um cenário simples e muito comum no desenvolvimento de aplicações.

A ideia principal estudada foi o seguinte problema:

- primeiro criamos um **banco de dados**
- depois criamos um **repositório**, que depende do banco
- depois criamos um **use case**, que depende do repositório
- e, conforme o sistema cresce, esse encadeamento manual de dependências vai ficando mais trabalhoso, mais verboso e mais propenso a erro

O professor usou esse cenário para introduzir o conceito de um **container** ou mecanismo de composição de dependências, cujo objetivo é facilitar a montagem da aplicação.

---

## O problema do encadeamento manual de dependências

Sem DI, é comum montar tudo manualmente no `main.go`.

Exemplo conceitual:

```go
db, _ := sql.Open("sqlite3", "./test.db")
repository := product.NewProductRepository(db)
usecase := product.NewProductUseCase(repository)
````

Em um exemplo pequeno isso parece simples. Porém, em sistemas reais, esse encadeamento cresce rapidamente:

* o banco pode depender de configuração
* o repositório pode depender do banco
* o use case pode depender de vários repositórios
* um handler pode depender de vários use cases
* middlewares podem depender de serviços adicionais

Com isso, o `main` começa a virar um ponto de montagem muito acoplado, difícil de manter e difícil de escalar.

Os principais problemas desse modelo manual são:

1. **aumento de acoplamento**
2. **dificuldade de manutenção**
3. **maior chance de erro na montagem**
4. **baixa legibilidade do grafo de dependências**
5. **mais dificuldade para evoluir a aplicação**

---

## O que é Injeção de Dependências

**Injeção de Dependências** é uma técnica em que um objeto não cria sozinho aquilo de que precisa. Em vez disso, suas dependências são fornecidas de fora.

No exemplo deste módulo:

* `ProductRepository` depende de `*sql.DB`
* `ProductUseCase` depende de `ProductRepositoryInterface`

Essas dependências não são criadas dentro das structs. Elas são recebidas por construtores:

```go
func NewProductRepository(db *sql.DB) *ProductRepository
func NewProductUseCase(repository ProductRepositoryInterface) *ProductUseCase
```

Isso traz vantagens importantes:

* reduz acoplamento
* melhora testabilidade
* deixa a estrutura mais explícita
* facilita troca de implementações
* melhora organização arquitetural

---

## O papel do container de dependências

Quando a aplicação cresce, alguém precisa ser responsável por montar o grafo de dependências.

Esse papel pode ser tratado como:

* um **container de DI**
* um **composition root**
* ou um **mecanismo de wiring**

A função dele é resolver perguntas como:

* qual implementação concreta atende esta interface?
* em que ordem os objetos devem ser criados?
* quais dependências cada construtor exige?
* como compor tudo sem fazer isso manualmente o tempo todo?

Neste módulo, esse papel foi resolvido com a biblioteca **Google Wire**.

---

## Bibliotecas de DI em Go

Existem diferentes abordagens para DI em Go.

### 1. Injeção manual

É a forma mais simples e idiomática em muitos projetos Go.

Vantagens:

* simples
* explícita
* sem dependências externas

Desvantagens:

* cresce mal em projetos maiores
* aumenta verbosidade
* dificulta manutenção do grafo de dependências

### 2. Containers/bibliotecas de DI

Algumas bibliotecas ajudam a resolver esse problema. Entre elas, é comum encontrar:

* **Google Wire**
* **Uber Dig**
* **Uber Fx** (mais voltado a aplicações com ciclo de vida)
* outras soluções menores ou abordagens internas do projeto

### 3. Diferença importante de abordagem

#### Wire

O **Wire** trabalha com **geração de código** em tempo de desenvolvimento.

Ou seja:

* você descreve como as dependências se relacionam
* o Wire gera código Go normal para montar essas dependências
* no runtime, não existe mágica nem container reflexivo ativo

#### Dig / Fx

Já bibliotecas como **Dig** e **Fx** costumam usar mais recursos dinâmicos em runtime, incluindo reflexão.

---

## Por que o professor escolheu o Google Wire

Neste módulo, o professor escolheu o **Google Wire** porque ele se encaixa muito bem no estilo da linguagem Go:

* mantém a construção das dependências explícita
* gera código Go comum
* evita reflexão em runtime
* facilita leitura do fluxo de criação
* ajuda a organizar o composition root da aplicação

O Wire não “esconde” a lógica da aplicação. Ele apenas automatiza a geração do código de montagem.

---

## Como o Google Wire funciona

O fluxo conceitual do Wire é este:

1. você cria funções construtoras
2. você informa ao Wire quais providers existem
3. você informa quais interfaces devem ser associadas a quais implementações concretas
4. você declara uma função injetora
5. o Wire gera um arquivo com o código final de montagem

No seu módulo, isso acontece por meio de:

* `wire.NewSet(...)`
* `wire.Bind(...)`
* `wire.Build(...)`
* geração do arquivo `wire_gen.go`

---

## Estrutura do módulo

```text
7-DI/
├─ product/
│  ├─ entity.go
│  ├─ repository.go
│  └─ usecase.go
├─ go.mod
├─ go.sum
├─ main.go
├─ wire.go
└─ wire_gen.go
```

---

## Análise dos arquivos

## `product/entity.go`

```go
package product

type Product struct {
	ID   int
	Name string
}
```

### Papel

Define a entidade `Product` usada no exemplo.

### Observação

É uma estrutura propositalmente simples, suficiente para demonstrar o fluxo de DI sem adicionar complexidade desnecessária.

---

## `product/repository.go`

```go
package product

import "database/sql"

type ProductRepositoryInterface interface {
	GetProduct(id int) (*Product, error)
}

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db}
}

func (r *ProductRepository) GetProduct(id int) (*Product, error) {
	return &Product{
		ID:   id,
		Name: "Product Name",
	}, nil
}
```

### Papel

Implementa a camada de repositório.

### Pontos importantes

#### Interface

```go
type ProductRepositoryInterface interface {
	GetProduct(id int) (*Product, error)
}
```

O use case não depende diretamente da struct concreta `ProductRepository`. Ele depende da interface.

Isso é importante porque:

* reduz acoplamento
* facilita testes
* permite trocar implementação sem alterar quem consome

#### Dependência do banco

```go
type ProductRepository struct {
	db *sql.DB
}
```

Aqui aparece a primeira dependência real do exemplo: o repositório precisa do banco.

#### Construtor

```go
func NewProductRepository(db *sql.DB) *ProductRepository
```

O repositório recebe o banco por injeção, em vez de criá-lo internamente.

#### Método `GetProduct`

No exemplo, ele retorna um produto fixo apenas para simplificar a demonstração:

```go
func (r *ProductRepository) GetProduct(id int) (*Product, error)
```

O foco do módulo não é persistência real, mas a montagem das dependências.

---

## `product/usecase.go`

```go
package product

type ProductUseCase struct {
	repository ProductRepositoryInterface
}

func NewProductUseCase(repository ProductRepositoryInterface) *ProductUseCase {
	return &ProductUseCase{repository}
}

// GetProduct returns a product by id
// This Product was not supposed to be returned. We should return a DTO instead.
// However, we will return it for now to keep the example simple.
func (u *ProductUseCase) GetProduct(id int) (*Product, error) {
	return u.repository.GetProduct(id)
}
```

### Papel

Implementa a camada de caso de uso.

### Pontos importantes

#### Dependência por interface

```go
type ProductUseCase struct {
	repository ProductRepositoryInterface
}
```

O use case depende da abstração, não da implementação concreta.

Esse é um ponto arquitetural importante do módulo.

#### Construtor

```go
func NewProductUseCase(repository ProductRepositoryInterface) *ProductUseCase
```

O use case recebe o repositório pronto.

#### Método `GetProduct`

```go
func (u *ProductUseCase) GetProduct(id int) (*Product, error)
```

Neste exemplo, ele apenas delega a chamada ao repositório.

### Observação importante do próprio código

O comentário explica que, em um cenário mais maduro, o ideal seria retornar um DTO em vez da entidade diretamente. Mas aqui o retorno da própria entidade foi mantido para simplificar o estudo.

---

## `main.go`

```go
package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./test.db")
	if err != nil {
		panic(err)
	}

	usecase := NewUseCase(db)

	product, err := usecase.GetProduct(1)
	if err != nil {
		panic(err)
	}

	fmt.Println(product.Name)
}
```

### Papel

É o ponto de entrada da aplicação.

### Fluxo executado

1. abre uma conexão SQLite
2. chama `NewUseCase(db)`
3. obtém um `ProductUseCase` pronto
4. executa `GetProduct(1)`
5. imprime o nome do produto

### Ponto central

O `main` não monta manualmente o repositório e o use case. Ele delega isso para a função `NewUseCase(db)`, que foi gerada pelo Wire.

Essa é justamente a simplificação buscada pelo módulo.

---

## `wire.go`

```go
//go:build wireinject
// +build wireinject

package main

import (
	"database/sql"

	"github.com/google/wire"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_5/7-DI/product"
)

var setRepositoryDependency = wire.NewSet(
	product.NewProductRepository,
	wire.Bind(new(product.ProductRepositoryInterface), new(*product.ProductRepository)),
)

func NewUseCase(db *sql.DB) *product.ProductUseCase {
	wire.Build(
		setRepositoryDependency,
		product.NewProductUseCase,
	)
	return &product.ProductUseCase{}
}
```

### Papel

Este é o arquivo onde você descreve ao Wire como o grafo de dependências deve ser montado.

### Conceitos importantes

#### Build tag `wireinject`

```go
//go:build wireinject
// +build wireinject
```

Esse arquivo existe apenas para o processo de geração do Wire. Ele não é o arquivo que entra no build normal da aplicação.

#### `wire.NewSet`

```go
var setRepositoryDependency = wire.NewSet(
	product.NewProductRepository,
	wire.Bind(new(product.ProductRepositoryInterface), new(*product.ProductRepository)),
)
```

Aqui você criou um conjunto de providers.

Esse set informa ao Wire:

* a função `product.NewProductRepository` pode ser usada para construir um repositório
* a implementação concreta `*product.ProductRepository` satisfaz a interface `product.ProductRepositoryInterface`

#### `wire.Bind`

```go
wire.Bind(new(product.ProductRepositoryInterface), new(*product.ProductRepository))
```

Esse ponto é fundamental.

Ele informa ao Wire que, quando alguém precisar de `ProductRepositoryInterface`, a implementação concreta a ser usada será `*ProductRepository`.

Sem isso, o Wire não saberia automaticamente como satisfazer a interface.

#### `wire.Build`

```go
func NewUseCase(db *sql.DB) *product.ProductUseCase {
	wire.Build(
		setRepositoryDependency,
		product.NewProductUseCase,
	)
	return &product.ProductUseCase{}
}
```

Essa função é uma espécie de “declaração de intenção”.

Você está dizendo ao Wire:

* use este `db *sql.DB` como dependência de entrada
* use os providers declarados
* monte um `*product.ProductUseCase`

O `return &product.ProductUseCase{}` é apenas um placeholder exigido para a função compilar antes da geração. O valor real será substituído pelo código gerado em `wire_gen.go`.

---

## `wire_gen.go`

```go
// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"database/sql"
	"github.com/google/wire"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_5/7-DI/product"
)

import (
	_ "github.com/mattn/go-sqlite3"
)

// Injectors from wire.go:

func NewUseCase(db *sql.DB) *product.ProductUseCase {
	productRepository := product.NewProductRepository(db)
	productUseCase := product.NewProductUseCase(productRepository)
	return productUseCase
}

// wire.go:

var setRepositoryDependency = wire.NewSet(product.NewProductRepository, wire.Bind(new(product.ProductRepositoryInterface), new(*product.ProductRepository)))
```

### Papel

Este arquivo é gerado automaticamente pelo Wire.

### O que ele mostra de forma prática

O Wire converteu a definição declarativa do `wire.go` em código Go normal e explícito:

```go
func NewUseCase(db *sql.DB) *product.ProductUseCase {
	productRepository := product.NewProductRepository(db)
	productUseCase := product.NewProductUseCase(productRepository)
	return productUseCase
}
```

Ou seja, no fim das contas, o Wire apenas gerou automaticamente o encadeamento que antes seria escrito manualmente.

### Vantagem disso

Você passa a manter:

* a intenção arquitetural no `wire.go`
* a montagem concreta no `wire_gen.go`

Sem precisar escrever manualmente esse boilerplate toda vez.

### Regra importante

**Nunca editar manualmente o `wire_gen.go`.**

Sempre que precisar alterar o grafo de dependências:

1. edite `wire.go`
2. execute `wire`
3. deixe o arquivo gerado ser recriado

---

## `go.mod`

```go
module github.com/luishenriqs/GoProject/GoExperts_Phase_5/7-DI

go 1.24.3

require (
	github.com/google/wire v0.7.0
	github.com/mattn/go-sqlite3 v1.14.38
)
```

### Dependências do módulo

#### `github.com/google/wire`

Biblioteca usada para geração do código de injeção de dependências.

#### `github.com/mattn/go-sqlite3`

Driver SQLite usado no exemplo para abrir a conexão com banco.

---

## Fluxo completo de dependências neste módulo

O grafo de dependências estudado aqui é:

```text
sql.DB -> ProductRepository -> ProductUseCase
```

### Em detalhes

#### 1. Banco

No `main.go`, a aplicação cria:

```go
db, err := sql.Open("sqlite3", "./test.db")
```

#### 2. Repositório

O Wire sabe que `ProductRepository` precisa de `*sql.DB`:

```go
product.NewProductRepository(db)
```

#### 3. Use case

O Wire também sabe que `ProductUseCase` precisa de `ProductRepositoryInterface`:

```go
product.NewProductUseCase(productRepository)
```

#### 4. Bind entre interface e implementação

O `wire.Bind(...)` resolve a associação:

```go
ProductRepositoryInterface -> *ProductRepository
```

#### 5. Resultado

A função `NewUseCase(db)` retorna o `ProductUseCase` já montado.

---

## O que este módulo ensina na prática

Este módulo consolida os seguintes aprendizados:

### 1. Dependências devem ser recebidas, não criadas internamente

Em vez de cada struct instanciar suas próprias dependências, elas recebem o que precisam via construtor.

### 2. Camadas devem depender de abstrações

O use case depende da interface do repositório, não da struct concreta.

### 3. O composition root deve centralizar a montagem

A composição das dependências fica concentrada em um ponto específico da aplicação.

### 4. Ferramentas como Wire reduzem boilerplate

Em vez de escrever manualmente toda a cadeia de construção, você descreve o grafo e deixa a ferramenta gerar o código.

### 5. Wire não é mágica em runtime

Ele gera código Go explícito. Isso o torna alinhado com a filosofia da linguagem.

---

## Vantagens do uso do Wire

No contexto estudado, as principais vantagens são:

* simplifica montagem de dependências
* reduz repetição de código
* deixa o grafo mais explícito
* ajuda na manutenção de aplicações maiores
* mantém type safety em tempo de compilação
* evita reflexão em runtime
* gera código legível

---

## Limitações e cuidados

Apesar das vantagens, é importante entender os limites:

### 1. Não substitui boa arquitetura

Wire não corrige acoplamento ruim por si só. Ele apenas organiza a montagem das dependências.

### 2. Exige disciplina nos construtores

Para funcionar bem, o projeto precisa ter construtores claros e dependências bem definidas.

### 3. Arquivo gerado não deve ser editado

Toda mudança deve ser feita no `wire.go`.

### 4. Pode ser excessivo em projetos muito pequenos

Em aplicações muito simples, DI manual pode continuar sendo suficiente.

---

## Build tags usadas no módulo

### `wire.go`

```go
//go:build wireinject
```

Indica que o arquivo serve para o processo de geração.

### `wire_gen.go`

```go
//go:build !wireinject
```

Indica que o arquivo gerado entra no build normal da aplicação.

Essas tags permitem separar claramente:

* o arquivo declarativo de geração
* do arquivo concreto que será compilado

---

## Como gerar novamente o arquivo do Wire

Dentro da pasta do módulo:

```bash
wire
```

Ou, alternativamente, usando o comando indicado pelo próprio arquivo:

```bash
go generate ./...
```

Se o binário `wire` ainda não estiver instalado:

```bash
go install github.com/google/wire/cmd/wire@latest
```

Se necessário, adicionar o binário ao `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

---

## Como rodar o projeto

### 1. Entrar na pasta do módulo

```bash
cd GoExperts_Phase_5/7-DI
```

### 2. Gerar o arquivo do Wire

```bash
wire
```

### 3. Compilar

```bash
go build ./...
```

### 4. Executar

```bash
go run .
```

### Saída esperada

```text
Product Name
```

---

## Como interpretar a saída do exemplo

Ao executar:

```bash
go run .
```

o fluxo é:

1. abre a conexão com SQLite
2. chama `NewUseCase(db)`
3. o `ProductUseCase` já vem montado pelo código gerado
4. `GetProduct(1)` devolve o produto mockado no repositório
5. o nome é impresso no terminal

A saída `Product Name` confirma que o encadeamento de dependências foi resolvido corretamente.

---

## Resumo técnico final

Este módulo estudou o problema clássico do encadeamento manual de dependências em aplicações Go.

O exemplo mostrou que, mesmo em um caso pequeno, já existe uma cadeia clara:

* banco
* repositório
* use case

A partir disso, foi introduzido o conceito de **Injeção de Dependências**, em que cada componente recebe suas dependências externamente por meio de construtores.

Para evitar que a montagem da aplicação fique totalmente manual e verbosa, o módulo utilizou a biblioteca **Google Wire**, que permite declarar o grafo de dependências e gerar automaticamente o código de composição.

No projeto desenvolvido:

* `ProductRepository` depende de `*sql.DB`
* `ProductUseCase` depende de `ProductRepositoryInterface`
* `wire.go` descreve como essas dependências devem ser montadas
* `wire_gen.go` contém o código gerado automaticamente
* `main.go` usa a função injetora pronta e mantém a aplicação mais limpa

Em resumo, este módulo serve como referência para entender:

* o problema da montagem manual de dependências
* o conceito de DI em Go
* o papel de interfaces e construtores
* a função de um container/mecanismo de composição
* e o funcionamento prático do **Google Wire** como ferramenta de geração de código para injeção de dependências

```
