package main

import "fmt"

/*
Resumo — Consumindo um channel com `range` até o fechamento (publisher/reader)

Este exemplo demonstra o padrão clássico “publisher/consumer” em Go usando channels,
onde uma goroutine produtora envia uma sequência de valores e a função consumidora
lê esses valores até o channel ser fechado.

Fluxo de execução
1) A `main` (Thread 1) cria um channel não bufferizado: `ch := make(chan int)`.
2) A `publish(ch)` é iniciada em uma goroutine (Thread 2) para produzir valores.
3) A `reader(ch)` roda na goroutine principal e consome os valores do channel.

Como o `reader` funciona com `range`
- O loop `for x := range ch { ... }` faz leituras repetidas do channel.
- Ele só termina quando o channel é FECHADO (`close(ch)`) e todos os valores enviados
  forem consumidos.
- Isso evita ter que controlar manualmente quando parar de receber.

Como o `publish` produz e sinaliza término
- O `publish` envia 10 valores para o channel (1..10), imprimindo "Execução do publish"
  a cada iteração, e depois chama `close(ch)`.
- O `close(ch)` é o sinal definitivo de “não haverá mais mensagens”.
  Sem esse fechamento, o `reader` ficaria bloqueado para sempre esperando novos valores.

Sincronização implícita (channel sem buffer)
- Como `ch` não tem buffer, cada envio `ch <- valor` bloqueia até o `reader` receber.
- Isso coordena automaticamente o ritmo entre produtor e consumidor, fazendo com que
  as execuções de `publish` e `reader` se intercalem conforme o scheduler.

Objetivo didático
- Mostrar como `range` em channels é usado para consumir dados “até acabar”,
  e como `close(ch)` é essencial para permitir que o consumidor finalize sem deadlock.
*/

// Thread 1
func main() {
	ch := make(chan int)
	go publish(ch)
	reader(ch)
}

func reader(ch chan int) {
	for x := range ch {
		fmt.Printf("Execução do reader %d\n", x)
	}
}

func publish(ch chan int) {
	for i := range 10 {
		ch <- i + 1
		fmt.Println("Execução do publish")
	}
	close(ch)
}
