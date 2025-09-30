# Benchmarks em Go

## 🔹 O que é um Benchmark no Go
Em Go, um **benchmark** é um tipo especial de teste que mede **desempenho** de funções.  
Ele não verifica se o resultado está correto (como os testes unitários), mas sim **quanto tempo e quanta memória uma função consome** ao ser executada repetidamente.

Benchmarks em Go usam o pacote `testing` e funções com a assinatura:

```go
func BenchmarkXxx(b *testing.B) {
    for b.Loop() {
        // código a ser medido
    }
}
```

---

## 🔹 Para que serve
- **Medir tempo médio por operação (ns/op):** saber se uma função é rápida ou lenta.  
- **Avaliar alocações de memória (B/op, allocs/op):** identificar desperdícios e gargalos.  
- **Comparar alternativas de implementação:** ajuda a escolher a versão mais eficiente de um algoritmo.  
- **Monitorar regressões:** você roda benchmarks em diferentes commits para verificar se houve piora de performance.

---

## 🔹 Relevância em projetos reais
Em projetos reais, benchmarks são cruciais quando:
- O código será executado **milhões de vezes** (ex.: funções de parsing, cálculo, criptografia, compressão).  
- Há requisitos de **baixa latência** (ex.: APIs em tempo real, trading, streaming).  
- Você precisa reduzir **custo em cloud** (menos CPU/memória → menos $$$).  
- Precisa garantir que otimizações realmente melhoraram algo, e não foram só “achismos”.

> Exemplo prático: no backend de pagamentos, benchmarks podem revelar se uma validação roda em microssegundos ou em milissegundos. Essa diferença, multiplicada por milhares de requisições, impacta diretamente na escalabilidade.

---

## 🔹 Possibilidades que ele oferece
1. **Medição de tempo**  
   - Resultado padrão: `ns/op` (nanosegundos por operação).

2. **Medição de memória**  
   - Usando `-benchmem`: exibe `B/op` e `allocs/op`.

3. **Sub-benchmarks**  
   - Você pode rodar vários cenários no mesmo benchmark:
     ```go
     func BenchmarkCalculateTax(b *testing.B) {
         b.Run("below-1000", func(b *testing.B) {
             for b.Loop() { _ = CalculateTax(500) }
         })
         b.Run("equal-1000", func(b *testing.B) {
             for b.Loop() { _ = CalculateTax(1000) }
         })
         b.Run("above-1000", func(b *testing.B) {
             for b.Loop() { _ = CalculateTax(1500) }
         })
     }
     ```

4. **Controle de tempo de execução**  
   - `go test -bench=. -benchtime=5s` → cada benchmark roda por 5 segundos.  
   - Útil para métricas mais estáveis.

5. **Controle de CPUs**  
   - `go test -bench=. -cpu=1,2,4` → roda benchmarks simulando diferentes quantidades de núcleos.

6. **Perfis e análise profunda**  
   - `go test -bench=. -cpuprofile=cpu.out` → gera perfil de CPU.  
   - `go tool pprof cpu.out` → análise visual para identificar gargalos.

---

## 🔹 Resumindo
- **Conceito:** benchmark = teste de performance em Go.  
- **Serve para:** medir tempo, memória e comparar implementações.  
- **Relevância:** essencial em partes críticas do sistema (latência, escala, custo).  
- **Possibilidades:** tempo, memória, cenários múltiplos, controle de execução, perfis de CPU/memória.

