package Deployment_test

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

var _ = Describe("Check the ActiveMQ Artemis Operator Pod", func() {
	It("should have operator pod running in 'activemq-artermis-operator' namespace", func() {
		config, err := rest.InClusterConfig()
		Expect(err).To(BeNil(), "Error getting in-cluster config: %v", err)

		clientset, err := kubernetes.NewForConfig(config)
		Expect(err).To(BeNil(), "Error creating Kubernetes client: %v", err)

		namespace := "activemq-artermis-operator"
		expectedPodPrefix := "activemq-artemis-controller-manager"

		pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
		Expect(err).To(BeNil(), "Error getting pods: %v", err)

		var actualPodCount int
		for _, pod := range pods.Items {
			if strings.HasPrefix(pod.Name, expectedPodPrefix) && pod.Status.Phase == "Running" {
				fmt.Printf("Operator Pod Name: %s\n", pod.Name)
				actualPodCount++
			}
		}

		// Set your expected number of operator pods here
		expectedPodCount := 1
		Expect(actualPodCount).To(Equal(expectedPodCount), "Expected %d 'operator-pod' pod, but found %d", expectedPodCount, actualPodCount)
	})
})
