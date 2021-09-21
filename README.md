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


This example create a new State Machine instance by sending a POST request to the following endpoint
```
curl -X POST http://localhost:8080/statemachines
```
Once you have the Id of the State Machine Instance, then you can send events to that specific instance. 

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

Another option that might work better if you don't want to add an extra request to track the state of an existing flow is to use a `Correlation Key`. 

In this case you don't need to create the instance, but you need to specify a correlation key that will be used for all the future events against that instance. 

```
curl -X POST -H "Content-Type: application/json" \
  -H "ce-specversion: 1.0" \
  -H "ce-source: curl-command" \
  -H "ce-type: JoinedQueue" \
  -H "ce-id: 123-abc" \
  -H "ce-correlationkey: tickets-123" \
  -d '{"name":"Salaboy"}' \
  http://localhost:8080/statemachines/events
```

Now, `tickets-123` will be used as a correlation key to correlate future events. 

