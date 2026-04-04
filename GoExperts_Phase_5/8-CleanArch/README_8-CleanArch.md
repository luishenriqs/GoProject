# 8-CleanArch — README Técnico Consolidado

## Visão geral do módulo

Este módulo consolida vários fundamentos centrais de arquitetura de software em Go a partir de um caso de uso simples, porém completo: **criação de pedidos (orders)**. A aplicação foi estruturada para demonstrar, na prática, como separar responsabilidades entre domínio, casos de uso, infraestrutura e mecanismos de entrada/saída, mantendo o núcleo da regra de negócio desacoplado de detalhes externos.

O projeto expõe o mesmo caso de uso por **três interfaces diferentes**:

- **HTTP REST**
- **gRPC**
- **GraphQL**

Além disso, ele demonstra:

- **persistência em banco relacional**
- **publicação de evento de domínio** após a criação do pedido
- **integração com RabbitMQ**
- **injeção de dependências com Google Wire**
- **carregamento de configuração com Viper**
- **testes unitários e de infraestrutura**

Em termos didáticos, este é um módulo muito importante porque mostra como uma regra de negócio única pode ser reutilizada por diferentes canais de entrada sem duplicar a lógica central.

---

## Objetivo arquitetural

O problema que este módulo ajuda a resolver é o acoplamento excessivo entre:

- lógica de domínio
- acesso a banco
- transporte da requisição
- mensageria
- inicialização da aplicação

Em uma implementação ingênua, seria comum colocar tudo dentro do handler HTTP ou do serviço gRPC. Aqui, o projeto separa essas responsabilidades em camadas para que:

- o domínio permaneça simples e previsível
- o caso de uso concentre a orquestração da ação de negócio
- as interfaces de entrada apenas adaptem dados
- a infraestrutura cuide dos detalhes técnicos
- os componentes possam ser testados de forma isolada

---

## Estrutura do módulo

A árvore observada no módulo mostra a seguinte organização lógica:

```text
8-CleanArch/
├─ api/
│  └─ create_order.http
├─ cmd/
│  └─ ordersystem/
│     ├─ .env
│     ├─ main.go
│     ├─ wire.go
│     └─ wire_gen.go
├─ configs/
│  └─ config.go
├─ internal/
│  ├─ entity/
│  │  ├─ interface.go
│  │  ├─ order.go
│  │  └─ order_test.go
│  ├─ event/
│  │  ├─ order_created.go
│  │  └─ handler/
│  │     └─ order_created_handler.go
│  ├─ infra/
│  │  ├─ database/
│  │  │  ├─ order_repository.go
│  │  │  └─ order_repository_test.go
│  │  ├─ graph/
│  │  │  ├─ generated.go
│  │  │  ├─ resolver.go
│  │  │  ├─ schema.graphqls
│  │  │  ├─ schema.resolvers.go
│  │  │  └─ model/
│  │  │     └─ models_gen.go
│  │  ├─ grpc/
│  │  │  ├─ pb/
│  │  │  │  ├─ order.pb.go
│  │  │  │  └─ order_grpc.pb.go
│  │  │  ├─ protofiles/
│  │  │  │  └─ order.proto
│  │  │  └─ service/
│  │  │     └─ order_service.go
│  │  └─ web/
│  │     ├─ order_handler.go
│  │     └─ webserver/
│  │        ├─ starter.go
│  │        └─ webserver.go
│  └─ usecase/
│     └─ create_order.go
├─ pkg/
│  └─ events/
│     ├─ event_dispatcher.go
│     ├─ event_dispatcher_test.go
│     └─ interface.go
├─ docker-compose.yaml
├─ go.mod
├─ go.sum
├─ gqlgen.yml
└─ tools.go
```

---

## Leitura arquitetural por camadas

### 1. `internal/entity`

Esta é a camada mais próxima do domínio.

O arquivo `internal/entity/order.go` define a entidade `Order` com os campos:

- `ID`
- `Price`
- `Tax`
- `FinalPrice`

A entidade possui três responsabilidades centrais:

1. **representar o pedido**
2. **validar seu estado mínimo**
3. **calcular o preço final**

#### Regras de validação

O método `IsValid()` impõe:

- `ID` não pode ser vazio
- `Price` deve ser maior que zero
- `Tax` deve ser maior que zero

#### Regra de cálculo

