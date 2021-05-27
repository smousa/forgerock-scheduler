package scheduler

import (
	"context"

	"github.com/adjust/rmq/v4"
	"github.com/pkg/errors"
	"github.com/smousa/forgerock-scheduler/api"
	"google.golang.org/protobuf/proto"
)

// TaskQueueName is the name of the task queue
const TaskQueueName = "tasks"

// TaskProducer inserts tasks into the task queue
type TaskProducer interface {
	Publish(ctx context.Context, task *api.TaskRequest) error
}

// TaskConsumer reads tasks from the task queue
type TaskConsumer interface {
	Consume(ctx context.Context, w TaskWorker) error
}

// RedisTaskQueueOption represents an option for the job redis queue
type RedisTaskQueueOption func(*RedisTaskQueue)

// WithTaskTag sets the consumer tag
func WithTaskTag(tag string) RedisTaskQueueOption {
	return func(q *RedisTaskQueue) {
		q.tag = tag
	}
}

// RedisTaskQueue represents a redis queue for a particular job
type RedisTaskQueue struct {
	q   rmq.Queue
	tag string
}

// NewRedisTaskQueue creates a new queue for a job in redis
func NewRedisTaskQueue(q rmq.Queue, options ...RedisTaskQueueOption) *RedisTaskQueue {
	jq := &RedisTaskQueue{
		q: q,
	}
	for _, f := range options {
		f(jq)
	}
	return jq
}

// Publish writes the payload to the redis queue
func (q *RedisTaskQueue) Publish(ctx context.Context, task *api.TaskRequest) error {
	// serialize the job message
	msg, err := proto.Marshal(task)
	if err != nil {
		return errors.Wrap(err, "could not serialize message")
	}
	err = q.q.PublishBytes(msg)
	return errors.Wrap(err, "could not send message to queue")
}

// Consume reads the payload from the redis queue
func (q *RedisTaskQueue) Consume(ctx context.Context, w TaskWorker) error {
	_, err := q.q.AddConsumerFunc(q.tag, func(delivery rmq.Delivery) {
		task := new(api.TaskRequest)
		err := proto.Unmarshal([]byte(delivery.Payload()), task)
		if err != nil {
			// TODO: handle rejection errors
			delivery.Reject()
			return
		}

		// TODO: handle ack errors
		delivery.Ack()
		w.Consume(ctx, task)
	})
	return err
}
