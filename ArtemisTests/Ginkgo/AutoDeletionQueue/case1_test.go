package AutoDeletionQueue_test

import (
    "context"
    "github.com/onsi/ginkgo/v2"
    "github.com/onsi/gomega"
    "pack.ag/amqp" // AMQP library for Go
    "time"
)

var _ = ginkgo.Describe("Artemis Queue Deletion Test with AMQP", func() {
    var client *amqp.Client
    var session *amqp.Session
    var sender *amqp.Sender
    var receiver *amqp.Receiver
    var ctx context.Context
    var err error

    ginkgo.BeforeEach(func() {
        ctx = context.Background()
        client, err = amqp.Dial("amqp://ex-aao-hdls-svc.activemq-artemis-brokers.svc.cluster.local:61619", amqp.ConnSASLPlain("cgi", "cgi")) // Replace with actual credentials and Artemis server address
        gomega.Expect(err).NotTo(gomega.HaveOccurred())
        session, err = client.NewSession()
        gomega.Expect(err).NotTo(gomega.HaveOccurred())
    })

    ginkgo.It("should check if the queue is deleted after use", func() {
        queueName := "TESTKUBE"
        messageText := "Test message for deletion check"

        // Create a sender
        sender, err = session.NewSender(
            amqp.LinkTargetAddress(queueName),
        )
        gomega.Expect(err).NotTo(gomega.HaveOccurred())

        // Send a message
        err = sender.Send(ctx, amqp.NewMessage([]byte(messageText)))
        gomega.Expect(err).NotTo(gomega.HaveOccurred())
        sender.Close(ctx)

        // Create a receiver
        receiver, err = session.NewReceiver(
            amqp.LinkSourceAddress(queueName),
        )
        gomega.Expect(err).NotTo(gomega.HaveOccurred())

        // Receive the message
        msg, err := receiver.Receive(ctx)
        gomega.Expect(err).NotTo(gomega.HaveOccurred())
        gomega.Expect(string(msg.GetData())).To(gomega.Equal(messageText))
        msg.Accept()
        receiver.Close(ctx)

        // Wait for 2 minute
        time.Sleep(2 * time.Minute)

        // Check if the queue is deleted
        _, err = session.NewReceiver(
            amqp.LinkSourceAddress(queueName),
        )
        gomega.Expect(err).To(gomega.BeNil()) // Expect a nil if the queue is deleted
    })

    ginkgo.AfterEach(func() {
        if sender != nil {
            sender.Close(ctx)
        }
        if receiver != nil {
            receiver.Close(ctx)
        }
        if session != nil {
            session.Close(ctx)
        }
        if client != nil {
            client.Close()
        }
    })
})