O método `CalculateFinalPrice()` calcula:

```go
FinalPrice = Price + Tax
```

Depois do cálculo, a entidade ainda chama `IsValid()` novamente para garantir consistência.

#### Factory function

A função `NewOrder(id, price, tax)` cria uma nova instância e valida o objeto antes de retorná-lo.

### 2. `internal/usecase`

A camada de caso de uso contém a orquestração da ação de negócio.

No arquivo `internal/usecase/create_order.go`, o caso de uso `CreateOrderUseCase` depende de três contratos:

- `OrderRepositoryInterface`
- `EventInterface`
- `EventDispatcherInterface`

Ou seja, o caso de uso **não conhece detalhes de banco, RabbitMQ, HTTP, gRPC ou GraphQL**. Ele conhece apenas interfaces.

#### Fluxo do caso de uso

O método `Execute(input OrderInputDTO)` faz o seguinte:

1. monta uma entidade `Order`
2. calcula o preço final
3. persiste o pedido via repositório
4. monta um DTO de saída
5. injeta o DTO como payload no evento `OrderCreated`
6. dispara esse evento via dispatcher
7. retorna o DTO ao chamador

Esse desenho é o coração da Clean Architecture neste módulo.

### 3. `internal/infra/database`

Esta camada implementa a persistência.

`order_repository.go` implementa `OrderRepositoryInterface` usando `database/sql`.

#### Responsabilidades atuais

- inserir pedidos na tabela `orders`
- contar total de registros com `GetTotal()`

O método `Save(order *entity.Order)` executa um `INSERT` com:

- `id`
- `price`
- `tax`
- `final_price`

A camada de domínio não conhece SQL. Ela apenas depende da interface.

### 4. `pkg/events`

Esta camada contém uma infraestrutura reutilizável de eventos.

O `EventDispatcher` mantém um mapa de handlers por nome de evento e fornece operações para:

- `Register`
- `Dispatch`
- `Remove`
- `Has`
- `Clear`

#### Como o dispatcher funciona

Ao receber um evento em `Dispatch(event)`:

1. localiza todos os handlers registrados para aquele nome
2. cria um `sync.WaitGroup`
3. dispara cada handler em uma goroutine
4. espera todos terminarem

Isso demonstra um mecanismo simples e didático de processamento assíncrono com sincronização.

### 5. `internal/event`

Aqui ficam os tipos concretos de evento relacionados ao domínio.

`order_created.go` define o evento `OrderCreated`, com:

- nome do evento
- payload genérico
- data/hora de ocorrência

O evento é criado por `NewOrderCreated()` com nome fixo `OrderCreated`.

### 6. `internal/event/handler`

O handler `OrderCreatedHandler` trata o evento após sua emissão.

No estado atual, ele:

1. imprime o payload no terminal
2. serializa o payload em JSON
3. publica a mensagem no RabbitMQ

Isso evidencia bem a separação entre:

- **regra de negócio**: criar pedido
- **efeito colateral**: publicar um evento

### 7. `internal/infra/web`

Esta camada expõe o caso de uso via HTTP.

`order_handler.go` recebe a requisição, faz decode do JSON de entrada em `OrderInputDTO`, instancia o caso de uso e retorna o resultado em JSON.

`webserver/webserver.go` encapsula um servidor HTTP com `chi`, permitindo registrar handlers por rota e iniciar o servidor com middleware de log.

### 8. `internal/infra/grpc`

Esta camada expõe o mesmo caso de uso via gRPC.

- `protofiles/order.proto` define o contrato do serviço
- `pb/*.pb.go` são arquivos gerados automaticamente pelo `protoc`
- `service/order_service.go` adapta a requisição gRPC para o DTO do caso de uso e converte a resposta de volta para protobuf

A lógica de negócio continua centralizada no `CreateOrderUseCase`.

### 9. `internal/infra/graph`

Esta camada expõe o caso de uso via GraphQL.

- `schema.graphqls` define o schema
- `schema.resolvers.go` implementa o resolver da mutation `createOrder`
- `resolver.go` injeta dependências no resolver raiz
- `generated.go` e `model/models_gen.go` são arquivos gerados pelo `gqlgen`

Mais uma vez, o resolver apenas adapta entrada e saída. A regra permanece no caso de uso.

### 10. `cmd/ordersystem`

Esta é a composição da aplicação.

