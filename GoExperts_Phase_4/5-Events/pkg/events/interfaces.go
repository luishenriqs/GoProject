package events

import "time"

type EventInterface interface {
	GetName() string
	GetDateTime() time.Time
	GetPayload() interface{}
}

type EventHandlerInterface interface {
	Handle(event EventInterface)
}

type EventDispatcherInterface interface {
	Register(eventName string, handler EventHandlerInterface) error
	Dispatch(event EventInterface) error
	Remove(eventName string, handler EventHandlerInterface) error
	Has(eventName string, handler EventHandlerInterface) bool
	Clear() error
}

/*
# Módulo de Eventos em Go — conceitos e desenho do Dispatcher/Handlers

Neste módulo, a ideia é modelar **eventos** como “algo aconteceu” no sistema e, a partir disso,
executar **operações** desacopladas (handlers).
Em vez de um fluxo rígido e linear (“faz A, depois B, depois C”), você dispara um evento e o
sistema executa automaticamente tudo o que estiver **registrado** para reagir àquele evento.

Um exemplo típico:

* **Evento:** “Usuário cadastrado”
* **Operações (handlers):**

  * Notificar esse usuário no **Slack**
  * Inserir esse usuário no **Salesforce**
  * (e outras operações futuras, sem mexer no ponto onde o evento é disparado)

Essa abordagem é útil quando você quer:

* reduzir acoplamento entre partes do sistema;
* adicionar/remover integrações sem alterar o core;
* organizar integrações e efeitos colaterais como “reações a eventos”.

---

## Conceitos centrais

### 1) Evento

Um **evento** carrega as informações necessárias para as operações reagirem:

* **Nome** do evento (ex.: `"UserCreated"`, `"LoadData"`)
* **Data/hora** do evento (para auditoria, ordenação, logs)
* **Payload** (os dados relevantes: usuário, pedido, arquivo importado etc.)

A primeira imagem do professor resume bem esse raciocínio:

* **Evento (Carregar dados)**
* **Operações que serão executadas quando um evento é chamado**
* **Gerenciador dos nossos eventos/operações**

  * Registrar os eventos e suas operações
  * Despachar / “Fire” o evento para suas operações serem executadas

---

### 2) Handler (Operação)

Um **handler** é uma unidade de reação: uma implementação que recebe o evento e executa sua responsabilidade.

Exemplos de handlers para o mesmo evento:

* `SlackNotificationHandler`
* `SalesforceCreateLeadHandler`

O ponto-chave é que cada handler:

* conhece **apenas** o evento recebido;
* não precisa saber quem mais vai executar;
* pode ser adicionado/removido sem mudar o código que dispara o evento.

---

### 3) Dispatcher (Gerenciador de Eventos)

O **dispatcher** (ou “event manager”) é quem controla o ciclo:

1. **Register**: registra handlers associados a um `eventName`
2. **Dispatch**: ao receber um evento, executa todos os handlers registrados para o nome daquele evento
3. **Remove / Has / Clear**: utilitários para manutenção do registro

---

## As interfaces do módulo (`interfaces.go`)

A segunda imagem mostra a base do desenho: tudo começa com três contratos (interfaces), para manter o sistema extensível e testável.

```go
package events

import "time"

type EventInterface interface {
	GetName() string
	GetDateTime() time.Time
	GetPayload() interface{}
}

type EventHandlerInterface interface {
	Handle(event EventInterface)
}

type EventDispatcherInterface interface {
	Register(eventName string, handler EventHandlerInterface) error
	Dispatch(event EventInterface) error
	Remove(eventName string, handler EventHandlerInterface) error
	Has(eventName string, handler EventHandlerInterface) bool
	Clear() error
}
```

### O que cada interface garante

**EventInterface**

* padroniza como ler nome, timestamp e dados (payload);
* permite que handlers trabalhem com qualquer evento que respeite esse contrato.

**EventHandlerInterface**

* define o formato único de uma operação reativa: `Handle(event)`;
* mantém o dispatcher independente das operações concretas.

**EventDispatcherInterface**

* define o “coração” do módulo: registrar, despachar e administrar handlers por evento;
* permite trocar implementações (ex.: dispatcher em memória, dispatcher thread-safe, dispatcher assíncrono etc.) sem mudar o restante do sistema.

---

## Fluxo mental do módulo (o que “acontece” quando um evento é disparado)

1. Em algum ponto do sistema, você cria um evento (ex.: `LoadDataEvent`).
2. Antes disso (na inicialização do sistema), você registra handlers:

   * `Register("LoadData", handlerA)`
   * `Register("LoadData", handlerB)`
3. Quando `Dispatch(event)` é chamado:

   * o dispatcher encontra os handlers do `event.GetName()`;
   * executa cada `Handle(event)`.

Esse é o motivo do design ser tão útil: **você não precisa acoplar “carregar dados” com “notificar slack” e “inserir no salesforce”** no mesmo trecho de código. Você só dispara o evento.

---

## Boas práticas que normalmente entram na discussão (ainda sem implementar nada além do contrato)

Mesmo antes de codar o dispatcher concreto, vale ter claros alguns pontos que influenciam a implementação:

* **Ordem de execução:** handlers precisam rodar em ordem fixa ou tanto faz?
* **Tratamento de erro:** `Handle` não retorna erro na interface atual — então erros podem virar:

  * logs internos no handler; ou
  * evolução futura da interface (se o curso pedir).
* **Sincronia vs assincronia:** `Dispatch` pode executar handlers em sequência (simples) ou em goroutines (mais complexo).
* **Idempotência:** handlers de integração (Salesforce/Slack) frequentemente precisam ser seguros para reexecução.

Essas decisões geralmente aparecem quando o professor começa a implementar o dispatcher e os handlers reais.

---

## Resultado do desenho até aqui

Com essas interfaces, você já tem a base para:

* criar tipos concretos de eventos (com payload específico);
* criar handlers independentes por integração/efeito colateral;
* criar um dispatcher em memória que registra handlers por nome e dispara quando necessário;
* testar tudo isolado (mockando dispatcher ou handlers via interface).

*/
