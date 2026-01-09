package main

import (
	"fmt"
	"time"
)

/*
Resumo da aula — Go Routines e “multithreading” em Go

Este arquivo demonstra como o Go executa múltiplas tarefas de forma concorrente usando Go routines.

1) O que é uma Go routine
- Uma Go routine é uma função executada de forma concorrente ao fluxo principal, iniciada com a palavra-chave `go`.
- Ao chamar `go task("A")`, `go task("B")` e `go task("C")`, o programa dispara três execuções independentes da mesma
 função `task`, cada uma com um parâmetro diferente.

2) Concorrência vs paralelismo (o “multithreading” na prática)
- Concorrência: o programa faz progresso em várias tarefas ao mesmo tempo, alternando entre elas conforme o agendador
(scheduler) decide.
- Paralelismo: duas ou mais tarefas executam literalmente ao mesmo tempo, em múltiplos núcleos/threads do sistema.
- Em Go, você escreve concorrência com Go routines; o runtime do Go decide como agendar essas Go routines em threads
 do sistema operacional (M:N scheduling).
  Ou seja, você não cria/gerencia threads diretamente: o runtime gerencia as threads e “encaixa” as Go routines nelas.

3) O que este exemplo faz
- A função `task(name)` executa um loop de 0 a 9, imprime o índice e o nome da tarefa, e dorme 1 segundo a cada iteração.
- Em `main`, as três tarefas são iniciadas de forma concorrente. A saída no terminal aparece intercalada (“misturada”)
 porque as Go routines rodam ao mesmo tempo (concorrentes) e o scheduler alterna a execução entre elas.
- O `time.Sleep(10 * time.Second)` no `main` é apenas um “bloqueio” simples para impedir que o programa termine imediatamente.
  Sem isso, o `main` acabaria e o processo terminaria, interrompendo as Go routines antes delas completarem.

4) Observação importante (limitação do exemplo)
- Usar `Sleep` para “esperar” Go routines é didático, mas não é a forma recomendada em código real.
  Mais adiante, o padrão correto para coordenar finalização é usar `sync.WaitGroup` (ou canais, dependendo do caso).
*/

func task(name string) {
	for i := 0; i < 10; i++ {
		fmt.Printf("%d: Task %s is running\n", i, name)
		time.Sleep(500 * time.Millisecond)
	}
}

// Thread 1
func main() {
	// Thread 2
	go task("A")
	// Thread 3
	go task("B")
	// Thread 4
	go task("C")

	time.Sleep(10 * time.Second)
}

// OBS: SE RETIRAR O "go" antes das funções elas serão executadas de forma concorrente (1 por vez)
