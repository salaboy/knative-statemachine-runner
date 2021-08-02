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
	"github.com/salaboy/knative-workflow-runner/pkg/workflow"
	"github.com/salaboy/knative-workflow-runner/pkg/workflow/workflowextension"
)

// This is going to keep in memory the statemachines that this runner is handling.
//   Hence this should be persisted in a key-value store (example Redis)
var workflows map[string]*workflow.StateMachine

// This should identify the runner, so if we persist state in a storage we split it by runnners
//  - This id should probably related to worklow metadata
var runnerId string

var SINK = os.Getenv("EVENT_SINK")
var RUNNER_ID = os.Getenv("RUNNER_ID")
var WORKFLOW_NAME = os.Getenv("WORKFLOW_NAME")
var WORKFLOW_VERSION = os.Getenv("WORKFLOW_VERSION")
var WORKFLOW_DEF = os.Getenv("WORKFLOW_DEF")
var WORKFLOW_DEF_PATH = os.Getenv("WORKFLOW_DEF_PATH") // Full path for the Workflow Definition YAML file

var REDIS_HOST = os.Getenv("REDIS_HOST")

var workflowDefinition = workflow.Workflow{}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/info", RunnerInfoHandler).Methods("GET")
	r.HandleFunc("/definition", RunnerWorkflowDefintionHandler).Methods("GET")
	r.HandleFunc("/workflows", WorkflowsNewHandler).Methods("POST")
	r.HandleFunc("/workflows/events", WorkflowEventsHandler).Methods("POST")
	r.HandleFunc("/workflows", WorkflowsGETHandler).Methods("GET")
	r.HandleFunc("/workflows/{id}", WorkflowByIdGETHandler).Methods("GET")
	log.Printf("Workflow Runner Started in port 8080!")
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func initWorkflow(){

	states, err := workflow.ReadStatesFromEnvString(WORKFLOW_DEF)
	if err != nil {
		fmt.Println(err)
		return
	}

	workflowDefinition.Name = WORKFLOW_NAME
	workflowDefinition.Version = WORKFLOW_VERSION
	workflowDefinition.States = states

	initRedis()

}


func initRedis(){
	// Connect, Do I need to store the definition?
	workflows = make(map[string]*workflow.StateMachine)

}

func RunnerInfoHandler(writer http.ResponseWriter, request *http.Request) {
	// return workflow definition name + version
	// number of instances.. and Id for the runner

	var runnerInfo = workflow.RunnerInfo{
		Id:      RUNNER_ID,
		Name:    WORKFLOW_NAME,
		Version: WORKFLOW_VERSION,
	}

	respondWithJSON(writer, http.StatusOK, &runnerInfo)
}

func RunnerWorkflowDefintionHandler(writer http.ResponseWriter, request *http.Request) {
	// return workflow definition name + version
	// number of instances.. and Id for the runner
	respondWithJSON(writer, http.StatusOK, &workflowDefinition)
}

func WorkflowEventsHandler(writer http.ResponseWriter, request *http.Request) {
	workflowExtension := workflowextension.WorkflowExtension{}

	ctx := context.Background()
	message := cehttp.NewMessageFromHttpRequest(request)
	event, _ := binding.ToEvent(ctx, message, workflowExtension.ReadTransformer(), workflowExtension.WriteTransformer())

	fmt.Printf("Got an Event: %s", event)
	eventContext := workflow.EventContext{}
	err := json.Unmarshal(event.Data(), &eventContext)
	if err != nil {
		fmt.Println(err)
		return
	}

	event.ExtensionAs(workflowextension.WorkflowIdCloudEventsExtension, workflowExtension)

	workflowRun := workflows[workflowExtension.WorkflowId]

	workflowRun.SendEvent(workflow.EventType(event.Type()), eventContext)

}

func WorkflowsNewHandler(writer http.ResponseWriter, request *http.Request) {
	// Create a new instance using the States from the Definition
	var stateMachine = workflow.StateMachine{}
	stateMachine.States = workflowDefinition.States
	stateMachine.SINK = SINK
	// Creating Instance Id
	id, _ := uuid.NewUUID()
	stateMachine.Id = id.String()

	log.Printf("New Workflow Instance: %s", stateMachine.Id)
	log.Printf("Workflow Instance SINK Set to: %s", stateMachine.SINK)

	storeInstance(stateMachine.Id , &stateMachine)

	respondWithJSON(writer, http.StatusOK, &stateMachine)

}

func storeInstance(stateMachineId string , stateMachine *workflow.StateMachine){
	workflows[stateMachineId] = stateMachine
}

func WorkflowsGETHandler(writer http.ResponseWriter, request *http.Request) {
	respondWithJSON(writer, http.StatusOK, &workflows)
}

func WorkflowByIdGETHandler(writer http.ResponseWriter, request *http.Request) {
	id := mux.Vars(request)["id"]
	workflowRun := workflows[id]
	respondWithJSON(writer, http.StatusOK, &workflowRun)
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
