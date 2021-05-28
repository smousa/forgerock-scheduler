package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/adjust/rmq/v4"
	"github.com/sirupsen/logrus"
	"github.com/smousa/forgerock-scheduler/scheduler"
)

const (
	// ServiceName is the name of the service
	ServiceName = "JobWorker"
	// ServicePort is the port the service is hosted on
	ServicePort = ":8080"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// set up logger
	log := logrus.StandardLogger()
	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		log.WithError(err).Fatal("Could not retrieve log level")
	}
	log.Level = level
	entry := logrus.NewEntry(log)
	ctx = scheduler.WithLogger(ctx, entry)

	// connect to firestore
	log.Debug("Connecting to firestore")
	projectID := os.Getenv("PROJECT_ID")
	firestoreClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.WithError(err).WithField("project_id", projectID).Fatal("Could not connect to firestore")
	}

	// connect to redis
	log.Debug("Connecting to redis")
	redisAddress := os.Getenv("REDIS_ADDRESS")
	if redisAddress == "" {
		log.Fatal("Redis address is required")
	}
	// TODO: handle channel errors?
	redisConn, err := rmq.OpenConnection(ServiceName, "tcp", redisAddress, 0, nil)
	if err != nil {
		log.WithError(err).
			WithField("address", redisAddress).
			Fatal("Could not connect to redis")
	}
	jobQueue, err := redisConn.OpenQueue(scheduler.JobQueueName)
	if err != nil {
		log.WithError(err).WithField("queue", scheduler.JobQueueName).Fatal("Could not open queue")
	}
	taskQueue, err := redisConn.OpenQueue(scheduler.TaskQueueName)
	if err != nil {
		log.WithError(err).WithField("queue", scheduler.TaskQueueName).Fatal("Could not open queue")
	}

	// initialize the store
	store := scheduler.NewFirestoreStore(firestoreClient)

	// initialize the jobConsumer for consuming jobs
	jobConsumer := scheduler.NewRedisJobQueue(jobQueue, scheduler.WithJobTag(ServiceName))

	// initialize the taskProducer for producing tasks
	taskProducer := scheduler.NewRedisTaskQueue(taskQueue)

	// add the consumer and start consuming jobs
	log.WithField("queue", scheduler.JobQueueName).Debug("Starting consumer")
	err = jobQueue.StartConsuming(10, time.Second)
	if err != nil {
		log.WithError(err).WithField("queue", scheduler.JobQueueName).Fatal("Could not start consumer")
	}
	worker := scheduler.NewJobWorker(store, taskProducer)
	err = jobConsumer.Consume(ctx, worker)
	if err != nil {
		log.WithError(err).WithField("queue", scheduler.JobQueueName).Fatal("Could not add consumer")
	}
	log.Info("Ready!")

	// wait for shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)
	<-sigChan
	cancel()
	<-jobQueue.StopConsuming()
}
