# Testify (Mock) em Go

## 🔹 O que é Mock
Mocks são objetos falsos que imitam o comportamento de dependências externas (como banco de dados, serviços externos, APIs).  
Em testes, usamos mocks para **isolar a lógica interna** da função testada, sem precisar de integrações reais.

---

## 🔹 O papel do Testify Mock
O pacote `testify/mock` fornece ferramentas para criar **mocks automatizados** em Go.  
Ele permite:
- Definir expectativas (quais métodos devem ser chamados e com quais parâmetros).  
- Retornar valores simulados (sucesso, erro etc.).  
- Verificar se as chamadas aconteceram como esperado.

---

## 🔹 Por que usar Mock
- **Isolamento**: testamos apenas a lógica da função, sem depender de serviços externos.  
- **Controle**: simulamos erros e cenários difíceis de reproduzir em ambiente real.  
- **Velocidade**: elimina latência de chamadas externas.  
- **Confiabilidade**: torna os testes determinísticos (sempre o mesmo resultado).

---

## 🔹 Exemplo prático
```go
type Repository interface {
    SaveTax(amount float64) error
}

type RepositoryMock struct {
    mock.Mock
}

func (m *RepositoryMock) SaveTax(amount float64) error {
    args := m.Called(amount)
    return args.Error(0)
}
```

Teste com mock:

```go
func TestCalculateTaxAndSave(t *testing.T) {
    repo := new(RepositoryMock)
    repo.On("SaveTax", 10.0).Return(nil).Once()

    got, err := CalculateTaxAndSave(1000, repo)

    assert.NoError(t, err)
    assert.Equal(t, 10.0, got)
    repo.AssertExpectations(t)
}
```

---

## 🔹 Resumindo
- **Conceito**: mocks simulam dependências externas.  
- **Importância**: permitem testar funções isoladamente, garantindo previsibilidade.  
- **Benefício real**: aumentam a qualidade dos testes, possibilitando cobrir cenários de sucesso e falha sem depender de sistemas externos.
