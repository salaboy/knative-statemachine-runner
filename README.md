# Knative Workflow Runner

This project runs a workflow definition and keeps track of its state. 

For more information look at the [Knative Workflow Controller](http://github.com/salaboy/knative-workflow) project which creates new Workflow Runners.

# APIs

Each Runner instance expose the following REST endpoints: 
- POST `/workflows/` Creates a new instance of a workflow with a unique id
- POST `/workflows/events` Consume Cloud Events 
- GET `/workflows/` Get all available Workflow instances
- GET `/workflows/{id}` Get workflow instance by id

# Running the project from Source

```
go build
```

```
go run main.go
```

Publish to local registry, it uses `KO_DOCKER_REGISTRY` var to decide where to publish

```
ko publish .
```


# Example
```
curl -X POST http://localhost:8080/workflows
```


```
curl -X POST -H "Content-Type: application/json" \
  -H "ce-specversion: 1.0" \
  -H "ce-source: curl-command" \
  -H "ce-type: JoinedQueue" \
  -H "ce-id: 123-abc" \
  -H "ce-workflowid: 194c70ae-edfa-11eb-ae55-367ddaa504e1" \
  -d '{"name":"Salaboy"}' \
  http://localhost:8080/workflows/events
```