# gRPC com Go — Guia Completo do Projeto

## Visão geral

Este projeto implementa um servidor gRPC em Go com persistência em SQLite e contrato definido em Protocol Buffers. O módulo cobre:

- definição do contrato `.proto`
- geração de código Go com `protoc`
- implementação de RPCs unary e streaming
- subida do servidor gRPC
- testes com Evans e grpcurl
- organização em camadas simples: `proto`, `internal/pb`, `internal/service`, `internal/database` e `cmd/grpcServer`

O servidor sobe na porta `50051`, registra o `CategoryService` e habilita reflection.

## Tecnologias utilizadas

### Go
Usado para implementar o servidor, serviços e persistência.

**Vantagens:** simplicidade, performance, concorrência nativa e ótima integração com gRPC.

### gRPC
Framework de RPC baseado em HTTP/2 e Protocol Buffers.

**Vantagens:** contratos fortemente tipados, alta performance, suporte a streaming e geração automática de cliente/servidor.

### Protocol Buffers
Formato do contrato da API.

**Vantagens:** payloads compactos, contrato explícito e geração de código em múltiplas linguagens.

### SQLite
Banco local usado no projeto.

**Vantagens:** simplicidade, sem servidor separado e ótimo para estudo/desenvolvimento local.

### Evans e grpcurl
Ferramentas de teste e inspeção para serviços gRPC.

## Estrutura do projeto

```text
2-gRPC/
├─ cmd/
│  └─ grpcServer/
│     └─ main.go
├─ internal/
│  ├─ database/
│  │  ├─ category.go
│  │  └─ course.go
│  ├─ pb/
│  │  ├─ course_category.pb.go
│  │  └─ course_category_grpc.pb.go
│  └─ service/
│     └─ category.go
├─ proto/
│  └─ course_category.proto
└─ db.sqlite
```

### Responsabilidade das camadas

- `proto/`: contrato gRPC.
- `internal/pb/`: código gerado pelo `protoc`.
- `internal/database/`: acesso a dados com `database/sql` e SQLite.
- `internal/service/`: implementação dos RPCs.
- `cmd/grpcServer/`: bootstrap do servidor.

## Contrato gRPC atual

O arquivo `proto/course_category.proto` define o package `pb`, o `go_package` como `internal/pb` e o `CategoryService` com os seguintes RPCs:

### Unary
- `CreateCategory(CreateCategoryRequest) returns (CategoryResponse)`
- `ListCategories(Blank) returns (CategoryList)`
- `GetCategory(CategoryGetRequest) returns (Category)`

### Client streaming
- `CreateCategoryStream(stream CreateCategoryRequest) returns (CategoryList)`

### Bidirectional streaming
- `CreateCategoryStreamBidirectional(stream CreateCategoryRequest) returns (stream Category)`

## Implementação atual

O servidor abre `./db.sqlite`, cria o repositório de categorias, instancia `CategoryService`, registra o serviço e ativa reflection.

Trecho essencial:

```go
categoryDb := database.NewCategory(db)
categoryService := service.NewCategoryService(*categoryDb)

grpcServer := grpc.NewServer()
pb.RegisterCategoryServiceServer(grpcServer, categoryService)
reflection.Register(grpcServer)
```

## Requisitos para rodar

Instale:

- Go
- `protoc`
- `protoc-gen-go`
- `protoc-gen-go-grpc`
- Evans (opcional)
- grpcurl (opcional)

Validar:

```bash
go version
protoc --version
```

Instalar plugins Go do protobuf:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Se necessário, adicione o binário Go ao `PATH`:

```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bashrc
source ~/.bashrc
```

## Como gerar os arquivos protobuf

Sempre que o `.proto` mudar:

```bash
protoc \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/course_category.proto
```

Esse comando gera/atualiza:

- `internal/pb/course_category.pb.go`
- `internal/pb/course_category_grpc.pb.go`

## Como rodar o projeto

No diretório raiz do módulo:

```bash
go run ./cmd/grpcServer
```

O servidor ficará escutando em `127.0.0.1:50051`.

### Se a porta já estiver ocupada

```bash
ss -ltnp | grep ':50051'
kill -9 <PID>
```

Depois rode novamente:

```bash
go run ./cmd/grpcServer
```

## Como testar com Evans

### Instalação

```bash
go install github.com/ktr0731/evans@latest
```

### Conexão

Com o servidor rodando em outro terminal:

```bash
evans -r -p 50051
```

### Fluxo validado neste projeto

Na versão usada aqui, o caminho que funcionou foi:

```text
show package
package pb
service CategoryService
show rpc
```

Nessa versão, o comando correto foi `show rpc`.

### Testar `CreateCategory`

```text
call CreateCategory
```

Payload de exemplo:

```json
{
  "name": "Categoria Teste",
  "description": "Descrição Teste"
}
```

### Testar `ListCategories`

```text
call ListCategories
```

Como o request é `Blank`, normalmente basta enviar `{}`.

### Testar `GetCategory`

```text
call GetCategory
```

Payload:

```json
{
  "id": "ID_DA_CATEGORIA"
}
```

### Testar `CreateCategoryStream`

Comportamento esperado:

- enviar vários `CreateCategoryRequest`
- encerrar o envio
- receber um `CategoryList` com as categorias criadas

