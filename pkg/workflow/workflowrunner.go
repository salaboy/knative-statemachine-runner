package workflow

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type StateMachineBuilder struct{

}

func (st *StateMachineBuilder) ReadFromYAML(workflowFilePath string) (StateMachine, error) {
    stateMachine := StateMachine{}
	content, err := ioutil.ReadFile(workflowFilePath)
	if err != nil {
		log.Fatal(err)
	}

	// Convert []byte to string and print to screen
	workflowYAML := string(content)

	err = yaml.Unmarshal([]byte(workflowYAML), &stateMachine)
	if err != nil {
		return StateMachine{}, err
	}
	fmt.Printf(">> StateMachine :\n%v\n\n", stateMachine)

	return stateMachine, nil
}

func (st *StateMachineBuilder) ReadFromENVString(workflowContent string) (StateMachine, error) {
	stateMachine := StateMachine{}

	err := yaml.Unmarshal([]byte(workflowContent), &stateMachine)
	if err != nil {
		return StateMachine{}, err
	}
	fmt.Printf(">> StateMachine :\n%v\n\n", stateMachine)

	return stateMachine, nil
}

