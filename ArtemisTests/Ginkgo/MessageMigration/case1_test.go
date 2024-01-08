package MessageMigration_test

import (
	"context"
	"fmt"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"pack.ag/amqp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = ginkgo.Describe("MessageMigration Test", func() {
	var client *amqp.Client
	var session *amqp.Session
	var sender *amqp.Sender
	var receiver *amqp.Receiver
	var ctx context.Context
	var err error

	var kubeClient *kubernetes.Clientset
	var namespace string

	ginkgo.BeforeEach(func() {
		ctx = context.Background()

		fmt.Println("Connecting to the Artemis broker...")
		// Establish connection to the generic Artemis broker (ex-aao-hdls-svc)
		client, err = amqp.Dial(
			"amqp://ex-aao-hdls-svc.activemq-artemis-brokers.svc.cluster.local:61619",
			amqp.ConnSASLPlain("cgi", "cgi"),
			amqp.ConnIdleTimeout(30*time.Second),
		)
		if err != nil {
			fmt.Println("Error during connection:", err)
			return
		}
		fmt.Println("Connected successfully.")

		fmt.Println("Creating a session...")
		// Create a session
		session, err = client.NewSession()
		if err != nil {
			fmt.Println("Error during session creation:", err)
			// Close the client on error
			client.Close()
			return
		}
		fmt.Println("Session created successfully.")

		// Initialize Kubernetes client with in-cluster config
		config, err := clientcmd.BuildConfigFromFlags("", "")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		kubeClient, err = kubernetes.NewForConfig(config)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// Set the namespace
		namespace = "activemq-artemis-brokers"

		// Ensure the StatefulSet (deployment) exists before proceeding
		statefulSetName := "ex-aao-ss"
		_, err = kubeClient.AppsV1().StatefulSets(namespace).Get(ctx, statefulSetName, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("Error getting StatefulSet %s: %v\n", statefulSetName, err)
			return
		}
	})

	ginkgo.It("should send, delete, and check messages", func() {
		queueName := "SpecificQueue"
		messageText := "Hello, this is a test message"

		// Create a sender and send a message to the specific queue in ex-aao-ss-2 broker
		sender, err = session.NewSender(
			amqp.LinkTargetAddress(queueName),
			amqp.LinkSourceAddress("ex-aao-ss-2."+queueName),
		)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		err = sender.Send(ctx, amqp.NewMessage([]byte(messageText)))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// Print a message indicating that the message has been sent to ex-aao-ss-2
		fmt.Printf("Message sent to ex-aao-ss-2.\n")

		// Wait for a short duration
		time.Sleep(60 * time.Second)

		// Delete the ex-aao-ss-2 pod
		deletePodName := "ex-aao-ss-2"
		deletePodNamespace := "activemq-artemis-brokers"
		deletePropagationPolicy := metav1.DeletePropagationForeground
		deleteOptions := &metav1.DeleteOptions{PropagationPolicy: &deletePropagationPolicy}
		err = kubeClient.CoreV1().Pods(deletePodNamespace).Delete(ctx, deletePodName, *deleteOptions)
		gomega.Expect(err).To(gomega.BeNil(), "Error deleting pod: %v", err)
		fmt.Printf("Pod '%s' deleted successfully.\n", deletePodName)

		// Print a message indicating the start of the search
		fmt.Println("Searching for the message in other brokers...")
		time.Sleep(120 * time.Second)
		// Loop through the pod names (ex-aao-ss-0, ex-aao-ss-1) to find the specific message
		for _, broker := range []string{"ex-aao-ss-0", "ex-aao-ss-1"} {
			receiver, err = session.NewReceiver(
				amqp.LinkSourceAddress(queueName),
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Receive messages from the queue
			msg, err := receiver.Receive(ctx)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Check if the received message matches the specific message
			if string(msg.GetData()) == messageText {
				// Print where the message was found
				fmt.Printf("Message found in broker '%s'.\n", broker)

				// Accept the message
				msg.Accept()

				// Close the receiver
				receiver.Close(ctx)

				// Exit the loop as the message is found
				break
			}

			// Close the receiver
			receiver.Close(ctx)
		}

		// Print a message indicating the end of the search
		fmt.Println("Message search completed.")

		// Delete the queue
		deleteQueueManagementCommand := amqp.NewMessage([]byte(
			"DELETE QUEUE '" + queueName + "'",
		))
		managementSender, err := session.NewSender(
			amqp.LinkTargetAddress("activemq.management"),
		)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// Send the delete queue command
		err = managementSender.Send(ctx, deleteQueueManagementCommand)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// Receive the response
		response, err := managementSender.Receive(ctx)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())


		// Check if the response indicates success
		if response.ApplicationProperties["statusCode"] != 200 {
			// Print an error message if the deletion fails
			fmt.Printf("Error deleting the queue. StatusCode: %v, StatusDescription: %v\n",
				response.ApplicationProperties["statusCode"],
				response.ApplicationProperties["statusDescription"],
			)
		} else {
			// Print a success message if the deletion succeeds
			fmt.Println("Queue deleted successfully.")
		}

		// Close the management sender
		managementSender.Close(ctx)

	})

	ginkgo.AfterEach(func() {
		if sender != nil {
			sender.Close(ctx)
		}
		if session != nil {
			session.Close(ctx)
		}
		if client != nil {
			client.Close()
		}
	})
})
