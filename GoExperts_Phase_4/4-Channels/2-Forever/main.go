package main

// import (
// 	"fmt"
// 	"time"
// )

/*
Resumo — “Forever channel” como bloqueio da goroutine principal (e variação para múltiplas goroutines)

Este arquivo demonstra um padrão simples de sincronização usando channels para impedir que a função `main`
termine antes de uma (ou mais) goroutine(s) concluírem seu trabalho.

1) Caso base: 1 goroutine + 1 sinal (forever)
- `forever := make(chan bool)` cria um channel sem buffer (capacidade 0). Ele começa “vazio”.
- A goroutine (Thread 2) executa um loop e, ao final, envia um sinal: `forever <- true`.
- A `main` (Thread 1) fica bloqueada em `<-forever` até receber esse sinal.
- Na prática, isso funciona como um “aguarde até terminar”: a `main` só continua (e encerra) quando a goroutine
  sinaliza que acabou.

Ponto importante:
- Em channel sem buffer, o envio (`forever <- true`) e o recebimento (`<-forever`) precisam “se encontrar”:
  o envio só completa quando existe alguém recebendo. Isso cria sincronização automática entre as goroutines.

2) Equivalente para 3 goroutines: 1 channel “done” + 3 recebimentos
- O trecho comentado mostra como estender a ideia para múltiplas goroutines.
- Em vez de apenas um sinal, a `main` precisa receber um sinal para cada goroutine concluída.
- Cada goroutine executa sua tarefa e envia `done <- true` ao terminar.
- A `main` faz três recebimentos (`<-done` três vezes), garantindo que todas finalizaram antes do programa encerrar.

Relação com WaitGroup (intuition)
- Esse padrão com channel (“receber N sinais”) é equivalente à ideia de um WaitGroup onde:
  - cada `done <- true` representa uma conclusão de trabalho (como um `Done()`),
  - e cada `<-done` representa “aguardar” uma conclusão (contabilizando N vezes).

Objetivo didático
- Mostrar que channels podem ser usados não só para transportar dados, mas também como mecanismo de
  sincronização/coordenação (sinalização de término), controlando explicitamente quando a `main` pode finalizar.
*/

// Thread 1
func main() {
	forever := make(chan bool) // Vazio

	// Thread 2
	go func() {
		for i := 0; i < 10; i++ {
			println(i)
		}
		forever <- true
	}()

	<-forever

}

// #######################################################
// Como ficaria o equivalente do forever para 3 goroutines

// func task(name string) {
// 	for i := 0; i < 10; i++ {
// 		fmt.Printf("%d: Task %s is running\n", i, name)
// 		time.Sleep(1 * time.Second)
// 	}
// }

// func main() {
// 	done := make(chan bool)

// 	go func() {
// 		task("A")
// 		done <- true
// 	}()

// 	go func() {
// 		task("B")
// 		done <- true
// 	}()

// 	go func() {
// 		task("C")
// 		done <- true
// 	}()

// 	// "WaitGroup-like": espera 3 sinais de término
// 	<-done
// 	<-done
// 	<-done
// }
