package workflow

import (
	"github.com/google/uuid"
	"testing"
)

const (

	MyState2     StateType = "MyState2"
	MyStateWrong     StateType = "MyStateWrong"

	MoveToState2       EventType = "MoveToState2"
)

type DataContext struct {
	Data map[string]string `json:"data,omitempty"`
	Err   error    `json:"error,omitempty"`
}


type TestDataItem struct{
	input *StateMachine
	events []EventType
	context *DataContext
	result StateType

}

func TestStateMachine(t *testing.T) {
	newUUID, _ := uuid.NewUUID()
	dataItems := []TestDataItem{
		{
					input: &StateMachine{
						Id: newUUID.String(),
						States:  States{
							Default: State{
								Events: Events{
									MoveToState2: MyState2,
								},
							},
							MyState2: State{

							},
						},
					},
					result:  MyState2,
					events: []EventType{MoveToState2},
					context: &DataContext{
						Data: map[string]string{
							"hello":   "world",
						},

					},
		},

	}

	for _, item := range dataItems {
		for _, event := range item.events {
			item.input.SendEvent(event, item.context)
		}
		if item.input.Current != item.result{
			t.Error("Wrong end state: ", item.input.Current)
		}
	}



}

