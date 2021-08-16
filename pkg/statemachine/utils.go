package statemachine

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)


func ReadStatesFromYAML(stateMachineFilePath string) (States, error){
	statesDefinition := States{}
	content, err := ioutil.ReadFile(stateMachineFilePath)
	if err != nil {
		log.Fatal(err)
	}

	stateMachineYAML := string(content)


	err = yaml.Unmarshal([]byte(stateMachineYAML), &statesDefinition)
	if err != nil {
		return States{}, err
	}
	log.Printf(">> StatesDefinition :\n%v\n\n", statesDefinition)

	return statesDefinition, nil

}

func ReadStatesFromEnvString(statesContent string) (States, error){
	statesDefinition := States{}
	err := yaml.Unmarshal([]byte(statesContent), &statesDefinition)
	if err != nil {
		return States{}, err
	}

	log.Printf(">> StatesDefinition :\n%v\n\n", statesDefinition)

	return statesDefinition, nil
}

