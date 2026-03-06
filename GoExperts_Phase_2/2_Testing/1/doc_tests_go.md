# Testes Automatizados em Go

## 🔹 O que são testes automatizados
Em Go, **testes automatizados** são funções que validam o comportamento de outras funções.  
Eles verificam se a saída obtida corresponde ao resultado esperado, garantindo que o código funciona corretamente e continua funcionando após mudanças.

---

## 🔹 Estrutura de um teste em Go
Testes utilizam o pacote `testing` e seguem a convenção:

```go
func TestXxx(t *testing.T) {
    // lógica de teste
}
```

- O nome da função deve começar com `Test` (ex.: `TestCalculateTax`).  
- O parâmetro é um ponteiro para `testing.T`, usado para reportar falhas.  
- Arquivos de teste devem terminar com `_test.go` (ex.: `tax_test.go`).

---

## 🔹 Exemplo prático
Suponha a função `CalculateTax` definida em `tax.go`:

```go
package tax

func CalculateTax(amount float64) float64 {
    if amount >= 1000 {
        return 10.0
    }
    return 5.0
}
```

O teste correspondente ficaria em `tax_test.go`:

```go
package tax

import "testing"

func TestCalculateTax(t *testing.T) {
    tests := []struct {
        name   string
        amount float64
        want   float64
    }{
        {"amount below 1000", 500.0, 5.0},
        {"amount equal 1000", 1000.0, 10.0},
        {"amount above 1000", 1500.0, 10.0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := CalculateTax(tt.amount)
            if got != tt.want {
                t.Errorf("CalculateTax(%v) = %v; want %v", tt.amount, got, tt.want)
            }
        })
    }
}
```

---

## 🔹 Como rodar os testes
Use o comando:

```bash
go test ./...
```

Ou para saída detalhada:

```bash
go test -v ./...
```

---

## 🔹 Boas práticas em testes Go
1. **Nomes descritivos** nos casos de teste (`amount below 1000`).  
2. **Table-driven tests** (estruturas de dados que listam cenários, como no exemplo).  
3. **Uso de `t.Run`** para isolar subcasos.  
4. **Cobrir casos de borda** (ex.: valores negativos, zero, limites exatos).  
5. **Testes pequenos, rápidos e determinísticos**.  

---

## 🔹 Relevância em projetos reais
- Garantem **qualidade** e **confiabilidade** do código.  
- Facilitam **refatorações** sem medo de quebrar funcionalidades.  
- Servem como **documentação executável** do comportamento esperado.  
- Integram-se facilmente em **pipelines CI/CD**, evitando que código com bugs seja liberado.

---

## 🔹 Resumindo
- **Conceito:** testes verificam se a função entrega o resultado esperado.  
- **Estrutura:** funções `TestXxx` no arquivo `_test.go`.  
- **Execução:** `go test` roda todos os testes do pacote.  
- **Relevância:** garantem que o código se mantém correto ao longo da evolução do projeto.
