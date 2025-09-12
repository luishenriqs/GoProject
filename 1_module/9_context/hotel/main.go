package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
    ctx := context.Background()
    ctx, cancel := context.WithTimeout(ctx, time.Second*10)
    defer cancel()

    bookHotel(ctx)
}

func bookHotel(ctx context.Context) {
    select {
        case <-ctx.Done(): // ctx.Done é contexto finalizado
            fmt.Println("Hotel booking cancelled. Timeout reached.")
            return

        case <-time.After(4 * time.Second):
            fmt.Println("Processing.")

        case <-time.After(8 * time.Second):
            fmt.Println("Hotel booked.")
    }
}


// Esse select é executado uma única vez. Ele aguarda a primeira das três opções 
// que estiver pronta, e descarta as outras. Neste caso:

// Após 4 segundos, o time.After(4 * time.Second) é o primeiro a ficar pronto.

// Ele imprime Processing..

// O select termina — não há novo bloco select para continuar.