O `main.go` é o ponto de bootstrap e integra todos os componentes.

Ele é responsável por:

1. carregar configurações
2. abrir conexão com MySQL
3. abrir canal com RabbitMQ
4. criar o dispatcher de eventos
5. registrar o handler `OrderCreated`
6. montar o caso de uso
7. subir servidor web
8. subir servidor gRPC
9. subir servidor GraphQL

Em outras palavras, o `main` é onde a aplicação finalmente conecta o mundo externo ao núcleo interno.

---

## Como a Clean Architecture aparece neste projeto

A essência da Clean Architecture, neste módulo, está em três princípios muito claros.

### Dependência apontando para dentro

As camadas externas dependem das internas, nunca o contrário.

Exemplos:

- o handler web depende do use case
- o service gRPC depende do use case
- o resolver GraphQL depende do use case
- o use case depende apenas de interfaces e entidades

### Regra de negócio isolada

A regra “criar pedido” não está espalhada entre REST, GraphQL e gRPC.

Ela existe em um único lugar: `CreateOrderUseCase`.

### Infraestrutura substituível

O repositório é acessado por interface. Isso significa que, conceitualmente, seria possível trocar a implementação atual por outra sem alterar o caso de uso.

---

## Fluxo completo da criação de pedido

### Passo 1 — entrada

O pedido pode chegar por três portas:

- requisição HTTP
- chamada gRPC
- mutation GraphQL

### Passo 2 — adaptação

A camada de entrada converte os dados recebidos em `OrderInputDTO`.

### Passo 3 — caso de uso

O `CreateOrderUseCase`:

- cria a entidade
- calcula o preço final
- salva o pedido
- monta o DTO de saída
- dispara evento

### Passo 4 — side effect

O evento `OrderCreated` é enviado ao `EventDispatcher`, que chama o `OrderCreatedHandler`.

### Passo 5 — publicação

O handler publica o payload no RabbitMQ.

### Passo 6 — resposta

O DTO de saída volta ao consumidor no formato correspondente ao transporte utilizado.

---

## Google Wire neste módulo

Um dos pontos mais importantes deste módulo é o uso de **injeção de dependências**.

### O problema que o Wire resolve

Sem uma ferramenta de DI, a montagem dos objetos começa a encadear muitas dependências manualmente.

Exemplo típico:

- criar conexão com banco
- criar repositório com banco
- criar evento
- criar dispatcher
- criar use case com repositório + evento + dispatcher
- criar handler com dependências similares

Esse encadeamento cresce rapidamente e deixa o bootstrap mais verboso e sujeito a erro.

### O que é o Wire

O **Google Wire** é uma ferramenta de **injeção de dependências em tempo de compilação**.

Diferente de containers reflexivos em runtime, ele gera código Go estático. Isso traz vantagens didáticas e práticas:

- sem reflection em runtime
- wiring explícito
- erros detectáveis em build time
- código final simples de depurar

### Como ele aparece no projeto

No diretório `cmd/ordersystem/` existem dois arquivos principais:

- `wire.go`
- `wire_gen.go`

#### `wire.go`

É o arquivo declarativo, onde você informa ao Wire como montar os objetos.

Nele aparecem sets como:

- `setOrderRepositoryDependency`
- `setOrderCreatedEvent`

E também as funções injetoras:

- `NewCreateOrderUseCase(...)`
- `NewWebOrderHandler(...)`

#### `wire_gen.go`

É o arquivo gerado automaticamente pelo Wire.

Nele vemos o resultado concreto da composição, por exemplo:

- criação do repositório
- criação do evento `OrderCreated`
- montagem do `CreateOrderUseCase`
- montagem do `WebOrderHandler`

### Limitação observável no estado atual

O projeto demonstra Wire principalmente para:

- `CreateOrderUseCase`
- `WebOrderHandler`

Mas o `main.go` ainda mistura composição manual com composição gerada. Isso é coerente com o objetivo didático do módulo: mostrar o conceito funcionando sem necessariamente automatizar toda a aplicação.

---

## Configuração com Viper

O arquivo `configs/config.go` centraliza o carregamento de configuração.

As variáveis previstas são:

