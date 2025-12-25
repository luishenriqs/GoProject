# Camada de Infraestrutura (Persistência) — `infra/database/`

A pasta `infra/database/` representa a **camada de infraestrutura responsável pelo acesso ao banco de dados**.  
Ela implementa o “CRUD real”, usando **GORM** para persistir e consultar dados das entidades da API.

Em uma arquitetura em camadas, esta pasta fica “embaixo” dos usecases:
- **usecases** definem o *roteiro* do que precisa acontecer (validar → persistir → retornar);
- **infra/database** executa a parte de **persistência**, por meio de **repositórios**.

---

## Responsabilidade desta camada

- **Abrir/fornecer conexão com o banco** conforme a configuração do projeto.
- **Implementar repositórios** que fazem operações de leitura/escrita (CRUD) no banco via GORM.
- Manter a camada de aplicação **testável**, porque os usecases conversam com o banco por meio de **interfaces de repositório** (e não acessam GORM diretamente).

> Regra prática: o domínio (`internal/entity`) não sabe nada de banco.  
> Quem “fala” com o banco é esta camada (`infra/database`).

---

## O que fica aqui (e o que não fica)

### Fica aqui
- Código que depende de **GORM**.
- Código que depende do **driver** do banco (SQLite/MySQL) e da **configuração** de conexão.
- Implementações concretas de repositório para as entidades (`User`, `Product`).

### Não fica aqui
- Regras de validação de negócio (isso é do domínio).
- Decisão de endpoints e regras HTTP (isso é da camada `api/`).
- Orquestração de casos de uso (isso é da camada `internal/usecase/`).

---

## Repositórios: como essa camada é consumida pelos usecases

Os **usecases** não dependem de GORM diretamente.  
Eles chamam métodos de repositório (definidos por **interfaces**) para realizar persistência e consulta.

Isso tem dois impactos importantes:

1. **Separação de responsabilidades**  
   O usecase não precisa “saber SQL” nem detalhes do GORM. Ele só pede: “salve este usuário”, “liste produtos”, “busque por ID”, etc.

2. **Testabilidade**  
   Em testes unitários de usecases, você pode substituir o repositório real por um mock/stub que implemente a mesma interface, sem precisar de banco de verdade.

---

## Bancos suportados e estratégia do projeto

A persistência é feita via GORM com suporte a dois cenários:

- **SQLite**: utilizado como base para **dev/test/E2E**.
- **MySQL**: utilizado em runtime quando configurado via `.env`.

A escolha do driver é feita pela variável `DB_DRIVER`.

---

## Configuração usada pela camada de banco

A API usa variáveis de ambiente para configurar o banco. Exemplo de `.env`:

```env
DB_DRIVER=mysql
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=root
DB_NAME=apiGoDB
```

Quando o driver for SQLite (especialmente em testes), a configuração segue o padrão do projeto para rodar localmente e em E2E.

> Observação: o carregamento de configuração do runtime deve usar `configs.LoadConfig` do pacote `configs/`, evitando imports duplicados de config.

---

## Como essa camada se encaixa no fluxo “de ponta a ponta”

Um request típico que chega via HTTP passa por este caminho:

1. **`api/routes.go`** registra a rota no `ServeMux` e aplica middleware nas rotas protegidas.
2. **middleware JWT** valida token e injeta `userID` no `context` (quando necessário).
3. **handler** decodifica JSON, monta inputs e chama o **usecase**.
4. **usecase** orquestra a ação e chama o **repositório**.
5. **repositório (infra/database)** executa a operação no banco via **GORM** e retorna o resultado ao usecase.
6. Handler traduz o retorno em status/JSON de resposta.

---

## Pontos de atenção (para evitar regressões)

- **Driver e ambiente**  
  Se `DB_DRIVER` estiver configurado para `mysql`, a API precisa de um MySQL acessível conforme `DB_HOST/DB_PORT/...`.  
  Para desenvolvimento e testes, o projeto utiliza SQLite como caminho mais simples.

- **Separação por camadas**  
  Evite que handlers/usecases “vazem” dependência de GORM. A regra do projeto é manter GORM confinado em `infra/database/`.

- **Config duplicada**  
  O runtime deve usar apenas `configs.LoadConfig` do pacote `configs/` para carregar `.env` e parâmetros de execução.

---

## Relação com as demais camadas

- **Consome**: entidades do domínio (`internal/entity`) como modelo de dados.
- **Implementa**: repositórios usados pelos usecases (`internal/usecase`) via interfaces.
- **É instanciada**: no composition root (`cmd/server/main.go`), onde a aplicação monta DB → repos → usecases → handlers → rotas.
