package events

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

/*
Testes do EventDispatcher — Visão Geral

Este arquivo valida o comportamento do `EventDispatcher` (registrar, despachar, remover, consultar e limpar handlers),
usando o framework `testify/suite` para organizar os cenários de teste em uma suíte com estado compartilhado.

Estratégia:
- Define estruturas de teste (`TestEvent`, `TestEventHandler`) que implementam as interfaces do pacote
  (`EventInterface` e `EventHandlerInterface`) para simular eventos e handlers reais.
- Usa `SetupTest()` para resetar o estado antes de cada teste (dispatcher novo e dados determinísticos).
- Para o teste de `Dispatch`, utiliza um mock (`MockHandler`) com `testify/mock` para garantir que `Handle()`
  é chamado o número esperado de vezes e com os argumentos corretos.
*/

type TestEvent struct {
	Name    string
	Payload interface{}
}

/*
GetName retorna o nome do evento, usado como chave no map interno do dispatcher.

- Este método existe para satisfazer `EventInterface`.
- O dispatcher usa este valor para localizar a lista de handlers a ser executada no `Dispatch`.
*/
func (e *TestEvent) GetName() string {
	return e.Name
}

/*
GetPayload retorna o payload associado ao evento.

- Este método existe para satisfazer `EventInterface`.
- Nos testes deste arquivo, o payload é usado apenas como dado de suporte.
*/
func (e *TestEvent) GetPayload() interface{} {
	return e.Payload
}

/*
GetDateTime retorna o timestamp do evento.

- Este método existe para satisfazer `EventInterface`.
- Aqui, usa `time.Now()` por simplicidade, pois os testes não validam especificamente o horário.
*/
func (e *TestEvent) GetDateTime() time.Time {
	return time.Now()
}

type TestEventHandler struct {
	ID int
}

/*
Handle é a ação executada pelo dispatcher quando um evento é despachado.

  - Este handler "real" de teste não executa nada (corpo vazio),
    sendo usado principalmente para testes de Register/Remove/Has/Clear.
*/
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

/*
SetupTest é executado antes de cada teste da suíte.

Objetivos:
- Criar uma nova instância de `EventDispatcher` para garantir isolamento entre testes.
- Recriar eventos e handlers com valores previsíveis.
- Evitar efeitos colaterais: cada teste começa com um dispatcher "limpo".
*/
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

/*
TestEventDispatcher_Register valida o registro de múltiplos handlers no mesmo evento.

Cenário:
- Registra dois handlers distintos para o mesmo `eventName`.
- Verifica ausência de erro, tamanho da lista e ordem dos handlers registrados.
*/
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

/*
TestEventDispatcher_Register_WithSameHandler valida a proteção contra registros duplicados.

Cenário:
- Registra um handler para um evento.
- Tenta registrar o mesmo handler novamente para o mesmo evento.
- Espera `ErrHandlerAlreadyRegistered` e mantém o tamanho da lista inalterado.
*/
func (suite *EventDispatcherTestSuite) TestEventDispatcher_Register_WithSameHandler() {
	err := suite.eventDispatcher.Register(suite.event.GetName(), &suite.handler)
	suite.Nil(err)
	suite.Equal(1, len(suite.eventDispatcher.handlers[suite.event.GetName()]))

	err = suite.eventDispatcher.Register(suite.event.GetName(), &suite.handler)
	suite.Equal(ErrHandlerAlreadyRegistered, err)
	suite.Equal(1, len(suite.eventDispatcher.handlers[suite.event.GetName()]))
}

type MockHandler struct {
	mock.Mock
}

/*
Handle é a implementação mockada do handler para capturar chamadas via `testify/mock`.

- Permite verificar se `Dispatch` realmente invoca `Handle(event)` com o evento esperado.
*/
func (m *MockHandler) Handle(event EventInterface) {
	m.Called(event)
}

/*
TestEventDispacth_Dispatch valida que `Dispatch` chama o handler registrado para um evento.

Cenário:
- Cria um handler mock, configura expectativa de chamada `Handle(&suite.event)`.
- Registra o mock para o evento.
- Executa `Dispatch(&suite.event)`.
- Verifica expectativas e garante exatamente 1 chamada ao método `Handle`.
*/
func (suite *EventDispatcherTestSuite) TestEventDispacth_Dispatch() {
	eh := &MockHandler{}
	eh.On("Handle", &suite.event)

	suite.eventDispatcher.Register(suite.event.GetName(), eh)
	suite.eventDispatcher.Dispatch(&suite.event)

	eh.AssertExpectations(suite.T())
	eh.AssertNumberOfCalls(suite.T(), "Handle", 1)
}