### Testar `CreateCategoryStreamBidirectional`

Comportamento esperado:

- enviar um `CreateCategoryRequest`
- receber um `Category`
- repetir o fluxo até encerrar o stream

## Como testar com grpcurl

Listar serviços:

```bash
grpcurl -plaintext 127.0.0.1:50051 list
```

Listar métodos do serviço:

```bash
grpcurl -plaintext 127.0.0.1:50051 list pb.CategoryService
```

Chamar `ListCategories`:

```bash
grpcurl -plaintext -d '{}' 127.0.0.1:50051 pb.CategoryService/ListCategories
```

Chamar `GetCategory`:

```bash
grpcurl -plaintext -d '{"id":"ID_DA_CATEGORIA"}' 127.0.0.1:50051 pb.CategoryService/GetCategory
```

Chamar `CreateCategory`:

```bash
grpcurl -plaintext -d '{"name":"Nova","description":"Teste"}' 127.0.0.1:50051 pb.CategoryService/CreateCategory
```

## Serviços implementados hoje

### `CreateCategory`
Cria uma categoria no banco e devolve `CategoryResponse`.

### `CreateCategoryStream`
Recebe vários requests via stream do cliente e ao final retorna um `CategoryList`.

### `CreateCategoryStreamBidirectional`
Recebe vários requests e devolve uma `Category` por vez.

### `ListCategories`
Consulta o repositório e devolve `CategoryList`.

### `GetCategory`
Busca uma categoria específica por ID.

## Camada de banco de dados

### `Category`
Métodos existentes:

- `Create(name, description)`
- `FindAll()`
- `FindByCourseID(courseID)`
- `Find(id)`

### `Course`
Métodos existentes:

- `Create(name, description, categoryID)`
- `FindAll()`
- `FindByCategoryID(categoryID)`

Isso já deixa o projeto preparado para evoluir um `CourseService`.

## Como criar um novo serviço gRPC

### 1. Definir o contrato no `.proto`

Exemplo:

```proto
message CreateCourseRequest {
  string name = 1;
  string description = 2;
  string category_id = 3;
}

message Course {
  string id = 1;
  string name = 2;
  string description = 3;
  string category_id = 4;
}

service CourseService {
  rpc CreateCourse(CreateCourseRequest) returns (Course) {}
}
```

### 2. Regenerar os arquivos Go

```bash
protoc \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/course_category.proto
```

### 3. Implementar a camada de banco
Criar ou reaproveitar um repositório em `internal/database/`.

### 4. Implementar o service
Criar `internal/service/course.go` implementando a interface gerada.

### 5. Registrar no servidor
No `main.go`:

```go
courseDb := database.NewCourse(db)
courseService := service.NewCourseService(*courseDb)
pb.RegisterCourseServiceServer(grpcServer, courseService)
```

### 6. Reiniciar o servidor
Sempre reinicie após alterar contrato ou registro.

## Quando usar gRPC

Mais indicado para:

- microsserviços
- APIs internas de alta performance
- comunicação fortemente tipada
- cenários com streaming
- integração backend-to-backend

Menos indicado para:

- APIs públicas simples consumidas diretamente por browser sem gateway
- sistemas muito simples em que REST/JSON seja suficiente

## Cuidados importantes

### Arquivos gerados precisam estar sincronizados
Use sempre o `.proto` como fonte de verdade e regenere os arquivos sempre que houver dúvida.

### Reflection precisa estar habilitado
Sem `reflection.Register(grpcServer)`, Evans e grpcurl não descobrem o serviço corretamente.

### Processo antigo preso na porta
Se houver um processo antigo na `50051`, o cliente pode se conectar ao binário errado.

### Evans varia por versão
Neste projeto, o fluxo correto foi:

```text
show package
package pb
service CategoryService
show rpc
```

## Fluxo recomendado de desenvolvimento

1. editar o `.proto`
2. regerar os arquivos protobuf
3. implementar ou ajustar o service
4. ajustar repositório, se necessário
5. registrar o service no `main.go`
6. rodar o servidor
7. testar com Evans ou grpcurl

## Comandos úteis

### Gerar protobuf

```bash
protoc \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/course_category.proto
```

### Subir o servidor

```bash
go run ./cmd/grpcServer
```

### Verificar porta em uso

```bash
ss -ltnp | grep ':50051'
```

### Conectar com Evans

```bash
evans -r -p 50051
```

### Listar serviços com grpcurl

```bash
grpcurl -plaintext 127.0.0.1:50051 list
```

## Próximas evoluções naturais

- criar `CourseService`
- listar cursos por categoria
- criar cliente Go
- tratar `sql.ErrNoRows` explicitamente no `GetCategory`
- adicionar testes automatizados
- evoluir o contrato protobuf

## Resumo final

Este módulo já cobre os fundamentos mais importantes de gRPC com Go:

- contrato com Protocol Buffers
- geração automática de código Go
- RPC unary
- client streaming
- bidirectional streaming
- SQLite como persistência local
- servidor em `:50051`
- testes com Evans e grpcurl
- reflection para inspeção de serviços

É uma base sólida para evoluir para múltiplos serviços, interceptors, observabilidade, testes automatizados e comunicação entre microsserviços.
