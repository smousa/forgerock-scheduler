package scheduler_test

import (
	"context"
	"time"

	"github.com/adjust/rmq/v4"
	. "github.com/smousa/forgerock-scheduler/scheduler"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Store", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc

		store *FirestoreStore
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
		store = NewFirestoreStore(firestoreClient)
	})

	AfterEach(func() {

		cancel()
	})

	It("creates a new job", func() {
		job := &Job{
			Tasks: []*Task{
				{
					Name:   "sleep-city",
					Action: SleepAction,
					Params: map[string]string{
						"duration_seconds": "5",
					},
				},
			},
		}

		id, err := store.Add(ctx, job)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(id).ShouldNot(BeEmpty())

		doc := firestoreClient.Doc("jobs/" + id)
		defer doc.Delete(ctx)
		snap, err := doc.Get(ctx)
		Ω(err).ShouldNot(HaveOccurred())
		actual := new(Job)
		err = snap.DataTo(actual)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(actual).Should(Equal(job))
	})

	It("updates the job status", func() {
		jobID := rmq.RandomString(7)
		status := &Status{
			TaskName: "sleep-city",
			State:    RunningState,
			Message:  "zzzz",
		}
		store.UpdateStatus(ctx, jobID, status)

		iter := firestoreClient.Collection("jobs/" + jobID + "/status").Documents(ctx)
		snaps, err := iter.GetAll()
		Ω(err).ShouldNot(HaveOccurred())

		defer func() {
			for _, snap := range snaps {
				snap.Ref.Delete(ctx)
			}
		}()

		Ω(snaps).Should(HaveLen(1))
		actual := new(Status)
		err = snaps[0].DataTo(actual)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(actual).Should(Equal(status))
	})

	It("waits for a task to complete", func(done Done) {
		jobID := rmq.RandomString(7)
		status := &Status{
			TaskName: "sleep-city",
			State:    SuccessState,
		}

		go func() {
			defer GinkgoRecover()
			actual, err := store.WaitForTaskToComplete(ctx, jobID, "sleep-city")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(actual).Should(Equal(status))

			iter := firestoreClient.Collection("jobs/" + jobID + "/status").Documents(ctx)
			snaps, err := iter.GetAll()
			Ω(err).ShouldNot(HaveOccurred())
			for _, snap := range snaps {
				snap.Ref.Delete(ctx)
			}
			close(done)
		}()

		time.Sleep(1 * time.Second)
		store.UpdateStatus(ctx, jobID, status)
	}, 10)
})
