package main

import (
	"fmt"
	"time"
)

/*
Resumo (Exemplo de Load Balancer com Channels + Goroutines)

Este arquivo demonstra um padrão simples de “load balancer” em Go usando:
- 1 channel compartilhado (`data`) como fila de trabalho (work queue)
- N goroutines (`worker`) como consumidores concorrentes dessa fila
- 1 produtor (a `main`) enviando tarefas para a fila

Como o “load balancer” acontece aqui
- O balanceamento NÃO é feito por um algoritmo explícito (round-robin, prioridade, etc.).
- Ele ocorre de forma implícita pelo runtime do Go:
  - Várias goroutines ficam bloqueadas esperando receber do mesmo channel.
  - Cada envio (`data <- i`) entrega o valor para “algum” worker disponível naquele momento.
  - Na prática, o trabalho é distribuído entre os workers conforme eles vão ficando livres,
    o que produz um efeito de balanceamento de carga.

Fluxo do programa
1) `main` cria o channel `data` (não-bufferizado, portanto cada envio sincroniza com um receive).
2) `main` inicia `QtdWorkes` goroutines chamando `worker(i, data)`.
3) Cada `worker` entra em um `for x := range data`, ficando continuamente disponível para receber tarefas.
4) `main` produz 1.000.000 de tarefas (0..999999) e envia no channel.
5) Quando um worker recebe uma tarefa:
  - imprime “Worker <id> received <tarefa>”
  - dorme 1 segundo simulando processamento (trabalho “lento”)
  - volta a esperar a próxima tarefa

Observações importantes deste exemplo
  - Sem `close(data)`, o `for range` dos workers nunca termina. Este arquivo é apenas demonstrativo.
  - Como `data` é não-bufferizado, o produtor pode ficar bloqueado quando não houver worker pronto para receber.
  - Criar 100.000 goroutines é propositalmente extremo para demonstrar concorrência,
    mas pode gerar alto consumo de memória/overhead em um cenário real.
  - O “balanceamento” aqui é “por disponibilidade”: workers mais rápidos/menos ocupados tendem a receber mais itens.

Parâmetros e responsabilidades
- worker(workerId int, data chan int)
  - workerId: identificador do worker apenas para logging.
  - data: channel de entrada das tarefas.

- main()
  - cria a fila (`data`)
  - inicia os workers (pool de consumidores)
  - publica as tarefas (produtor)

Este padrão é a base de muitos designs em Go: produtor -> channel -> pool de workers,
onde o channel funciona como fila e o runtime distribui o trabalho entre consumidores concorrentes.
*/
func worker(workerId int, data chan int) {
	for x := range data {
		fmt.Printf("Worker %d received %d\n", workerId, x)
		time.Sleep(time.Second)
	}
}

func main() {
	data := make(chan int)
	QtdWorkes := 100000

	for i := 0; i < QtdWorkes; i++ {
		go worker(i, data)
	}

	for i := 0; i < 1000000; i++ {
		data <- i
	}
}
