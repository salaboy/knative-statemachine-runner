package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cloudevents/sdk-go/v2/binding"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/salaboy/knative-statemachine-runner/pkg/statemachine"
	"github.com/salaboy/knative-statemachine-runner/pkg/statemachine/extension"
)

// This is going to keep in memory the statemachines that this runner is handling.
//   Hence this should be persisted in a key-value store (example Redis)
var stateMachines map[string]*statemachine.StateMachine

// This is going to keep in memory the statemachines that this runner is handling.
// This map index statemachines by CorrelationKey
//   Hence this should be persisted in a key-value store (example Redis)
var stateMachinesCorrelationKey map[string]*statemachine.StateMachine

// This should identify the runner, so if we persist state in a storage we split it by runnners
//  - This id should probably related to worklow metadata
var runnerId string

var SINK = os.Getenv("EVENT_SINK")
var RUNNER_ID = os.Getenv("RUNNER_ID")
var STATEMACHINE_NAME = os.Getenv("STATEMACHINE_NAME")
var STATEMACHINE_VERSION = os.Getenv("STATEMACHINE_VERSION")
var STATEMACHINE_DEF = os.Getenv("STATEMACHINE_DEF")
var STATEMACHINE_DEF_PATH = os.Getenv("STATEMACHINE_DEF_PATH") // Full path for the StateMachine Definition YAML file

// For example purposes
var REDIS_HOST = os.Getenv("REDIS_HOST")

var stateMachineDefinition = statemachine.StateMachineDefinition{}