- `DB_DRIVER`
- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`
- `WEB_SERVER_PORT`
- `GRPC_SERVER_PORT`
- `GRAPHQL_SERVER_PORT`

O método `LoadConfig(path string)` usa `viper` para:

- localizar o arquivo `.env`
- carregar valores
- preencher a struct `conf`

No `.env` atual, os valores indicam:

- MySQL em `localhost:3306`
- RabbitMQ em `localhost:5672` (configurado separadamente no código)
- servidor HTTP em `:8000`
- gRPC em `50051`
- GraphQL em `8080`

---

## Contratos de entrada

### HTTP

O arquivo `api/create_order.http` mostra um exemplo de chamada REST:

```http
POST http://localhost:8000/order HTTP/1.1
Content-Type: application/json

{
  "id": "a",
  "price": 100.5,
  "tax": 0.5
}
```

### gRPC

O contrato está em `internal/infra/grpc/protofiles/order.proto`.

Ele define:

- `CreateOrderRequest`
- `CreateOrderResponse`
- serviço `OrderService`
- RPC `CreateOrder`

### GraphQL

O schema define uma mutation:

```graphql
type Mutation {
  createOrder(input: OrderInput): Order
}
```

Isso deixa explícito que o módulo implementa o mesmo comportamento por três interfaces distintas.

---

## Eventos e mensageria

A parte de eventos é um dos tópicos mais relevantes do módulo.

### Evento de domínio

Após a criação do pedido, o caso de uso dispara o evento `OrderCreated`.

### Dispatcher

O dispatcher faz o roteamento do evento para os handlers cadastrados.

### Handler concreto

O `OrderCreatedHandler` é responsável por publicar o payload no RabbitMQ.

### Benefício arquitetural

Com isso, o caso de uso não precisa conhecer RabbitMQ diretamente.

Ele apenas comunica que “um pedido foi criado”. O que será feito com essa informação passa a ser responsabilidade da infraestrutura de eventos.

---

## Testes existentes

O módulo contém testes em pontos importantes.

### 1. Testes da entidade

`internal/entity/order_test.go` cobre:

- ID vazio
- preço vazio
- taxa vazia
- criação válida
- factory `NewOrder`
- cálculo de preço final

Esses testes garantem a integridade mínima da entidade.

### 2. Testes do repositório

`internal/infra/database/order_repository_test.go` usa SQLite in-memory para validar persistência sem depender de infraestrutura externa permanente.

O teste atual verifica:

- criação de pedido válido
- cálculo de `FinalPrice`
- persistência correta em tabela `orders`
- leitura posterior do registro salvo

### 3. Testes do dispatcher

`pkg/events/event_dispatcher_test.go` cobre:

- registro de handlers
- prevenção de handler duplicado
- limpeza do mapa interno
- verificação de existência
- remoção de handlers
- dispatch para múltiplos handlers

Esse conjunto é importante porque valida o comportamento do mecanismo de eventos, que é uma peça central da arquitetura.

---

## Docker Compose

O arquivo `docker-compose.yaml` sobe dois serviços necessários ao módulo:

### MySQL

- imagem: `mysql:5.7`
- porta: `3306`
- banco inicial: `orders`
- usuário root com senha `root`

### RabbitMQ

- imagem: `rabbitmq:3-management`
- AMQP na porta `5672`
- painel de gerenciamento na porta `15672`
- usuário `guest`
- senha `guest`

Isso oferece a infraestrutura mínima para executar a aplicação localmente.

---

## Como executar o projeto

## Pré-requisitos

Você precisa ter instalado:

- Go
- Docker e Docker Compose
- `protoc` caso deseje regenerar arquivos protobuf
- plugin do gqlgen caso deseje regenerar artefatos GraphQL
- Google Wire caso deseje regenerar `wire_gen.go`

## 1. Subir infraestrutura

Na raiz do módulo:

```bash
docker compose up -d
```

## 2. Garantir a existência da tabela `orders`

O código atual usa a tabela `orders`, mas não há migration anexada neste contexto. Portanto, antes de executar a aplicação, é necessário garantir manualmente que a tabela exista no banco `orders`.

Uma estrutura compatível com o repositório atual é:

```sql
CREATE TABLE orders (
  id VARCHAR(255) NOT NULL PRIMARY KEY,
  price FLOAT NOT NULL,
  tax FLOAT NOT NULL,
  final_price FLOAT NOT NULL
);
```

## 3. Executar a aplicação

A partir de `cmd/ordersystem`:

```bash
go run .
```

Ou, a partir da raiz, ajustando o working directory conforme necessário para o `.env` ser encontrado.

## 4. Testar a interface HTTP

Use o arquivo `api/create_order.http` ou um cliente como curl/Postman:

```bash
curl -X POST http://localhost:8000/order \
  -H "Content-Type: application/json" \
  -d '{"id":"a","price":100.5,"tax":0.5}'
