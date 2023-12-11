package test_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func TestArtemis(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Artemis Suite")
}

var _ = ginkgo.Describe("Artemis Broker Setup", func() {
	g := gomega.NewGomegaWithT(ginkgo.GinkgoT())

	// Your test goes here
	ginkgo.It("should have three brokers running", func() {
		// Load Kubernetes config
		config, err := loadKubeConfig()
		g.Expect(err).NotTo(gomega.HaveOccurred())

		// Create Kubernetes client
		clientset, err := kubernetes.NewForConfig(config)
		g.Expect(err).NotTo(gomega.HaveOccurred())

		// Get Artemis broker pods
		podList, err := clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{
			LabelSelector: "app=artemis-broker",
		})
		g.Expect(err).NotTo(gomega.HaveOccurred())

		// Assert that there are exactly 3 Artemis broker pods
		g.Expect(len(podList.Items)).To(gomega.Equal(3), "Expected 3 Artemis brokers, but found %d", len(podList.Items))
	})
})

func loadKubeConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		// If not running inside a cluster, use kubeconfig
		kubeconfig, found := os.LookupEnv("KUBECONFIG")
		if !found {
			kubeconfig = filepath.Join(homeDir(), ".kube", "config")
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}
