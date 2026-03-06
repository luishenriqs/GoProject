package main

import "github.com/luishenriqs/GoProject/GoExperts_Phase_4/5-Events/pkg/rabbitmq"

/*
Producer (cmd/producer/main.go) — Visão Geral

Este executável implementa um **producer** RabbitMQ em Go, responsável por:
- abrir um channel AMQP;
- publicar uma mensagem simples no RabbitMQ.

Conceitos principais demonstrados:
- **Producer**: processo que envia mensagens para o broker.
- **Exchange**: destino lógico onde a mensagem é publicada (aqui: `"amq.direct"`).
- **Channel AMQP**: canal lógico usado para executar `Publish`.
- **Fire-and-forget**: este exemplo apenas publica uma mensagem e encerra.

Fluxo de execução:
1) Abre um channel com `rabbitmq.OpenChannel()`.
2) Garante fechamento do channel ao final com `defer ch.Close()`.
3) Publica `"Hello World"` no exchange `"amq.direct"` via `rabbitmq.Publish(...)`.

Observações:
- Neste exemplo, não há verificação do erro retornado por `rabbitmq.Publish`.
  (A função retorna `error`, mas o caller não trata.)
- O roteamento depende do exchange e da routing key (no helper `Publish`, a routing key é `""`).
- Para observar a mensagem chegando, um consumer deve estar consumindo de uma fila que receba mensagens desse exchange.
*/
func main() {
	ch, err := rabbitmq.OpenChannel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	rabbitmq.Publish(ch, "Hello World", "amq.direct")
}
