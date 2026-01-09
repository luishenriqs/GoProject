package main

import (
	"fmt"
	"sync"
)

/*
Resumo — Channel + range + WaitGroup: aguardando o consumo de N mensagens

Este exemplo combina dois mecanismos:
1) Channel (`ch chan int`) para comunicação entre goroutines (publisher → reader).
2) WaitGroup (`sync.WaitGroup`) para garantir que a função `main` só finalize após
   todas as mensagens terem sido efetivamente processadas pelo `reader`.

Fluxo geral
- A `main` cria um channel de inteiros e um WaitGroup.
- `wg.Add(10)` define que existem 10 “unidades de trabalho” a serem concluídas.
- `publish(ch)` roda em uma goroutine e envia 10 valores para o channel, depois fecha o channel.
- `reader(ch, &wg)` roda em outra goroutine e consome o channel com `for x := range ch`,
  imprimindo cada valor e chamando `wg.Done()` uma vez por item processado.
- A `main` chama `wg.Wait()` e fica bloqueada até que o contador do WaitGroup chegue a zero.

Por que usar WaitGroup se o channel já é fechado?
- O `close(ch)` faz o `range` do `reader` terminar, mas a `main` não “espera automaticamente”
  o `reader` acabar apenas porque o channel foi fechado.
- O WaitGroup é o mecanismo explícito que garante: “só encerre o programa depois que
  o `reader` tiver processado todos os 10 valores”.

Pontos importantes deste padrão
- `range ch` no `reader`:
  - bloqueia esperando valores enquanto o channel estiver aberto,
  - encerra naturalmente quando o channel é fechado e não há mais valores.
- `wg.Done()` deve ser chamado exatamente 10 vezes (uma por mensagem).
  - Aqui isso é garantido porque o `publish` envia 10 itens e o `reader` chama `Done`
    uma vez por item recebido.
- O `wg.Add(10)` precisa refletir exatamente a quantidade de mensagens esperadas.
  Se enviar menos e `wg.Add` for maior, a `main` ficará bloqueada para sempre;
  se enviar mais e chamar `Done` mais do que o `Add`, ocorrerá panic por contador negativo.

Objetivo didático
- Mostrar a separação de responsabilidades:
  - Channel = transporte/sincronização de dados entre goroutines.
  - WaitGroup = coordenação de término (quando a aplicação pode encerrar).
*/

// Thread 1
func main() {
	ch := make(chan int)
	wg := sync.WaitGroup{}
	wg.Add(10)

	go publish(ch)
	go reader(ch, &wg)

	wg.Wait()
}

func reader(ch chan int, wg *sync.WaitGroup) {
	for x := range ch {
		fmt.Printf("Received %d\n", x)
		wg.Done()
	}
}

func publish(ch chan int) {
	for i := range 10 {
		ch <- i
	}
	close(ch)
}
