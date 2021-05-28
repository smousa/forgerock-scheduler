package scheduler_test

import (
	"context"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/adjust/rmq/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	redisConn       rmq.Connection
	firestoreClient *firestore.Client
)

func TestScheduler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scheduler Suite")
}

var _ = BeforeSuite(func() {
	redisAddress := os.Getenv("REDIS_ADDRESS")

	var err error
	redisConn, err = rmq.OpenConnection("test", "tcp", redisAddress, 0, nil)
	Ω(err).ShouldNot(HaveOccurred())

	firestoreProjectID := os.Getenv("FIRESTORE_PROJECT_ID")
	firestoreClient, err = firestore.NewClient(context.Background(), firestoreProjectID)
	Ω(err).ShouldNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	if redisConn != nil {
		<-redisConn.StopAllConsuming()
		redisConn = nil
	}

	if firestoreClient != nil {
		firestoreClient.Close()
		firestoreClient = nil
	}
})
