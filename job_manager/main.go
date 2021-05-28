package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/firestore"
	"github.com/adjust/rmq/v4"
	"github.com/sirupsen/logrus"
	"github.com/smousa/forgerock-scheduler/api"
	"github.com/smousa/forgerock-scheduler/scheduler"
	"google.golang.org/grpc"
)

const (
	// ServiceName is the name of the service
	ServiceName = "JobManager"
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
	projectID := os.Getenv("FIRESTORE_PROJECT_ID")
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

	// initialize the store
	store := scheduler.NewFirestoreStore(firestoreClient)

	// initialize the queue for publishing
	queue := scheduler.NewRedisJobQueue(jobQueue)

	// set up the server for listening
	log.Debug("Starting server")
	lis, err := net.Listen("tcp", ServicePort)
	if err != nil {
		log.WithError(err).WithField("port", ServicePort).Fatal("Failed to listen")
	}
	server := grpc.NewServer()
	service := scheduler.NewService(store, queue)
	api.RegisterJobServiceServer(server, service)
	go server.Serve(lis)
	log.Info("Ready!")

	// wait for shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)
	<-sigChan
	cancel()
	server.GracefulStop()
}
