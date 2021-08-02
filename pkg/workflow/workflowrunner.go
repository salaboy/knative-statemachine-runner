package workflow

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)


func ReadStatesFromYAML(workflowFilePath string) (States, error){
	statesDefinition := States{}
	content, err := ioutil.ReadFile(workflowFilePath)
	if err != nil {
		log.Fatal(err)
	}

	workflowYAML := string(content)

	err = yaml.Unmarshal([]byte(workflowYAML), &statesDefinition)
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
	return statesDefinition, nil
}

