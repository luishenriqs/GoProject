package events

import "slices"

import "errors"

/*
Package events implementa um mecanismo simples de despacho de eventos (Event Dispatcher) baseado no padrão
Observer/Pub-Sub, permitindo desacoplar o ponto que dispara um evento das operações que reagem a ele.

Este arquivo define o EventDispatcher, responsável por manter o registro de handlers (EventHandlerInterface)
associados a um nome de evento (eventName). Internamente, o dispatcher utiliza um map onde a chave é o nome
do evento e o valor é uma lista de handlers registrados para aquele evento.

Responsabilidades principais:
- Armazenar handlers por tipo/nome de evento.
- Permitir o registro de múltiplos handlers para o mesmo evento.
- Evitar registro duplicado do mesmo handler para o mesmo eventName, retornando o erro
  ErrHandlerAlreadyRegistered quando uma duplicidade é detectada.

Funções/métodos:
- NewEventDispatcher(): cria uma instância do dispatcher com a estrutura interna inicializada.
- (*EventDispatcher).Register(eventName, handler): registra um handler para um evento, garantindo que o
  mesmo handler não seja adicionado duas vezes para o mesmo eventName.

Observação:
Este arquivo cobre apenas a etapa de registro de handlers. O disparo do evento (Dispatch) e operações de
remoção/verificação/limpeza (Remove/Has/Clear) são normalmente implementados em etapas seguintes do módulo.
*/

var ErrHandlerAlreadyRegistered = errors.New("handler already registered")

type EventDispatcher struct {
	handlers map[string][]EventHandlerInterface
}

func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		handlers: make(map[string][]EventHandlerInterface),
	}
}

func (ed *EventDispatcher) Register(eventName string, handler EventHandlerInterface) error {
	if _, ok := ed.handlers[eventName]; ok {
		if slices.Contains(ed.handlers[eventName], handler) {
			return ErrHandlerAlreadyRegistered
		}
	}

	ed.handlers[eventName] = append(ed.handlers[eventName], handler)
	return nil
}
