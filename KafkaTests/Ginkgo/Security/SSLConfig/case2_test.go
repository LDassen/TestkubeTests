package SSLConfig_test

import (
	"context"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	certmanagerclient "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var _ = ginkgo.Describe("Kafka Certificates and Secrets", func() {
	var clientset *kubernetes.Clientset
	var certManagerClientset *certmanagerclient.Clientset

	ginkgo.BeforeEach(func() {
		// Set up the Kubernetes client
		config, err := rest.InClusterConfig()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		clientset, err = kubernetes.NewForConfig(config)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// Set up the cert-manager client
		certManagerClientset, err = certmanagerclient.NewForConfig(config)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.Context("in kafka-brokers namespace", func() {
		namespace := "kafka-brokers"

		ginkgo.It("should have two ready certificates and specific secrets", func() {
			certNames := []string{
				"kafka-brokers-controller.kafka-brokers.mgt.cluster.local",
				"kafka-brokers-headless.kafka-brokers.svc.cluster.local",
			}

			secretNames := []string{
				"kafka-brokers-controller",
				"kafka-brokers-server-certificate",
			}

			for _, certName := range certNames {
				cert, err := certManagerClientset.CertmanagerV1().Certificates(namespace).Get(context.TODO(), certName, metav1.GetOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Check if certificate is ready
				ready := false
				for _, condition := range cert.Status.Conditions {
					if condition.Type == certmanagerv1.CertificateConditionReady && condition.Status == "True" {
						ready = true
						break
					}
				}
				gomega.Expect(ready).To(gomega.BeTrue())
			}

			for _, secretName := range secretNames {
				_, err := clientset.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			}
		})
	})
})
