# Camada de Aplicação — `internal/usecase`

A pasta `internal/usecase` representa a **camada de aplicação** da API. Ela contém os **usecases** (“casos de uso”), que funcionam como **roteiros de execução**: recebem uma intenção (por exemplo, *criar produto*) e coordenam as etapas necessárias para realizar essa ação, de forma organizada e testável.

Em termos práticos, o usecase é quem **orquestra o fluxo**:
1) recebe os **inputs** do caso de uso,  
2) aciona as regras do domínio (validações/comportamentos das entidades),  
3) conversa com o banco **via interfaces de repositório**,  
4) retorna o resultado (ou erro) para a camada Web (handlers).

---

## Responsabilidade desta camada

- **Orquestrar casos de uso** (o “como executar” cada ação do sistema).
- **Manter o domínio independente**: o usecase usa o domínio, mas não mistura regras de HTTP/JSON nem detalhes de banco.
- **Garantir testabilidade**: ao conversar com persistência via **interfaces de repositório**, os usecases podem ser testados com mocks/stubs, sem depender de banco real.

---

## O que fica aqui (e o que não fica)

### Fica aqui
- Implementação dos casos de uso da API (User e Product).
- Coordenação entre domínio e persistência (via repositórios).

### Não fica aqui
- HTTP, JSON, status codes e parsing de request/response (isso é dos handlers).
- Autenticação JWT e middleware (isso é da camada Web).
- Implementação concreta de banco (GORM/SQLite/MySQL) (isso é de `infra/database`).

---

## Ponto importante do estado atual: **usecases não recebem `context.Context`**

No estado atual do projeto, os métodos de usecase **recebem apenas os inputs** do caso de uso — **não recebem `context.Context` na assinatura**.

Isso reforça o papel do usecase como “roteiro” puro do caso de uso, evitando acoplamento direto com detalhes de request (HTTP). Quando alguma informação do request for necessária (por exemplo, identificação do usuário autenticado), ela deve chegar ao usecase como parte dos **inputs** fornecidos pela camada Web.

---

## Usecases disponíveis

A API tem dois conjuntos de usecases, um para cada entidade.

### `UserUseCase`
Casos de uso relacionados ao usuário autenticado e autenticação:

- `RegisterUser`
- `Login`
- `GetMe`
- `UpdateMe`
- `DeleteMe`

### `ProductUseCase`
Casos de uso relacionados ao CRUD de produtos (protegidos por JWT):

- `CreateProduct`
- `ListProducts` *(paginação + ordenação)*
- `GetProduct`
- `UpdateProduct`
- `DeleteProduct`

---

## Como essa camada se encaixa no fluxo da API

De forma resumida:

- A camada **Web** (handlers) recebe a requisição HTTP, decodifica JSON, valida autenticação via middleware e prepara os dados.
- O handler chama o método do **usecase**, passando somente os **inputs** necessários.
- O usecase executa o “roteiro” do caso de uso, usando:
  - o **domínio** (`internal/entity`) para validações/comportamentos,
  - e a persistência via **interfaces de repositório** (implementadas em `infra/database`).
- O handler transforma o retorno do usecase em resposta HTTP.

---

## Por que isso é útil (e o que observar ao evoluir)

- Usecases deixam o código **mais previsível**: cada operação tem um fluxo claro e único.
- A separação por interfaces de repositório deixa o projeto **mais fácil de testar**.
- Mantendo o usecase sem `context.Context`, a regra prática é: **tudo que for “do request” deve ser resolvido antes (na camada Web) e repassado como input explícito**.
