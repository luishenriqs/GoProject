# 5-Events — Eventos em Go + RabbitMQ (Produtor/Consumidor)

Este módulo consolida dois tópicos estudados:

1) **Eventos em Go (Event Dispatcher / Observer Pattern)** — implementação didática de um dispatcher em memória com registro e disparo de handlers.  
2) **RabbitMQ (mensageria)** — uso prático de filas com um **producer** e um **consumer** em Go, conectando no RabbitMQ local via Docker.

---

## Estrutura do diretório

```
5-Events/
  cmd/
    consumer/main.go
    producer/main.go
  pkg/
    events/
      interfaces.go
      event_dispatcher.go
      event_dispatcher_test.go
    rabbitmq/
      rabbitmq.go
  docker-compose.yaml
  go.mod
  go.sum
```

- **`pkg/events`**: foco em arquitetura e conceitos de eventos/handlers/dispatcher.
- **`pkg/rabbitmq` + `cmd/*`**: foco em mensageria com RabbitMQ (abrir canal, publicar e consumir mensagens).

---

## Parte 1 — Eventos em Go (Event Dispatcher)

### Ideia central
Um **evento** representa “algo aconteceu” no sistema. Em vez de acoplar ações diretamente no fluxo principal, você:

- dispara um evento (ex.: `"UserCreated"`, `"OrderPlaced"`);
- e o dispatcher executa automaticamente todas as operações registradas (handlers) para aquele evento.

Isso reduz acoplamento e facilita adicionar/remover reações sem alterar o ponto onde o evento é disparado.

### Contratos (interfaces) — `pkg/events/interfaces.go`
O módulo define três contratos principais:

- **`EventInterface`**
  - `GetName() string`: nome do evento (chave de roteamento no dispatcher)
  - `GetDateTime() time.Time`: timestamp do evento
  - `GetPayload() any`: dados do evento

- **`EventHandlerInterface`**
  - `Handle(event EventInterface, wg *sync.WaitGroup)`
  - O handler é a “reação” ao evento. Neste módulo, o handler recebe um `WaitGroup` para sinalizar conclusão via `wg.Done()`.

- **`EventDispatcherInterface`**
  - `Register`, `Dispatch`, `Remove`, `Has`, `Clear`
  - Controla registro e execução de handlers por nome de evento.

### Implementação do Dispatcher — `pkg/events/event_dispatcher.go`
O dispatcher é um registry em memória:

- mantém `handlers map[string][]EventHandlerInterface`
- permite registrar múltiplos handlers por evento
- impede registro duplicado do mesmo handler (`ErrHandlerAlreadyRegistered`)
- ao dar `Dispatch(event)`:
  - busca handlers por `event.GetName()`
  - chama `Handle(event, wg)` para cada handler
  - faz `wg.Wait()` para aguardar todos finalizarem

**Conceito importante:** neste desenho, a finalização do `Dispatch` só ocorre quando *todas* as reações (handlers) terminam.

### Testes — `pkg/events/event_dispatcher_test.go`
O módulo valida o comportamento do dispatcher com `testify/suite` e `testify/mock`:

- **Register**: registra múltiplos handlers e garante ordem
- **Register (duplicado)**: retorna `ErrHandlerAlreadyRegistered`
- **Dispatch**: garante que todos os handlers registrados são chamados exatamente 1x
- **Remove**: remove handlers mantendo os demais
- **Has**: verifica presença de handler no evento
- **Clear**: zera o registry

Rodar testes:
```bash
go test ./...
```

---

## Parte 2 — RabbitMQ (mensageria) com Go

### Visão teórica (mapeando com a UI do RabbitMQ)
Componentes principais:

- **Producer**: publica mensagens.
- **Exchange**: recebe mensagens do producer e roteia para filas.
- **Queue**: armazena mensagens até serem consumidas.
- **Consumer**: lê mensagens de uma fila e processa.
- **Channel**: “canal” lógico sobre uma conexão AMQP, usado para Publish/Consume.
- **ACK**: confirmação manual de processamento para remover a mensagem da fila.

Pelos prints do painel de gerenciamento:
- Há exchanges padrão como **`amq.direct`**.
- As filas podem ser criadas e visualizadas no dashboard (ex.: `orders`).
- O consumo com `autoAck=false` faz a mensagem ficar como **unacked** até o consumer chamar `Ack()`.

---

## Implementação no projeto

### Docker (RabbitMQ local)
O arquivo `docker-compose.yaml` sobe o RabbitMQ com management UI e portas padrão.

Subir o RabbitMQ:
```bash
docker compose up -d
```

Acessar o painel:
- Management UI: `http://localhost:15672`
- Usuário/senha: `guest / guest`

> Observação prática: no seu fluxo, a fila `orders` aparece criada no painel.

---

## Biblioteca RabbitMQ — `pkg/rabbitmq/rabbitmq.go`

### `OpenChannel()`
- Abre conexão via `amqp.Dial("amqp://guest:guest@localhost:5672/")`
- Abre um channel com `conn.Channel()`
- Retorna `*amqp.Channel`

### `Publish(ch, body, exName)`
- Publica no exchange informado (no producer, foi usado `amq.direct`)
- Usa routing key `""` (string vazia)
- Envia payload como `text/plain`

### `Consume(ch, out, queue)`
- Consome mensagens da fila informada (`queue`)
- Usa `autoAck=false` (ACK manual)
- Encaminha cada `amqp.Delivery` para o canal `out`

---

## Producer e Consumer (executáveis) — `cmd/*`

### Consumer — `cmd/consumer/main.go`
Fluxo:
1. abre channel (`rabbitmq.OpenChannel()`)
2. cria `msgs := make(chan amqp.Delivery)`
3. inicia consumo em goroutine: `go rabbitmq.Consume(ch, msgs, "orders")`
4. lê mensagens do canal `msgs`
5. imprime e confirma processamento:
   - `fmt.Println(string(msg.Body))`
   - `msg.Ack(false)`

Rodar consumer:
```bash
go run ./cmd/consumer
```

### Producer — `cmd/producer/main.go`
Fluxo:
1. abre channel
2. publica mensagem no exchange `amq.direct`:
   - `rabbitmq.Publish(ch, "Hello World", "amq.direct")`

Rodar producer:
```bash
go run ./cmd/producer
```

---

## Roteiro recomendado de execução (passo a passo)

1) Suba o RabbitMQ:
```bash
docker compose up -d
```

2) Confira no painel (`http://localhost:15672`) se a fila `orders` existe (ou crie).

3) Em um terminal, rode o consumer:
```bash
go run ./cmd/consumer
```

4) Em outro terminal, rode o producer:
```bash
go run ./cmd/producer
```

5) Verifique:
- o consumer imprimindo `Hello World`
- no painel, métricas de mensagens (Ready/Unacked/Total) e taxa de entrega/ack

---

## O que este módulo consolidou

- **Eventos em Go** como mecanismo de desacoplamento (dispatcher + handlers).
- Uso de `sync.WaitGroup` para coordenar execução e finalização de múltiplos handlers no `Dispatch`.
- **RabbitMQ** como middleware de mensageria:
  - conexão/channel AMQP
  - publish em exchange
  - consume de queue
  - ack manual garantindo confiabilidade do processamento
- Integração prática (producer/consumer) e observação no painel do RabbitMQ (queues/exchanges/métricas).
