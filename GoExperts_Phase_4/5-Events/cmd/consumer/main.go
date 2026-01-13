package main

import (
	"fmt"

	"github.com/luishenriqs/GoProject/GoExperts_Phase_4/5-Events/pkg/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

/*
Consumer (cmd/consumer/main.go) — Visão Geral

Este executável implementa um **consumer** RabbitMQ em Go, responsável por:
- abrir um channel AMQP;
- iniciar o consumo de mensagens de uma fila específica (`orders`);
- processar cada mensagem recebida (aqui: imprimir no stdout);
- confirmar o processamento via **ACK manual**.

Conceitos principais demonstrados:
  - **Channel AMQP**: canal lógico usado para consumir/publicar mensagens.
  - **Goroutine**: o consumo é iniciado em paralelo (não bloqueia o fluxo principal).
  - **Canal Go (chan amqp.Delivery)**: ponte entre a rotina de consumo e o loop de processamento.
  - **ACK manual (autoAck=false)**: a mensagem só é removida da fila após `msg.Ack(false)`.
    Isso garante confiabilidade: se o processo cair antes do ACK, a mensagem pode ser reentregue.

Fluxo de execução:
1) Abre um channel com `rabbitmq.OpenChannel()`.
2) Cria `msgs` para receber as mensagens (`amqp.Delivery`).
3) Inicia a rotina de consumo: `go rabbitmq.Consume(ch, msgs, "orders")`.
4) Itera em `for msg := range msgs`, processando cada delivery.
5) Faz ACK manual após processar: `msg.Ack(false)`.

Observações:
- `defer ch.Close()` garante o fechamento do channel ao finalizar o programa.
- O loop `for msg := range msgs` fica bloqueado aguardando mensagens.
- A fila `"orders"` deve existir no RabbitMQ para que o consumer consiga consumir.
*/
func main() {
	ch, err := rabbitmq.OpenChannel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	msgs := make(chan amqp.Delivery)

	go rabbitmq.Consume(ch, msgs, "orders")

	for msg := range msgs {
		fmt.Println(string(msg.Body))
		msg.Ack(false)
	}
}
