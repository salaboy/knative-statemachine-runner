package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
//   Hence this should be persisted in a key-value store
var workflows map[string]*workflow.StateMachine

var SINK = os.Getenv("EVENT_SINK")

var WORKFLOW = os.Getenv("WORKFLOW")

func main() {

	workflows = make(map[string]*workflow.StateMachine)

	r := mux.NewRouter()
	r.HandleFunc("/workflows/new", WorkflowsNewHandler).Methods("POST")
	r.HandleFunc("/workflows/", WorkflowEventsHandler).Methods("POST")
	r.HandleFunc("/workflows", WorkflowsGETHandler).Methods("GET")
	log.Printf("Workflow Runner Started in port 8080!")
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))

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

	smb := workflow.StateMachineBuilder{}
	wf := workflow.StateMachine{}

	if WORKFLOW != "" {
		log.Printf("Reading Workflow from ENV: %s", WORKFLOW)
		wf, _ = smb.ReadFromENVString(WORKFLOW)
	} else {
		log.Printf("Reading Workflow from File: %s", "workflow-buy-tickets.yaml")
		wf, _ = smb.ReadFromYAML("workflow-buy-tickets.yaml")

	}
	id, _ := uuid.NewUUID()
	wf.Id = id.String()

	log.Printf("New Workflow: %s", wf.Id)
	workflows[wf.Id] = &wf

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	io.WriteString(writer, fmt.Sprintf("%+v", wf))

}

func WorkflowsGETHandler(writer http.ResponseWriter, request *http.Request) {

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	for _, workflow := range workflows {
		io.WriteString(writer, fmt.Sprintf("%s", workflow))
	}

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
