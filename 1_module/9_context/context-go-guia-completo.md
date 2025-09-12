
# 📚 Introdução ao `context` em Go

O pacote `context` da linguagem Go é essencial para controle de **cancelamento**, **timeout**, **deadlines** e **transporte de dados entre goroutines**. Ele é amplamente utilizado em servidores, chamadas HTTP e sistemas concorrentes.

---

## 🔁 Ciclo de Vida de um Context

1. Um `context.Context` é criado (geralmente com `Background()` ou `TODO()`).
2. Deriva-se um novo contexto com `WithCancel`, `WithTimeout` ou `WithDeadline`.
3. Recebe-se também uma `CancelFunc`, que **deve ser chamada** ao final para liberar recursos.
4. `ctx.Done()` é usado para detectar o cancelamento.

---

## 🧩 Interface `context.Context`

```go
type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key any) any
}
```

---

## 📌 Criação de Contextos

### `context.Background()`
- Contexto raiz. Nunca é cancelado. Usado em `main`, testes e inicializações.

### `context.TODO()`
- Placeholder. Use quando ainda não se sabe qual contexto usar.

---

## 🔄 Contextos Derivados

### `context.WithCancel(parent)`
- Cancela-se manualmente com `cancel()`.

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
```

---

### `context.WithTimeout(parent, duration)`
- Cancela automaticamente após o tempo definido.

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
```

---

### `context.WithDeadline(parent, time)`
- Cancela automaticamente no momento especificado.

---

### `context.WithValue(parent, key, value)`
- Armazena um valor ligado à requisição. **Use tipo de chave customizado.**

```go
type ctxKey string
ctx = context.WithValue(ctx, ctxKey("userID"), 123)
```

---

## 🔎 Cancelamento e Finalização

### `ctx.Done()`
- Canal fechado quando o contexto é cancelado.

### `ctx.Err()`
- Retorna `context.Canceled` ou `context.DeadlineExceeded`.

---

## 🚫 Boas Práticas

- **Não armazene context em struct.**
- **Passe como primeiro argumento de funções.**
- **Não use `nil`. Use `context.TODO()` se necessário.**
- **Use `WithValue` apenas para dados de escopo de requisição.**

---

## 🔥 Extensões com Motivo

- `WithCancelCause`, `WithTimeoutCause`, `WithDeadlineCause`
- Permitem adicionar um **motivo de cancelamento (`error`)** acessível via `context.Cause(ctx)`.

---

## 🧵 Concorrência

- Contexts são **seguros para múltiplas goroutines**.

---

## 📘 Exemplo Prático

```go
func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    select {
    case <-time.After(5 * time.Second):
        fmt.Println("trabalho concluído")
    case <-ctx.Done():
        fmt.Println("cancelado:", ctx.Err()) // context deadline exceeded
    }
}
```

---

## 📎 Recursos Úteis

- https://go.dev/blog/context
- https://go.dev/blog/context-and-structs
