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

var _ = Describe("JobQueue", func() {
	var (
		T = GinkgoT()

		ctx    context.Context
		cancel context.CancelFunc

		queue    rmq.Queue
		jq       *RedisJobQueue
		consumer *mocks.JobWorker
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())

		var err error
		queue, err = redisConn.OpenQueue("job-" + rmq.RandomString(12))
		Ω(err).ShouldNot(HaveOccurred())
		err = queue.StartConsuming(10, time.Second)
		Ω(err).ShouldNot(HaveOccurred())

		consumer = new(mocks.JobWorker)
		consumerTag := "job-consumer-" + rmq.RandomString(12)
		jq = NewRedisJobQueue(queue, WithJobTag(consumerTag))
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
		job := &api.JobRequest{
			Id: "testjob123",
			Tasks: []*api.Task{
				{
					Name: "sleep-city",
					Action: &api.Task_Sleep{
						Sleep: &api.SleepAction{
							DurationSeconds: 5,
						},
					},
				},
			},
		}

		consumer.On("Consume", ctx, mock.AnythingOfType("*api.JobRequest")).Return().Run(func(args mock.Arguments) {
			defer GinkgoRecover()
			actual := args.Get(1).(*api.JobRequest)
			Ω(actual.Id).Should(Equal(job.Id))
			close(done)
		})
		err := jq.Consume(ctx, consumer)
		Ω(err).ShouldNot(HaveOccurred())
		err = jq.Publish(ctx, job)
		Ω(err).ShouldNot(HaveOccurred())
	}, 10)
})
