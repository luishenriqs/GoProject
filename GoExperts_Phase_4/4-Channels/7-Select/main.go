package main

import (
	"fmt"
	"sync/atomic"
	"time"
)

type Message struct {
	id  int64
	Msg string
}

/*
Este exemplo demonstra como usar **channels + goroutines + select** para consumir mensagens de
múltiplas fontes concorrentes, com um **timeout** para evitar bloqueio indefinido.

Visão geral do fluxo:
- A `main` cria dois channels (`c1` e `c2`) do tipo `Message`. Cada channel representa uma fonte:
  - `c1` simula uma fila/event stream (ex.: RabbitMQ)
  - `c2` simula outra fonte (ex.: Kafka)
  - Um contador global `i` (int64) é compartilhado entre as duas goroutines produtoras para gerar IDs
    sequenciais de mensagem.
  - Cada produtor roda em loop infinito:
    1) Incrementa `i` com `atomic.AddInt64(&i, 1)`
    2) Monta um `Message{ id, "..." }`
    3) Envia esse `Message` para o seu channel (`c1` ou `c2`)

Por que `atomic` é usado:
  - As duas goroutines acessam e incrementam o mesmo `i` simultaneamente.
  - Sem sincronização, isso geraria **race condition** (o detector `-race` apontaria acesso concorrente).
  - `atomic.AddInt64` garante que o incremento seja **atômico**, evitando leituras/escritas concorrentes
    inconsistentes no contador.

Como o consumo funciona (select):
- A `main` fica em loop infinito executando um `select` com três possibilidades:
  - `case msg := <-c1`: recebe quando houver mensagem disponível no channel do RabbitMQ.
  - `case msg := <-c2`: recebe quando houver mensagem disponível no channel do Kafka.
  - `case <-time.After(3s)`: dispara quando nenhum dos dois channels entrega mensagem por 3 segundos,
    imprimindo "timeout".
  - O `select` escolhe **um caso pronto por vez**. Se mais de um caso estiver pronto ao mesmo tempo,
    o Go escolhe um deles de forma pseudo-aleatória (não existe garantia de ordem “justa”).

Observações importantes:
  - Como os producers rodam em loop infinito e os channels são **não-bufferizados** (sem capacidade),
    cada envio (`c1 <- msg`, `c2 <- msg`) só completa quando o `select` do consumidor realizar o receive.
    Na prática, o consumidor impõe o ritmo.
  - Com os producers sempre enviando, o `timeout` tende a não disparar; ele serve para ilustrar a técnica
    de “não travar para sempre” quando uma fonte fica silenciosa.
  - Não há `close(c1)`/`close(c2)` porque o exemplo não possui condição de encerramento: é um “listener”
    infinito de múltiplas fontes.
*/
func main() {
	c1 := make(chan Message)
	c2 := make(chan Message)
	var i int64 = 0

	// RabbitMQ
	go func() {
		for {
			atomic.AddInt64(&i, 1)
			msg := Message{i, "Hello from RabbitMQ"}
			c1 <- msg
		}
	}()

	// Kafka
	go func() {
		for {
			atomic.AddInt64(&i, 1)
			msg := Message{i, "Hello from Kafka"}
			c2 <- msg
		}
	}()

	for {
		select {
		case msg := <-c1: // rabbitmq
			fmt.Printf("Received from RabbitMQ: ID: %d - %s\n", msg.id, msg.Msg)

		case msg := <-c2: // kafka
			fmt.Printf("Received from Kafka: ID: %d - %s\n", msg.id, msg.Msg)

		case <-time.After(time.Second * 3):
			println("timeout")
		}
	}
}
