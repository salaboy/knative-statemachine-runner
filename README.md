# KNative StateMachine Runner

This project runs a StateMachine definition and keeps track of its state. 

For more information look at the [Knative StateMachine Controller](http://github.com/salaboy/knative-state) project which creates new StateMachine Runners.

# APIs

Each Runner instance expose the following REST endpoints: 
- POST `/statemachines/` Creates a new **StateMachineInstance**, returns a unique id
- POST `/statemachines/events` Consume Cloud Events for a given **StateMachineInstance**
- GET `/statemachines/` Get all available **StateMachineInstances**
- GET `/statemachines/{id}` Get **StateMachineInstance** by id

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
curl -X POST http://localhost:8080/statemachines
```


```
curl -X POST -H "Content-Type: application/json" \
  -H "ce-specversion: 1.0" \
  -H "ce-source: curl-command" \
  -H "ce-type: JoinedQueue" \
  -H "ce-id: 123-abc" \
  -H "ce-statemachineid: 7e3e2258-fe78-11eb-a85e-acde48001122" \
  -d '{"name":"Salaboy"}' \
  http://localhost:8080/statemachines/events
```