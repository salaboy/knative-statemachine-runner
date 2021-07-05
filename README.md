# Knative Workflow Runner

This project runs a workflow definition and keeps track of its state. 

For more information look at the [Knative Workflow Controller](http://github.com/salaboy/knative-workflow) project which creates new Workflow Runners.

# APIs

Each Runner instance expose the following REST endpoints: 
- POST `/workflows/` Creates a new instance of a workflow with a unique id
- POST `/workflows/events` Consume Cloud Events 
- GET `/workflows/` Get all avaialble Workflow instances

# Running the project from Source
