# Testify (Assert) em Go

## 🔹 O que é a biblioteca Testify
[Testify](https://github.com/stretchr/testify) é uma biblioteca popular para Go que fornece utilitários de testes que vão além do pacote nativo `testing`.  
Entre seus recursos, o mais usado é o **assert**, que facilita a escrita e leitura dos testes.

---

## 🔹 Por que usar Assert
No Go puro, normalmente validamos assim:

```go
if got != want {
    t.Errorf("esperado %v, obtido %v", want, got)
}
```

Com Testify:

```go
assert.Equal(t, want, got)
```

Benefícios:
- **Clareza**: menos código repetitivo, leitura mais simples.  
- **Mensagens melhores**: falhas mostram diferenças de forma clara.  
- **Padronização**: mantém testes consistentes em todo o projeto.  
- **Produtividade**: reduz boilerplate, agiliza a escrita de testes.

---

## 🔹 Principais métodos Assert
- `assert.Equal(t, expected, actual)` → compara valores.  
- `assert.NotEqual(t, unexpected, actual)` → garante que sejam diferentes.  
- `assert.Nil(t, object)` / `assert.NotNil(t, object)` → valida ponteiros, erros etc.  
- `assert.True(t, condition)` / `assert.False(t, condition)` → valida booleanos.  
- `assert.Error(t, err)` / `assert.NoError(t, err)` → simplifica validação de erros.  

---

## 🔹 Exemplo prático
```go
func TestCalculateTax(t *testing.T) {
    got := CalculateTax(1000)
    want := 10.0

    assert.Equal(t, want, got, "Taxa para 1000 deve ser 10.0")
}
```

---

## 🔹 Resumindo
- **Conceito**: Testify assert facilita e padroniza comparações em testes.  
- **Importância**: aumenta legibilidade, reduz boilerplate e torna falhas mais claras.  
- **Benefício real**: acelera o desenvolvimento de testes confiáveis em Go.
