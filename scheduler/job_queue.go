package scheduler

import (
	"context"

	"github.com/adjust/rmq/v4"
	"github.com/pkg/errors"
	"github.com/smousa/forgerock-scheduler/api"
	"google.golang.org/protobuf/proto"
)

// JobQueueName is the name of the job queue
const JobQueueName = "jobs"

// JobProducer inserts jobs into the job queue
type JobProducer interface {
	Publish(ctx context.Context, job *api.JobRequest) error
}

// JobConsumer reads jobs from the job queue
type JobConsumer interface {
	Consume(ctx context.Context, w JobWorker) error
}

// RedisJobQueue represents a redis queue for a particular job
type RedisJobQueue struct {
	q   rmq.Queue
	tag string
}

// RedisJobQueueOption represents an option for the job redis queue
type RedisJobQueueOption func(*RedisJobQueue)

// WithJobTag sets the consumer tag
func WithJobTag(tag string) RedisJobQueueOption {
	return func(q *RedisJobQueue) {
		q.tag = tag
	}
}

// NewRedisJobQueue creates a new queue for a job in redis
func NewRedisJobQueue(q rmq.Queue, options ...RedisJobQueueOption) *RedisJobQueue {
	jq := &RedisJobQueue{
		q: q,
	}
	for _, f := range options {
		f(jq)
	}
	return jq
}

// Publish writes the payload to the redis queue
func (q *RedisJobQueue) Publish(ctx context.Context, job *api.JobRequest) error {
	// serialize the job message
	msg, err := proto.Marshal(job)
	if err != nil {
		return errors.Wrap(err, "could not serialize message")
	}
	err = q.q.PublishBytes(msg)
	return errors.Wrap(err, "could not send message to queue")
}

// Consume reads the payload from the redis queue
func (q *RedisJobQueue) Consume(ctx context.Context, w JobWorker) error {
	_, err := q.q.AddConsumerFunc(q.tag, func(delivery rmq.Delivery) {
		job := new(api.JobRequest)
		err := proto.Unmarshal([]byte(delivery.Payload()), job)
		if err != nil {
			// TODO: handle rejection errors
			delivery.Reject()
			return
		}

		// TODO: handle ack errors
		delivery.Ack()
		w.Consume(ctx, job)
	})
	return err
}
