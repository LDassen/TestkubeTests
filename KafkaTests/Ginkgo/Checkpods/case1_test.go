package Checkpods_test

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

var _ = Describe("Kafka Broker Pods", func() {
	It("should have the correct 'kafka-brokers-' prefixed pods running", func() {
		config, err := rest.InClusterConfig()
		Expect(err).To(BeNil(), "Error getting in-cluster config: %v", err)

		clientset, err := kubernetes.NewForConfig(config)
		Expect(err).To(BeNil(), "Error creating Kubernetes client: %v", err)

		namespace := "kafka-brokers"
		expectedPodPrefix := "kafka-brokers-"

		pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
		Expect(err).To(BeNil(), "Error getting pods: %v", err)

		var actualPodCount int
		for _, pod := range pods.Items {
			if strings.HasPrefix(pod.Name, expectedPodPrefix) && pod.Status.Phase == "Running" {
				fmt.Printf("Pod Name: %s\n", pod.Name)
				actualPodCount++
			}
		}

		// Set your expected number of 'kafka-brokers-' pods here
		expectedPodCount := 3 
		Expect(actualPodCount).To(Equal(expectedPodCount), "Expected %d 'kafka-brokers-' pods, but found %d", expectedPodCount, actualPodCount)
	})
})
