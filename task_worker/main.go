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
	"github.com/spf13/viper"
)

const (
	// ServiceName is the name of the service
	ServiceName = "TaskWorker"
	// ServicePort is the port the service is hosted on
	ServicePort = ":8080"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// set up logger
	log := logrus.StandardLogger()
	level, err := logrus.ParseLevel(viper.GetString("LOG_LEVEL"))
	if err != nil {
		log.WithError(err).Fatal("Could not retrieve log level")
	}
	log.Level = level
	entry := logrus.NewEntry(log)
	ctx = scheduler.WithLogEntry(ctx, entry)

	// connect to firestore
	projectID := viper.GetString("PROJECT_ID")
	firestoreClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.WithError(err).WithField("project_id", projectID).Fatal("Could not connect to firestore")
	}

	// connect to redis
	redisAddress := viper.GetString("REDIS_ADDRESS")
	if redisAddress == "" {
		log.Fatal("Redis address is required")
	}
	// TODO: handle channel errors?
	redisConn, err := rmq.OpenConnection(ServiceName, "tcp", redisAddress, nil)
	if err != nil {
		log.WithError(err).
			WithField("address", redisAddress).
			WithField("address", redisAddress).
			Fatal("Could not connect to redis")
	}
	taskQueue, err := redisConn.OpenQueue(scheduler.TaskQueueName)
	if err != nil {
		log.WithError(err).WithField("queue", scheduler.TaskQueueName).Fatal("Could not open queue")
	}

	// initialize the store
	store := scheduler.NewFirestoreStore(firestoreClient)

	// initialize the jobConsumer for consuming jobs
	taskConsumer := scheduler.NewRedisJobQueue(taskQueue, scheduler.WithTaskTag(ServiceName))

	// initialize the task executor
	taskExec := scheduler.NewTaskExecutor()

	// add the consumer and start consuming jobs
	worker := scheduler.NewTaskWorker(store, taskExec)
	err = taskConsumer.Consume(ctx, worker)
	if err != nil {
		log.WithError(err).WithField("queue", scheduler.TaskQueueName).Fatal("Could not add consumer")
	}
	err = jobQueue.StartConsuming(10, time.Second)
	if err != nil {
		log.WithError(err).WithField("queue", scheduler.TaskQueueName).Fatal("Could not start consumer")
	}

	// wait for shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)
	<-sigChan
	cancel()
	<-jobQueue.StopConsuming()
}