/*
TestEventDispatcher_Remove valida a remoção de handlers por evento, preservando os demais.

Cenário:
- Event 1: registra dois handlers.
- Event 2: registra um handler.
- Remove o primeiro handler do Event 1 e valida que sobra apenas o segundo.
- Remove o segundo handler do Event 1 e valida lista vazia.
- Remove o handler do Event 2 e valida lista vazia.
*/
func (suite *EventDispatcherTestSuite) TestEventDispatcher_Remove() {
	// Event 1
	err := suite.eventDispatcher.Register(suite.event.GetName(), &suite.handler)
	suite.Nil(err)
	suite.Equal(1, len(suite.eventDispatcher.handlers[suite.event.GetName()]))

	err = suite.eventDispatcher.Register(suite.event.GetName(), &suite.handler2)
	suite.Nil(err)
	suite.Equal(2, len(suite.eventDispatcher.handlers[suite.event.GetName()]))

	// Event 2
	err = suite.eventDispatcher.Register(suite.event2.GetName(), &suite.handler3)
	suite.Nil(err)
	suite.Equal(1, len(suite.eventDispatcher.handlers[suite.event2.GetName()]))

	suite.eventDispatcher.Remove(suite.event.GetName(), &suite.handler)
	suite.Equal(1, len(suite.eventDispatcher.handlers[suite.event.GetName()]))
	assert.Equal(suite.T(), &suite.handler2, suite.eventDispatcher.handlers[suite.event.GetName()][0])

	suite.eventDispatcher.Remove(suite.event.GetName(), &suite.handler2)
	suite.Equal(0, len(suite.eventDispatcher.handlers[suite.event.GetName()]))

	suite.eventDispatcher.Remove(suite.event2.GetName(), &suite.handler3)
	suite.Equal(0, len(suite.eventDispatcher.handlers[suite.event2.GetName()]))
}

/*
TestEventDispatcher_Has valida o método `Has`, que verifica se um handler está registrado.

Cenário:
- Registra dois handlers no Event 1.
- Verifica `Has == true` para os handlers registrados.
- Verifica `Has == false` para um handler não registrado.
*/
func (suite *EventDispatcherTestSuite) TestEventDispatcher_Has() {
	err := suite.eventDispatcher.Register(suite.event.GetName(), &suite.handler)
	suite.Nil(err)
	suite.Equal(1, len(suite.eventDispatcher.handlers[suite.event.GetName()]))

	err = suite.eventDispatcher.Register(suite.event.GetName(), &suite.handler2)
	suite.Nil(err)
	suite.Equal(2, len(suite.eventDispatcher.handlers[suite.event.GetName()]))

	assert.True(suite.T(), suite.eventDispatcher.Has(suite.event.GetName(), &suite.handler))
	assert.True(suite.T(), suite.eventDispatcher.Has(suite.event.GetName(), &suite.handler2))
	assert.False(suite.T(), suite.eventDispatcher.Has(suite.event.GetName(), &suite.handler3))
}

/*
TestEventDispatcher_Clear valida o método `Clear`, que remove todos os registros do dispatcher.

Cenário:
- Registra handlers em dois eventos diferentes.
- Executa `Clear()`.
- Valida que o map `handlers` foi zerado (tamanho 0).
*/
func (suite *EventDispatcherTestSuite) TestEventDispatcher_Clear() {
	// Event 1
	err := suite.eventDispatcher.Register(suite.event.GetName(), &suite.handler)
	suite.Nil(err)
	suite.Equal(1, len(suite.eventDispatcher.handlers[suite.event.GetName()]))

	err = suite.eventDispatcher.Register(suite.event.GetName(), &suite.handler2)
	suite.Nil(err)
	suite.Equal(2, len(suite.eventDispatcher.handlers[suite.event.GetName()]))

	// Event 2
	err = suite.eventDispatcher.Register(suite.event2.GetName(), &suite.handler3)
	suite.Nil(err)
	suite.Equal(1, len(suite.eventDispatcher.handlers[suite.event2.GetName()]))

	suite.eventDispatcher.Clear()
	suite.Equal(0, len(suite.eventDispatcher.handlers))
}

/*
TestSuite é o ponto de entrada do Go test para executar a suíte do `testify/suite`.

- Registra `EventDispatcherTestSuite` para que todos os testes (methods `Test*`) sejam executados.
*/
func TestSuite(t *testing.T) {
	suite.Run(t, new(EventDispatcherTestSuite))
}
