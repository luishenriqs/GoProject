package events

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

/*
	Este arquivo contém os testes unitários do EventDispatcher utilizando o pacote testify/suite para
	organizar o ciclo de vida dos testes (SetupTest) e validar o comportamento do registro (Register)
	de handlers por nome de evento.

	Objetivo dos testes atuais:
	- Validar que o método Register adiciona handlers corretamente ao dispatcher para um eventName.
	- Garantir que múltiplos handlers distintos podem ser registrados para o mesmo eventName.
	- Confirmar que a ordem de registro é preservada, permitindo validar a posição de cada handler na lista.

	Estruturas auxiliares de teste:
	- TestEvent: implementação mínima de EventInterface para uso nos testes.
	- GetName() retorna o nome do evento.
	- GetPayload() retorna o payload genérico do evento.
	- GetDateTime() retorna um timestamp (time.Now()).
	- TestEventHandler: implementação mínima de EventHandlerInterface.
	- Contém um campo ID apenas para diferenciar instâncias durante os testes.
	- Handle(event) é vazio porque o foco destes testes é o registro, não o processamento.

	Suite de testes:
	- EventDispatcherTestSuite:
	- Mantém instâncias de eventos (event, event2), handlers (handler, handler2, handler3) e do dispatcher.
	- SetupTest() inicializa um dispatcher novo e recria os handlers e eventos para cada caso, garantindo
		isolamento entre testes.

	Casos de teste:
	- TestEventDispatcher_Register():
	- Registra handler e handler2 no mesmo eventName.
	- Verifica que não houve erro nas operações.
	- Assegura o crescimento da lista (1 e depois 2 handlers).
	- Valida que os handlers armazenados correspondem exatamente aos ponteiros registrados e na ordem correta.

	Execução:
	- TestSuite(t *testing.T) registra a suite no runner do Go.
	- Para executar:
	go test ./...
*/

type TestEvent struct {
	Name    string
	Payload interface{}
}

func (e *TestEvent) GetName() string {
	return e.Name
}

func (e *TestEvent) GetPayload() interface{} {
	return e.Payload
}

func (e *TestEvent) GetDateTime() time.Time {
	return time.Now()
}

type TestEventHandler struct {
	ID int
}

func (h *TestEventHandler) Handle(event EventInterface) {
}

type EventDispatcherTestSuite struct {
	suite.Suite
	event           TestEvent
	event2          TestEvent
	handler         TestEventHandler
	handler2        TestEventHandler
	handler3        TestEventHandler
	eventDispatcher *EventDispatcher
}

func (suite *EventDispatcherTestSuite) SetupTest() {
	suite.eventDispatcher = NewEventDispatcher()
	suite.handler = TestEventHandler{
		ID: 1,
	}
	suite.handler2 = TestEventHandler{
		ID: 2,
	}
	suite.handler3 = TestEventHandler{
		ID: 3,
	}
	suite.event = TestEvent{Name: "test", Payload: "test"}
	suite.event2 = TestEvent{Name: "test2", Payload: "test2"}
}

func (suite *EventDispatcherTestSuite) TestEventDispatcher_Register() {
	err := suite.eventDispatcher.Register(suite.event.GetName(), &suite.handler)
	suite.Nil(err)
	suite.Equal(1, len(suite.eventDispatcher.handlers[suite.event.GetName()]))

	err = suite.eventDispatcher.Register(suite.event.GetName(), &suite.handler2)
	suite.Nil(err)
	suite.Equal(2, len(suite.eventDispatcher.handlers[suite.event.GetName()]))

	assert.Equal(suite.T(), &suite.handler, suite.eventDispatcher.handlers[suite.event.GetName()][0])
	assert.Equal(suite.T(), &suite.handler2, suite.eventDispatcher.handlers[suite.event.GetName()][1])
}

func (suite *EventDispatcherTestSuite) TestEventDispatcher_Register_WithSameHandler() {
	err := suite.eventDispatcher.Register(suite.event.GetName(), &suite.handler)
	suite.Nil(err)
	suite.Equal(1, len(suite.eventDispatcher.handlers[suite.event.GetName()]))

	err = suite.eventDispatcher.Register(suite.event.GetName(), &suite.handler)
	suite.Equal(ErrHandlerAlreadyRegistered, err)
	suite.Equal(1, len(suite.eventDispatcher.handlers[suite.event.GetName()]))
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(EventDispatcherTestSuite))
}
