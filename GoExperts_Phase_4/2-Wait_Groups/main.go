package main

import (
	"fmt"
	"sync"
	"time"
)

func task(name string) {
	for i := 0; i < 10; i++ {
		fmt.Printf("%d: Task %s is running\n", i, name)
		time.Sleep(1 * time.Second)
	}
}

/*
Resumo da aula — Goroutines + WaitGroup (concorrência em Go)

Este exemplo demonstra como o Go executa funções em paralelo usando goroutines e
como sincronizar a finalização dessas execuções com sync.WaitGroup.

Conceitos principais:
  - Goroutine: é uma função executada de forma concorrente ao fluxo principal.
    Ao usar `go ...`, o Go agenda essa execução em paralelo (concorrência), sem
    exigir que você crie e gerencie threads manualmente como em muitas linguagens.
  - Concorrência vs. paralelismo: concorrência é a capacidade de organizar e
    alternar múltiplas tarefas ao longo do tempo; paralelismo é executar realmente
    em múltiplos núcleos ao mesmo tempo. O scheduler do Go decide quando e onde
    cada goroutine roda, por isso a ordem dos `Println/Printf` pode variar.
  - sync.WaitGroup: mecanismo para a goroutine principal (main) aguardar a
    conclusão de um conjunto de goroutines, evitando que o programa termine cedo
    ou fique preso esperando por uma contagem incorreta.

Como este código funciona:
 1. `task(name)` imprime 10 linhas (0..9), pausando 1 segundo entre cada impressão.
 2. Em `main()`, criamos um `WaitGroup` e chamamos `wg.Add(3)` para declarar que
    existem 3 unidades de trabalho (as 3 goroutines).
 3. Cada goroutine chama `task("A")`, `task("B")` e `task("C")`, e usa `defer wg.Done()`
    para garantir que, ao finalizar, decrementa a contagem do WaitGroup mesmo se
    ocorrer um retorno antecipado.
 4. `wg.Wait()` bloqueia a main até que as 3 goroutines chamem `Done()`, fazendo a
    contagem voltar a zero. Só então o programa encerra.

Observação:
  - A ordem das mensagens no console é não determinística, pois depende do scheduler.
  - Se `Add(n)` não bater exatamente com o total de `Done()`, o programa pode travar
    (deadlock) ou encerrar cedo.
*/
func main() {
	var wg sync.WaitGroup

	wg.Add(3)

	go func() {
		defer wg.Done()
		task("A")
	}()

	go func() {
		defer wg.Done()
		task("B")
	}()

	go func() {
		defer wg.Done()
		task("C")
	}()

	wg.Wait()
}
