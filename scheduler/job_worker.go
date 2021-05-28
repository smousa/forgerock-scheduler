package scheduler

import (
	"context"
	"time"

	"github.com/smousa/forgerock-scheduler/api"
	"golang.org/x/sync/errgroup"
)

// JobWorker executes an incoming job
type JobWorker interface {
	Consume(ctx context.Context, job *api.JobRequest)
}

type jobWorker struct {
	store Store
	queue TaskProducer
}

// NewJobWorker instantiates a new job worker
func NewJobWorker(store Store, queue TaskProducer) JobWorker {
	return &jobWorker{store, queue}
}

// Consume writes the tasks to the queue and awaits the result
func (w *jobWorker) Consume(ctx context.Context, job *api.JobRequest) {
	g, ctx := errgroup.WithContext(ctx)

	for _, task := range job.Tasks {
		task := task
		g.Go(func() error {
			w.store.UpdateStatus(ctx, job.Id, &Status{
				TaskName:  task.Name,
				State:     PendingState,
				Timestamp: time.Now(),
			})
			err := w.queue.Publish(ctx, &api.TaskRequest{
				JobId: job.Id,
				Task:  task,
			})
			if err != nil {
				return err
			}
			_, err = w.store.WaitForTaskToComplete(ctx, job.Id, task.Name)
			return err
		})
	}

	err := g.Wait()
	if err != nil {
		ContextLogger(ctx).WithError(err).WithField("job_id", job.Id).Error("Could not execute job.")
	}
}
