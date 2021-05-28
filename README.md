# forgerock-scheduler
This is an implementation of a work scheduler system using the power of [Firestore](https://firebase.google.com/docs/firestore) and [Redis Queues](https://github.com/adjust/rmq).

## Installation

### Prerequisites
* go1.16.4
* docker 17.10.0-ce-rc1
* docker-compose 1.29.2

And then just `docker-compose up`

## Usage
The server communicates using grpc, so you will need to install [grpcurl](https://github.com/fullstorydev/grpcurl#from-source).

```
go get github.com/fullstorydev/grpcurl/...
go install github.com/fullstorydev/grpcurl/cmd/grpcurl
```
### api.JobService/CreateJob
Inserts a job into the queue for execution. Returns the job id.
```json
{
  "tasks": [
	  {
	    "name": "my very complicated task name this field is required."
	    "sleep": {
	      "duration_seconds": 1
	    }
	  }
  ]
}
```
Here is a full sample request but with 2 tasks specified
```
grpcurl -plaintext -d '{"tasks": [{"name": "sleep 1", "sleep": {"duration_seconds": 1}}, {"name": "sleep 2", "sleep": {"duration_seconds":2}}]}' localhost:8888 api.JobService/CreateJob
```
* Just keep in mind that there is no field validation, so try not to make any mistakes
* Logging is available for monitoring the state of the job
* Job status is also stored in Firestore, with the assumption that it could be used to report back to the user

## Design
This implementation is divided into three pieces:
1. Job Manager: Accepts external requests from the user and writes to the Job Queue
2. Job Worker: Reads from the Job Queue and then writes individual tasks into the Task Queue
3. Task Manager: Reads from the Task Queue and then executes Tasks

As each task is processed, state updates are written to Firestore with values such as:
* Pending: Task has been enqueued
* Running: Task is being executed
* Success: Task completed successfully
* Failed: Task completed unsuccessfully

Having this information will help support other features like sequential tasks and user feedback.

## Testing
Due to time limitations, I only wrote a few integration tests, just to demonstrate my style of testing.  [Ginkgo](github.com/onsi/ginkgo) is my preferred testing framework of choice, because I really like the BDD style.

To run tests, first you have to start up redis and firestore
`docker-compose run -p 6379:6379 redis` 
`docker-compose run -p 8200:8200 firestore`
And then you can start the tests
```bash
# from project root
cd scheduler
REDIS_ADDRESS=localhost:6379 FIRESTORE_EMULATOR_HOST=localhost:8200 FIRESTORE_PROJECT_ID=development go test
```
