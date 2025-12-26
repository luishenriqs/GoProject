# Camada de Inicialização (Composition Root) — `cmd/server/`

A pasta `cmd/server/` contém o **ponto de entrada** da aplicação HTTP: o `main.go`.  
Essa camada é o **Composition Root** do projeto: o lugar onde a API é “montada”, conectando todas as camadas (config → banco → repositórios → usecases → handlers → rotas → servidor HTTP).

> Em outras palavras: as camadas “não se montam sozinhas”.  
> Quem faz o *wiring* (injeção de dependências manual) é o `cmd/server/main.go`.

---

## O que esta camada faz

No `main.go`, o fluxo principal é:

1. **Carregar configurações** (ex.: variáveis de ambiente) via `configs.LoadConfig(".")`.
2. **Abrir conexão com o banco** conforme a variável `DB_DRIVER` (ex.: `mysql` ou `sqlite`).
3. **Instanciar dependências** na ordem correta:
   - conexão/DB
   - repositórios (`infra/database`)
   - usecases (`internal/usecase`)
   - handlers (`api/handlers`)
4. **Registrar rotas** no `http.ServeMux` por meio de `api/routes.go` (aplicando middleware quando necessário).
5. **Subir o servidor HTTP** (iniciando o `net/http` no endereço/porta configurados).

Essa organização mantém as regras do projeto claras:
- o domínio não depende de nada “de fora”;
- a camada `api/` não decide como o sistema é montado;
- a camada `infra/database` não conhece handlers ou rotas;
- **`cmd/server` é o único ponto que conhece tudo e conecta tudo**.

---

## Por que `cmd/server` é importante

### 1) Isola a composição das camadas
Se você espalhar inicialização e dependências por todo o código, o acoplamento aumenta e testes ficam mais difíceis.  
Concentrando o wiring em `cmd/server`, as camadas continuam com responsabilidades bem definidas.

### 2) Facilita testes e evolução
- Você consegue trocar implementações (ex.: repositório real vs mock) de forma controlada.
- Você consegue evoluir o roteamento, middleware ou banco sem “vazar” dependências para o domínio/usecases.

---

## O que você deve encontrar no `main.go`

### Carregamento de configuração
O `main.go` usa `configs.LoadConfig(".")` para carregar a configuração de execução.

Esse ponto é onde o projeto decide:
- driver do banco (`DB_DRIVER`)
- credenciais e parâmetros de conexão (`DB_HOST`, `DB_PORT`, etc.)
- porta do servidor (quando aplicável no projeto)

### Conexão com banco
Após carregar config, o `main.go` abre a conexão via o módulo `infra/database`, respeitando `DB_DRIVER`:
- **mysql**: usa host/porta/usuário/senha/nome do banco
- **sqlite**: utiliza o caminho/arquivo conforme padrão do projeto (muito usado em dev/test)

### Montagem das dependências
O `main.go` instancia os componentes em cadeia:

- repositórios concretos (infra)  
- usecases (aplicação)  
- handlers (web)  
- rotas (`api/routes.go`)  

Isso evita que handlers precisem criar seus próprios usecases, ou que usecases criem repositórios diretamente.

### Inicialização do servidor HTTP
Por fim, o `main.go` sobe o servidor com `net/http`, usando o `ServeMux` configurado pelas rotas.

---

## Relação com as demais camadas

- **Usa** `configs/` para carregar variáveis de ambiente/config.
- **Usa** `infra/database/` para abrir DB e construir repositórios.
- **Usa** `internal/usecase/` para orquestrar as regras da aplicação.
- **Usa** `api/handlers`, `api/middleware` e `api/routes.go` para expor a API via HTTP.
- **Não contém regra de negócio**: apenas composição e inicialização.

---

## Pontos de atenção (para manter o padrão do projeto)

- Evite “pular” camadas: handlers não devem falar com GORM; usecases não devem depender de HTTP.
- O `main.go` é o local correto para:
  - criar dependências concretas;
  - conectar interfaces a implementações;
  - ligar rotas e iniciar o servidor.
- Qualquer mudança estrutural no wiring deve manter o projeto simples e previsível, já que o objetivo é aprendizado e referência acadêmica.
