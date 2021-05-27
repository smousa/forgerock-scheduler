package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/smousa/forgerock-scheduler/api"
	grpc_codes "google.golang.org/grpc/codes"
	grpc_status "google.golang.org/grpc/status"
)

type service struct {
	store Store
	queue JobProducer
}

// NewService creates a new job service
func NewService(store Store, queue JobProducer) api.JobServiceServer {
	return &service{
		store: store,
		queue: queue,
	}
}

// CreateJob adds a job to the job queue and adds a status record
func (s *service) CreateJob(ctx context.Context, req *api.CreateJobRequest) (*api.CreateJobResponse, error) {
	log := ContextLogger(ctx)

	tasks := make([]*Task, len(req.Tasks))
	for i, task := range req.Tasks {
		switch a := task.GetAction().(type) {
		case *api.Task_Sleep:
			tasks[i] = &Task{
				Name:   task.Name,
				Action: SleepAction,
				Params: map[string]string{
					"duration_seconds": fmt.Sprintf("%d", a.Sleep.DurationSeconds),
				},
			}
		default:
			log.Fatal("incompatible action type")
		}
	}

	// Create the job record
	jobID, err := s.store.Add(ctx, &Job{
		Tasks:   tasks,
		Created: time.Now(),
	})
	if err != nil {
		return nil, grpc_status.Error(
			grpc_codes.Internal,
			errors.Wrap(err, "could not create job record").Error(),
		)
	}

	// Add the job record to the queue
	err = s.queue.Publish(ctx, &api.JobRequest{Id: jobID, Tasks: req.Tasks})
	if err != nil {
		s.store.UpdateStatus(ctx, jobID, &Status{
			State:     FailedState,
			Message:   err.Error(),
			Timestamp: time.Now(),
		})
		return nil, grpc_status.Error(
			grpc_codes.Internal,
			errors.Wrap(err, "could not enqueue job record").Error(),
		)
	}
	return &api.CreateJobResponse{Id: jobID}, nil
}
