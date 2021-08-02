package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/RaveNoX/go-jsonmerge"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"log"
	"sync"
	"time"
)

// ErrEventRejected is the error returned when the state machine cannot process
// an event in the state that it is in.
var ErrEventRejected = errors.New("event rejected")

const (
	// Default represents the default state of the system.
	Default StateType = ""

	// NoOp represents a no-op event.
	NoOp EventType = "NoOp"
)

// StateType represents an extensible state type in the state machine.
type StateType string

// EventType represents an extensible event type in the state machine.
type EventType string

// WorkflowContext represents the context held by the state machine.
type WorkflowContext map[string]interface{}

// EventContext represents the context to be passed to the action implementation.
type EventContext map[string]interface{}

// Events represents a mapping of events and states.
type Events map[EventType]StateType

// State binds a state with an action and a set of events it can handle.
type State struct {
	Events Events `json:"events,omitempty"`
}


// States represents a mapping of states and their implementations.
type States map[StateType]State

// Workflow represent a workflow definition, just a set of states
type Workflow struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	States  States `json:"states"`
}

type RunnerInfo struct {
	Id string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`

}

// StateMachine represents the state machine.
type StateMachine struct {
	Id string `json:"id"`
	// Previous represents the previous state.
	Previous StateType `json:"-"`

	// Current represents the current state.
	Current StateType `json:"current"`

	// States holds the configuration of states and events handled by the state machine.
	States States `json:"-"`

	// mutex ensures that only 1 event is processed by the state machine at any given time.
	mutex sync.Mutex `json:"-"`

	WorkflowContext WorkflowContext `json:"context"`

	// Event SINK to emit change of state
	SINK string `json:"-"`
}

// getNextState returns the next state for the event given the machine's current
// state, or an error if the event can't be handled in the given state.
func (s *StateMachine) getNextState(event EventType) (StateType, error) {
	if state, ok := s.States[s.Current]; ok {
		if state.Events != nil {
			if next, ok := state.Events[event]; ok {
				return next, nil
			}
		}
	}
	return Default, ErrEventRejected
}

// SendEvent sends an event to the state machine.
func (s *StateMachine) SendEvent(event EventType, eventCtx EventContext) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for {

		log.Printf("Event Received: %s", event)
		log.Printf("Current State: %s", s.States[s.Current])
		log.Printf("Available Events in State: %s", s.States[s.Current].Events)

		// Determine the next state for the event given the machine's current state.
		nextState, err := s.getNextState(event)
		if err != nil {
			log.Printf("Event Rejected: %s for state %s ", event, s.States[s.Current] )
			return ErrEventRejected
		}
		log.Printf("Next State: %s", nextState)

		// Here I am going to change the state, so I need to emit the Post Events.
		//// Identify the state definition for the next state.
		// I might need this if I want to read the state definition to emit a specific event
		//state, ok := s.States[nextState]
		//if !ok  {
		//	// configuration error
		//	return nil
		//}

		// Transition over to the next state.
		s.Previous = s.Current
		s.Current = nextState

		if s.WorkflowContext == nil {
			s.WorkflowContext = WorkflowContext(eventCtx)

			log.Printf("Workflow Context JSON %s", s.WorkflowContext)
		} else {
			log.Printf("Event Context JSON %s \n", eventCtx)
			log.Printf("Workflow Context JSON %s \n ", s.WorkflowContext)
			merged, info := jsonmerge.Merge(eventCtx, s.WorkflowContext)

			log.Printf("Replacements JSON %+v \n", info)
			log.Printf("Workflow Context Result JSON %s \n", merged)
			s.WorkflowContext = WorkflowContext(merged.(EventContext))
		}
		// I have changed the state, so i need to emit the Pre Events


		// Emit Workflow State Change Event
		s.emitCloudEvent()

	}
}

func (s *StateMachine) emitCloudEvent() error {



	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	workflowContextJson, err := json.Marshal(s.WorkflowContext)

	if err != nil {
		log.Printf("failed serialize workflowContext %s", err)
	}

	log.Printf("Workflow Context JSON %s", string(workflowContextJson))

	// Create an Event.
	event := cloudevents.NewEvent()
	newUUID, _ := uuid.NewUUID()
	event.SetID(newUUID.String())
	event.SetTime(time.Now())
	event.SetSource("workflow")
	event.SetType("workflow.event")
	marshal, err := json.Marshal(s)
	if err != nil {
		return err
	}
	event.SetData(cloudevents.ApplicationJSON, marshal)

	log.Printf("Emitting an Event: %s to SINK: %s", event, s.SINK)


	// Set a target.
	ctx := cloudevents.ContextWithTarget(context.Background(), s.SINK)

	// Send that Event.
	result := c.Send(ctx, event)
	if result != nil {
		log.Printf("Resutl: %s", result)
		if cloudevents.IsUndelivered(result) {
			log.Printf("failed to send, %v", result)
		}
	}

	return nil
}
