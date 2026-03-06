package events

import (
	"errors"
	"slices"
	"sync"
)

/*
Event Dispatcher (Observer Pattern) — Visão Geral

Este arquivo implementa um "Event Dispatcher" simples para o pacote `events`, seguindo
um padrão semelhante ao Observer:
- O dispatcher mantém um registro (map) de *handlers* por nome de evento.
- Um *handler* (EventHandlerInterface) pode ser registrado para um determinado
evento (eventName).
- Quando um evento (EventInterface) é despachado, todos os handlers registrados
para aquele nome são executados.
- Também é possível remover handlers, verificar se um handler está registrado e
limpar todos os registros.

Estrutura principal:
- `EventDispatcher.handlers`: map[string][]EventHandlerInterface
  Chave: nome do evento
  Valor: lista de handlers associados ao evento
*/

var ErrHandlerAlreadyRegistered = errors.New("handler already registered")

type EventDispatcher struct {
	handlers map[string][]EventHandlerInterface
}

/*
NewEventDispatcher cria uma nova instância de EventDispatcher.

- Inicializa o map interno `handlers` para armazenar os handlers por nome de evento.
- Deve ser chamado antes de usar Register/Dispatch/Remove/Has/Clear.
*/
func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		handlers: make(map[string][]EventHandlerInterface),
	}
}

/*
Register registra um handler para um determinado nome de evento.

Regras e comportamento:
  - Se já existir uma lista de handlers para o `eventName`, o método verifica se o mesmo `handler`
    já está registrado (usando `slices.Contains`).
  - Caso já esteja registrado, retorna `ErrHandlerAlreadyRegistered`.
  - Caso contrário, adiciona o handler ao slice associado ao `eventName`.

Efeito:
  - Após o registro, o handler passará a ser chamado quando `Dispatch` for executado com um evento
    cujo `GetName()` seja igual a `eventName`.
*/
func (ed *EventDispatcher) Register(eventName string, handler EventHandlerInterface) error {
	if _, ok := ed.handlers[eventName]; ok {
		if slices.Contains(ed.handlers[eventName], handler) {
			return ErrHandlerAlreadyRegistered
		}
	}

	ed.handlers[eventName] = append(ed.handlers[eventName], handler)
	return nil
}

/*
Dispatch executa todos os handlers registrados para o nome do evento informado e
aguarda sua conclusão via `sync.WaitGroup`.

Regras e comportamento:
- Obtém o nome do evento via `event.GetName()`.
- Procura na estrutura interna (`handlers`) a lista de handlers registrada para esse nome.
- Se existir, cria um `WaitGroup`, incrementa (`wg.Add(1)`) para cada handler e
invoca `handler.Handle(event, wg)`.
- Aguarda a finalização de todos os handlers com `wg.Wait()` antes de retornar.

Observações:
- Se não existir nenhum handler registrado para o evento, o método não executa nada e retorna `nil`.
- A ordem de execução segue a ordem em que os handlers foram registrados (append no slice).
*/
func (ev *EventDispatcher) Dispatch(event EventInterface) error {
	if handlers, ok := ev.handlers[event.GetName()]; ok {
		wg := &sync.WaitGroup{}
		for _, handler := range handlers {
			wg.Add(1)
			handler.Handle(event, wg)
		}
		wg.Wait()
	}
	return nil
}

/*
Remove remove um handler associado a um determinado nome de evento.

Regras e comportamento:
- Verifica se existe entrada para `eventName` no map.
- Se existir, percorre o slice de handlers buscando uma referência exatamente igual (`h == handler`).
- Ao encontrar, remove o elemento do slice usando `append(slice[:i], slice[i+1:]...)`.
- Retorna `nil` em qualquer cenário (encontrando ou não).

Efeito:
- Após a remoção, o handler não será mais executado em `Dispatch` para aquele `eventName`.
*/
func (ed *EventDispatcher) Remove(eventName string, handler EventHandlerInterface) error {
	if _, ok := ed.handlers[eventName]; ok {
		for i, h := range ed.handlers[eventName] {
			if h == handler {
				ed.handlers[eventName] = append(ed.handlers[eventName][:i], ed.handlers[eventName][i+1:]...)
				return nil
			}
		}
	}
	return nil
}

/*
Has verifica se um handler está registrado para um determinado nome de evento.

Regras e comportamento:
- Confere se existe uma lista de handlers para `eventName`.
- Se existir, usa `slices.Contains` para checar se o `handler` está presente no slice.
- Retorna `true` se estiver registrado; caso contrário, `false`.

Uso típico:
- Validar estado do dispatcher em testes ou fluxos de controle (ex.: antes de remover ou registrar).
*/
func (ed *EventDispatcher) Has(eventName string, handler EventHandlerInterface) bool {
	if _, ok := ed.handlers[eventName]; ok {
		if slices.Contains(ed.handlers[eventName], handler) {
			return true
		}
	}
	return false
}

/*
Clear remove todos os handlers registrados no dispatcher.

Regras e comportamento:
- Recria o map `handlers` do zero, descartando todas as entradas anteriores.
- Após o Clear, nenhum evento terá handlers registrados até novos `Register`.
*/
func (ed *EventDispatcher) Clear() {
	ed.handlers = make(map[string][]EventHandlerInterface)
}
