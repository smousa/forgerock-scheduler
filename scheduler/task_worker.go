package scheduler

import (
	"context"
	"time"

	"github.com/smousa/forgerock-scheduler/api"
)

// TaskWorker executes an incoming task
type TaskWorker interface {
	Consume(ctx context.Context, task *api.TaskRequest)
}

// TaskExecutor executes the various tasks
type TaskExecutor interface {
	Sleep(ctx context.Context, sleep *api.SleepAction) error
}

type taskWorker struct {
	store Store
	exec  TaskExecutor
}

// NewTaskWorker instantiates a new task worker
func NewTaskWorker(store Store, exec TaskExecutor) TaskWorker {
	return &taskWorker{store, exec}
}

// Consume performs the prescribed tasks and awaits the result
func (w *taskWorker) Consume(ctx context.Context, task *api.TaskRequest) {
	w.store.UpdateStatus(ctx, task.JobId, &Status{
		TaskName:  task.Task.Name,
		State:     RunningState,
		Timestamp: time.Now(),
	})

	switch a := task.Task.GetAction().(type) {
	case *api.Task_Sleep:
		err := w.exec.Sleep(ctx, a.Sleep)
		if err != nil {
			w.store.UpdateStatus(ctx, task.JobId, &Status{
				TaskName:  task.Task.Name,
				State:     FailedState,
				Message:   err.Error(),
				Timestamp: time.Now(),
			})
		} else {
			w.store.UpdateStatus(ctx, task.JobId, &Status{
				TaskName:  task.Task.Name,
				State:     SuccessState,
				Timestamp: time.Now(),
			})
		}
	default:
		// TODO: Add fatal log message
		// OK, this might happen if your server isn't running the latest version
		// which means you may just want to FATAL out.
	}
}

type taskExecutor struct{}

// NewTaskExecutor creates a new task executor
func NewTaskExecutor() TaskExecutor {
	return &taskExecutor{}
}

// Sleep returns when the duration has passed
func (exec *taskExecutor) Sleep(ctx context.Context, action *api.SleepAction) error {
	t := time.NewTimer(time.Duration(action.DurationSeconds) * time.Second)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case err := <-ctx.Done():
		return err
	}
}