```

## 5. Testar GraphQL

Abra no navegador:

```text
http://localhost:8080/
```

A mutation esperada segue a ideia:

```graphql
mutation {
  createOrder(input: {id: "abc", Price: 100.5, Tax: 0.5}) {
    id
    Price
    Tax
    FinalPrice
  }
}
```

## 6. Testar gRPC

Você pode usar Evans, grpcurl ou outro cliente gRPC apontando para a porta `50051`.

---

## Como rodar os testes

Na raiz do módulo:

```bash
go test ./...
```

Para testar partes específicas:

```bash
go test ./internal/entity/...
go test ./internal/infra/database/...
go test ./pkg/events/...
```

---

## Como regenerar artefatos

### Regenerar Wire

No diretório `cmd/ordersystem`:

```bash
go generate
```

ou, caso necessário:

```bash
wire
```

### Regenerar protobuf

A partir do diretório adequado e com `protoc` instalado, regenere os arquivos a partir de `internal/infra/grpc/protofiles/order.proto`.

### Regenerar GraphQL

Com `gqlgen` disponível no ambiente:

```bash
go run github.com/99designs/gqlgen generate
```

---

## Pontos fortes do módulo

Este módulo ensina muito bem os seguintes conceitos:

1. **separação de responsabilidades**
2. **centralização da regra de negócio no use case**
3. **uso de interfaces para desacoplamento**
4. **publicação de eventos após ações de negócio**
5. **exposição do mesmo caso de uso por múltiplos transports**
6. **injeção de dependências com código gerado**
7. **testabilidade das peças centrais**

---

## Limitações e observações do estado atual

Como referência técnica, vale registrar alguns pontos observáveis no código atual.

### 1. O caso de uso monta a entidade diretamente

Em `CreateOrderUseCase.Execute`, a entidade é instanciada manualmente em vez de usar `NewOrder`. Isso funciona, mas reduz o reaproveitamento explícito da factory.

### 2. O retorno de erro de `CalculateFinalPrice()` não é tratado

O método é chamado, mas seu erro não é verificado no caso de uso.

### 3. `main.go` mistura composição manual e Wire

O projeto usa Wire, mas ainda não delega toda a composição ao código gerado.

### 4. Conexão e canal do RabbitMQ não são fechados explicitamente

No estado atual, o canal é retornado por `getRabbitMQChannel()` sem fechamento explícito do `conn`.

### 5. Não há migration anexada neste contexto

A tabela `orders` precisa existir previamente.

Esses pontos não invalidam a proposta do módulo. Pelo contrário: eles são comuns em exemplos didáticos cujo objetivo é enfatizar arquitetura e integração antes de refinar robustez operacional.

---

## Quando usar esta arquitetura

Esse tipo de organização é especialmente útil quando:

- a aplicação cresce em número de integrações
- o mesmo caso de uso precisa ser exposto por APIs diferentes
- há necessidade de isolar regra de negócio
- eventos de domínio precisam disparar ações adicionais
- testabilidade é importante

Ela tende a ser menos necessária em scripts pequenos ou aplicações extremamente simples, mas se torna muito valiosa conforme o sistema ganha mais entradas, saídas e integrações.

---

## Resumo final

Este módulo demonstra, de forma prática, como construir um sistema orientado a caso de uso em Go com forte separação entre domínio e infraestrutura.

O pedido é criado uma única vez no núcleo da aplicação, e esse mesmo comportamento é disponibilizado via REST, gRPC e GraphQL. A persistência fica isolada em repositório, o evento de domínio é disparado após a operação, e a publicação em RabbitMQ ocorre fora da regra de negócio. O bootstrap central usa Viper para configuração e Wire para auxiliar a montagem das dependências.

Como referência futura, este módulo deve ser lembrado como o ponto em que os conceitos de:

- entidade
- caso de uso
- repositório
- evento
- handler
- transporte
- injeção de dependência
- composição da aplicação

passam a trabalhar juntos dentro de uma estrutura coerente de **Clean Architecture**.
