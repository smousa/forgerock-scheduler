package scheduler

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"golang.org/x/sync/semaphore"
	"google.golang.org/api/iterator"
)

const (
	// PendingState represents a task that is enqueued and waiting for
	// processing
	PendingState = "pending"

	// RunningState represents a task that is currently in progress
	RunningState = "running"

	// SuccessState represents a task that has completed successfully
	SuccessState = "success"

	// FailedState represents a task that has failed
	FailedState = "failed"

	// SleepAction represents the sleep action
	SleepAction = "sleep"
)

// Job contains information about a Job
type Job struct {
	Tasks   []*Task   `firestore:"tasks"`
	Created time.Time `firestore:"created"`
}

// Task contains information about how to perform a task
type Task struct {
	Name   string            `firestore:"name"`
	Action string            `firestore:"action"`
	Params map[string]string `firestore:"params"`
}

// Status contains information relating to the state of a job
type Status struct {
	TaskName  string    `firestore:"task_name"`
	State     string    `firestore:"state"`
	Message   string    `firestore:"message"`
	Timestamp time.Time `firestore:"timestamp"`
}

// Store manages the progress of a given task
type Store interface {
	// Add creates a new job record and returns the job id
	Add(context.Context, *Job) (string, error)
	// UpdateStatus updates the status for a particular job
	UpdateStatus(ctx context.Context, jobID string, status *Status)
	// WaitForTaskToComplete returns the status when a task is completed
	WaitForTaskToComplete(ctx context.Context, jobID, taskName string) (*Status, error)
}

// FirestoreStoreOption describes options that can be applied to the firestore
// store
type FirestoreStoreOption func(store *FirestoreStore)

// WithMaxListeners sets the max number of snapshot listeners
func WithMaxListeners(n int64) FirestoreStoreOption {
	return func(store *FirestoreStore) {
		store.sem = semaphore.NewWeighted(n)
	}
}

// FirestoreStore implements Store
type FirestoreStore struct {
	sem    *semaphore.Weighted
	client *firestore.Client
}

// NewFirestoreStore creates a new firestore store
func NewFirestoreStore(client *firestore.Client, options ...FirestoreStoreOption) *FirestoreStore {
	s := &FirestoreStore{
		client: client,
		sem:    semaphore.NewWeighted(100),
	}
	for _, f := range options {
		f(s)
	}
	return s
}

// Add creates a new job document
func (s *FirestoreStore) Add(ctx context.Context, job *Job) (string, error) {
	doc, _, err := s.client.Collection("jobs").Add(ctx, job)
	if err != nil {
		return "", err
	}
	return doc.ID, nil
}

// UpdateStatus creates a new status document for a particular job
func (s *FirestoreStore) UpdateStatus(ctx context.Context, jobID string, status *Status) {
	log := ContextLogger(ctx).WithField("job_id", jobID)

	_, _, err := s.client.
		Collection("jobs").
		Doc(jobID).
		Collection("status").
		Add(ctx, status)
	if err != nil {
		log.WithError(err).Fatal("Could not update job status")
	}
	log.WithField("status", status).Debug("Update job status")
}

// WaitForTaskToComplete waits for a status document of a particular task to
// complete.
func (s *FirestoreStore) WaitForTaskToComplete(ctx context.Context, jobID, taskName string) (*Status, error) {
	if err := s.sem.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	defer s.sem.Release(1)

	iter := s.client.
		Collection("jobs").Doc(jobID).Collection("status").
		Where("task_name", "==", taskName).
		Where("state", "in", []string{SuccessState, FailedState}).
		Snapshots(ctx)

	defer iter.Stop()

	for {
		snap, err := iter.Next()
		if err != nil {
			return nil, err
		}

		doc, err := snap.Documents.Next()
		if err == iterator.Done {
			continue
		}
		if err != nil {
			return nil, err
		}
		snap.Documents.Stop()
		status := new(Status)
		if err := doc.DataTo(status); err != nil {
			return nil, err
		}

		return status, nil
	}
}
