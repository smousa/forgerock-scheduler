package scheduler_test

import (
	"context"
	"time"

	"github.com/adjust/rmq/v4"
	"github.com/smousa/forgerock-scheduler/api"
	. "github.com/smousa/forgerock-scheduler/scheduler"
	"github.com/smousa/forgerock-scheduler/scheduler/mocks"
	"github.com/stretchr/testify/mock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TaskQueue", func() {
	var (
		T = GinkgoT()

		ctx    context.Context
		cancel context.CancelFunc

		queue    rmq.Queue
		tq       *RedisTaskQueue
		consumer *mocks.TaskWorker
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())

		var err error
		queue, err = redisConn.OpenQueue("task-" + rmq.RandomString(12))
		Ω(err).ShouldNot(HaveOccurred())
		err = queue.StartConsuming(10, time.Second)
		Ω(err).ShouldNot(HaveOccurred())

		consumer = new(mocks.TaskWorker)
		consumerTag := "task-consumer-" + rmq.RandomString(12)
		tq = NewRedisTaskQueue(queue, WithTaskTag(consumerTag))
	})

	AfterEach(func() {
		cancel()
		if queue != nil {
			<-queue.StopConsuming()
			queue = nil
		}
		consumer.AssertExpectations(T)
	})

	It("publishes and receives a message from the queue", func(done Done) {
		task := &api.TaskRequest{
			JobId: "testtask123",
			Task: &api.Task{
				Name: "sleep-city",
				Action: &api.Task_Sleep{
					Sleep: &api.SleepAction{
						DurationSeconds: 5,
					},
				},
			},
		}

		consumer.On("Consume", ctx, mock.AnythingOfType("*api.TaskRequest")).Return().Run(func(args mock.Arguments) {
			defer GinkgoRecover()
			actual := args.Get(1).(*api.TaskRequest)
			Ω(actual.JobId).Should(Equal(task.JobId))
			close(done)
		})
		err := tq.Consume(ctx, consumer)
		Ω(err).ShouldNot(HaveOccurred())
		err = tq.Publish(ctx, task)
		Ω(err).ShouldNot(HaveOccurred())
	}, 10)
})
