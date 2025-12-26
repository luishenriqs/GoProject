# Pacotes Reutilizáveis — `pkg/`

A pasta `pkg/` existe para concentrar **código reutilizável** que pode ser consumido:
- **internamente**, por módulos da própria API; e
- **externamente**, por outros projetos (caso você queira reaproveitar utilitários “genéricos” fora desta API).

A ideia é que tudo que estiver em `pkg/` seja **independente do domínio específico** da aplicação (User/Product) e tenha valor como componente reutilizável.

---

## Conteúdo atual

Atualmente, `pkg/` possui apenas um submódulo:

- `pkg/entity/`
  - `id.go`

### `pkg/entity/id.go`
Este arquivo concentra a funcionalidade de **geração de IDs**, exposta como um utilitário de uso geral.

Ele é utilizado pelas entidades do domínio (ex.: `Product`) para criar IDs de forma padronizada, sem acoplar essa responsabilidade ao domínio em si.

---

## Por que isso fica em `pkg/` e não no domínio?

- O domínio (`internal/entity`) modela o “mundo” da API e suas regras (User/Product).
- Já a geração de ID é um **serviço genérico**, que pode ser útil em outros contextos além desta API.

Ao manter essa funcionalidade em `pkg/`, o projeto:
- evita misturar responsabilidade de infraestrutura/utilitários com regras de domínio;
- permite reaproveitar esse código em outros projetos Go, caso necessário.

---

## Objetivo do `pkg/` no projeto

O objetivo do `pkg/` é servir como “biblioteca interna” do repositório, com potencial de reutilização externa.

Exemplos de itens que fazem sentido nessa pasta (no futuro, caso o projeto evolua) são:
- helpers genéricos (sem dependência de HTTP e sem dependência das entidades específicas do projeto),
- utilitários de formatação ou validações genéricas,
- mecanismos comuns (ex.: geração de IDs, pequenas funções auxiliares).

> Importante: só devem entrar em `pkg/` componentes realmente reutilizáveis e genéricos.
> Tudo que for regra de negócio ou específico de User/Product deve continuar em `internal/`.