func main() {

	initStateMachine()

	r := mux.NewRouter()
	r.HandleFunc("/info", RunnerInfoHandler).Methods("GET")
	r.HandleFunc("/definition", RunnerStateMachineDefintionHandler).Methods("GET")
	r.HandleFunc("/statemachines", StateMachinesNewHandler).Methods("POST")
	r.HandleFunc("/statemachines/events", StateMachineEventsHandler).Methods("POST")
	r.HandleFunc("/statemachines", StateMachinesGETHandler).Methods("GET")
	r.HandleFunc("/statemachines/{id}", StateMachineByIdGETHandler).Methods("GET")
	log.Printf("StateMachine Runner Started in port 8080!")
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func initStateMachine() {
	if STATEMACHINE_DEF != "" {
		states, err := statemachine.ReadStatesFromEnvString(STATEMACHINE_DEF)
		if err != nil {
			fmt.Println(err)
			return
		}
		stateMachineDefinition.Name = STATEMACHINE_NAME
		stateMachineDefinition.Version = STATEMACHINE_VERSION
		stateMachineDefinition.States = states
		log.Printf("StateMachine loaded from STATEMACHINE_DEF env var \n%s", states)
	} else {
		//Load demo statemachine: statemachine-buy-tickets.yaml
		states, err := statemachine.ReadStatesFromYAML("statemachine-buy-tickets.yaml")
		if err != nil {
			fmt.Println(err)
			return
		}
		stateMachineDefinition.Name = "buy-tickets"
		stateMachineDefinition.Version = "1.0.0"
		stateMachineDefinition.States = states
		log.Printf("StateMachine loaded from path: statemachine-buy-tickets.yaml  \n%s", states)
	}

	initRedis()

}

func initRedis() {

	// StateMachines by Id
	stateMachines = make(map[string]*statemachine.StateMachine)
	// StateMachine by CorrelationKey
	stateMachinesCorrelationKey = make(map[string]*statemachine.StateMachine)

	// Connect, Do I need to store the definition?

	//client := redis.NewClient(&redis.Options{
	//	Addr: "localhost:6379",
	//	Password: "",
	//	DB: 0,
	//})
	//if err := client.Ping().Err(); err != nil {
	//
	//}

}

func RunnerInfoHandler(writer http.ResponseWriter, request *http.Request) {
	// return statemachine definition name + version
	// number of instances.. and Id for the runner

	var runnerInfo = statemachine.RunnerInfo{
		Id:      RUNNER_ID,
		Name:    STATEMACHINE_NAME,
		Version: STATEMACHINE_VERSION,
	}

	respondWithJSON(writer, http.StatusOK, &runnerInfo)
}

func RunnerStateMachineDefintionHandler(writer http.ResponseWriter, request *http.Request) {
	// return statemachine definition name + version
	// number of instances.. and Id for the runner
	respondWithJSON(writer, http.StatusOK, &stateMachineDefinition)
}

func StateMachineEventsHandler(writer http.ResponseWriter, request *http.Request) {
	stateMachineExtension := extension.StateMachineExtension{}

	ctx := context.Background()
	message := cehttp.NewMessageFromHttpRequest(request)
	event, _ := binding.ToEvent(ctx, message, stateMachineExtension.ReadTransformer(), stateMachineExtension.WriteTransformer())

	fmt.Printf("Got an Event: %s", event)
	eventContext := statemachine.EventContext{}
	err := json.Unmarshal(event.Data(), &eventContext)
	if err != nil {
		fmt.Println(err)
		return
	}

	event.ExtensionAs(extension.StateMachineIdCloudEventsExtension, stateMachineExtension)
	event.ExtensionAs(extension.CorrelationKeyCloudEventsExtension, stateMachineExtension)


	stateMachineInstance := loadStateMachineInstanceById(stateMachineExtension.StateMachineId)
	if stateMachineInstance != nil && stateMachineInstance.Id != "" {
		stateMachineInstance.SendEvent(statemachine.EventType(event.Type()), eventContext)
	} else {
		fmt.Printf("StateMachineInstance %s not found, checking for correlationkey: %s \n", stateMachineExtension.StateMachineId, stateMachineExtension.CorrelationKey)

		if stateMachineExtension.CorrelationKey != "" {
			// Try loading from correlation key index
			stateMachineInstance = loadStateMachineInstanceByCorrelationKey(stateMachineExtension.CorrelationKey)

			if stateMachineInstance != nil && stateMachineInstance.Id != ""{
				stateMachineInstance.SendEvent(statemachine.EventType(event.Type()), eventContext)
			} else{
				// if it doesn't exist let's create one with the provided correlation key
				stateMachineInstance = newInstance(stateMachineExtension.CorrelationKey)

				log.Printf("New StateMachineInstance: %s", stateMachineInstance.Id)
				log.Printf("StateMachineInstance SINK Set to: %s", stateMachineInstance.SINK)

				storeInstance(stateMachineInstance.Id, stateMachineInstance.CorrelationKey,  stateMachineInstance)

				stateMachineInstance.SendEvent(statemachine.EventType(event.Type()), eventContext)
			}

		} else{
			fmt.Printf("No correlation key provided, no instance available for event %s \n", event.Type())
		}
	}


}

func loadStateMachineInstanceById(id string) *statemachine.StateMachine {
	return stateMachines[id]
}

func loadStateMachineInstanceByCorrelationKey(correlationKey string) *statemachine.StateMachine {
	return stateMachinesCorrelationKey[correlationKey]
}

func StateMachinesNewHandler(writer http.ResponseWriter, request *http.Request) {
	// Create a new instance using the States from the Definition
	var stateMachine = newInstance("")

	log.Printf("New StateMachineInstance: %s", stateMachine.Id)
	log.Printf("StateMachineInstance SINK Set to: %s", stateMachine.SINK)

	storeInstance(stateMachine.Id, stateMachine.CorrelationKey, stateMachine)

	respondWithJSON(writer, http.StatusOK, &stateMachine)

}

func newInstance(correlationKey string) *statemachine.StateMachine {
	var stateMachine = statemachine.StateMachine{}
	stateMachine.States = stateMachineDefinition.States
	stateMachine.SINK = SINK
	// Creating Instance Id
	uuid, _ := uuid.NewUUID()
	stateMachine.Id = uuid.String()
	stateMachine.CorrelationKey = correlationKey

	return &stateMachine
}

func storeInstance(stateMachineId string, stateMachineCorrelationKey string, stateMachine *statemachine.StateMachine) {
	fmt.Printf("Trying to store instance with id: %s and correlationkey: %s \n", stateMachineId, stateMachineCorrelationKey)
	if stateMachineId != "" {
		stateMachines[stateMachineId] = stateMachine
	}
	if stateMachineCorrelationKey != "" {
		stateMachinesCorrelationKey[stateMachineCorrelationKey] = stateMachine
	}
}

func StateMachinesGETHandler(writer http.ResponseWriter, request *http.Request) {
	respondWithJSON(writer, http.StatusOK, &stateMachines)
}

func StateMachineByIdGETHandler(writer http.ResponseWriter, request *http.Request) {
	id := mux.Vars(request)["id"]
	stateMachineInstance := stateMachines[id]
	respondWithJSON(writer, http.StatusOK, &stateMachineInstance)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
