package rabbitmq

import amqp "github.com/rabbitmq/amqp091-go"

/*
OpenChannel abre uma conexão AMQP com o RabbitMQ local e cria um channel para operações de publish/consume.

Fluxo:
  - Conecta no broker via amqp.Dial usando a URI (guest/guest em localhost:5672).
  - Cria um channel (conn.Channel), que é o "canal lógico" usado para interagir com exchanges/queues.

Retorno:
  - (*amqp.Channel, nil) em caso de sucesso.
  - Esta implementação não propaga erro: em falhas de conexão ou criação do channel, faz panic.
    (O `error` no retorno fica sempre `nil` no sucesso.)

Observações:
  - O channel deve ser fechado pelo chamador (ex.: `defer ch.Close()`).
  - A conexão (`conn`) não é fechada aqui; na prática, o ideal seria também gerenciar e fechar a conexão.
*/
func OpenChannel() (*amqp.Channel, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	return ch, nil
}

/*
Consume inicia o consumo de mensagens de uma fila (queue) e encaminha cada delivery para o canal `out`.

Parâmetros:
  - ch: channel AMQP já aberto (obtido via OpenChannel).
  - out: canal de saída para onde as mensagens consumidas serão repassadas.
  - queue: nome da fila a ser consumida (ex.: "orders").

Comportamento:
  - Registra um consumer via ch.Consume.
  - autoAck=false: exige ACK manual do consumidor (msg.Ack) para remover a mensagem da fila.
  - Encaminha cada mensagem recebida no channel interno `msgs` para o channel `out`.
  - Bloqueia enquanto `msgs` estiver aberto; termina quando o channel de deliveries for encerrado.

Retorno:
  - Retorna erro apenas se o registro do consumer (ch.Consume) falhar.
  - Se o loop terminar (channel fechado), retorna nil.

Observações:
  - A confirmação (Ack) é feita fora desta função, pelo código que lê de `out`.
*/
func Consume(ch *amqp.Channel, out chan<- amqp.Delivery, queue string) error {
	msgs, err := ch.Consume(
		queue,
		"go-consumer",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for msg := range msgs {
		out <- msg
	}

	return nil
}

/*
Publish publica uma mensagem (body) em um exchange (exName) usando o channel AMQP fornecido.

Parâmetros:
  - ch: channel AMQP já aberto.
  - body: conteúdo da mensagem a ser publicada.
  - exName: nome do exchange onde a mensagem será publicada (ex.: "amq.direct").

Comportamento:
  - Publica no exchange `exName` com routing key vazia ("").
  - Usa DeliveryMode padrão (não definido explicitamente) e ContentType "text/plain".
  - Marca flags:
    - mandatory=false: não exige que a mensagem seja roteada obrigatoriamente para uma fila.
    - immediate=false: não exige entrega imediata (flag legacy).

Retorno:
  - nil em caso de sucesso.
  - erro retornado por ch.Publish em caso de falha.

Observações:
  - Em exchanges do tipo direct, a routing key normalmente é relevante para roteamento.
    Aqui, a routing key é vazia por decisão do exemplo.
*/
func Publish(ch *amqp.Channel, body string, exName string) error {
	err := ch.Publish(
		exName,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	)
	if err != nil {
		return err
	}

	return nil
}
