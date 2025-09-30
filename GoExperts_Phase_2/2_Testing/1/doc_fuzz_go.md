# Fuzz Testing em Go

## 🔹 O que é Fuzz Testing
No Go, **fuzz testing** (ou fuzzing) é uma técnica de testes automatizados que gera entradas aleatórias ou mutadas para uma função com o objetivo de encontrar falhas inesperadas, panics ou comportamentos incorretos.  
Enquanto testes unitários verificam cenários conhecidos e benchmarks medem desempenho, o fuzz explora entradas imprevisíveis para aumentar a cobertura e detectar problemas ocultos.

---

## 🔹 Como funciona no Go
- Introduzido oficialmente no **Go 1.18** e evoluído em versões posteriores.  
- Usa o mesmo pacote `testing`, com funções que começam com `FuzzXxx`.  
- O desenvolvedor fornece alguns **seeds** (`f.Add(...)`) como pontos de partida.  
- O motor do fuzz muta e combina esses valores, executando milhões de casos em busca de falhas ou novos caminhos de código.

Exemplo simples:

```go
func FuzzCalculateTax(f *testing.F) {
    f.Add(float64(0))
    f.Add(float64(1000))
    f.Add(float64(1500))

    f.Fuzz(func(t *testing.T, amount float64) {
        got := CalculateTax(amount)
        if got != 0.0 && got != 5.0 && got != 10.0 {
            t.Fatalf("valor inesperado: %v", got)
        }
    })
}
```

Execução:
```bash
go test -fuzz=Fuzz -fuzztime=10s
```

---

## 🔹 Importância do Fuzz Testing
1. **Descobrir falhas ocultas**: encontra casos extremos ou entradas inesperadas que o programador não antecipou.  
2. **Aumentar a cobertura de testes**: gera milhares/milhões de entradas automaticamente.  
3. **Reduzir riscos de panics em produção**: principalmente útil em funções que lidam com entrada de usuários, parsing de dados ou formatos complexos.  
4. **Complementar testes unitários**: cobre cenários que seriam inviáveis de escrever manualmente.  

---

## 🔹 Benefícios em projetos reais
- **Confiabilidade**: garante que funções se comportem bem mesmo com dados corrompidos, maliciosos ou extremos.  
- **Segurança**: ajuda a encontrar pontos vulneráveis que poderiam ser explorados (ex.: buffer overflow, panics).  
- **Economia de tempo**: evita a escrita manual de centenas de casos de teste.  
- **Evolução segura**: ao alterar funções críticas, o fuzz rapidamente valida que novos inputs não quebram o comportamento esperado.  

---

## 🔹 Cenários ideais para usar Fuzz
- Funções de parsing (JSON, XML, CSV, protocolos binários).  
- Manipulação de strings e regex.  
- Conversões numéricas e cálculos complexos.  
- Validações de entrada de usuário.  
- Qualquer função sujeita a valores externos ou imprevisíveis.

---

## 🔹 Resumindo
- **Conceito:** fuzz = gerar entradas aleatórias/mutadas para achar falhas.  
- **Importância:** amplia cobertura e encontra erros que unit tests não cobrem.  
- **Benefícios:** mais segurança, robustez e confiança no código.  
- **Aplicação prática:** ideal para funções que lidam com entrada não controlada ou formatos complexos.
