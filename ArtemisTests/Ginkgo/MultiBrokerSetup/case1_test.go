package MultiBrokerSetup_test

import (
    "bytes"
    "context"
    "fmt"
    "io/ioutil"
    "path/filepath"
    "time"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
    "k8s.io/apimachinery/pkg/runtime/schema"
    "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
    "k8s.io/client-go/dynamic"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/apimachinery/pkg/util/wait"
    "k8s.io/apimachinery/pkg/labels"

    "github.com/onsi/ginkgo/v2"
    "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("ActiveMQ Artemis Deployment Test", func() {
    var clientset *kubernetes.Clientset
    var dynamicClient dynamic.Interface

    ginkgo.BeforeEach(func() {
        config, err := rest.InClusterConfig()
        gomega.Expect(err).NotTo(gomega.HaveOccurred())

        clientset, err = kubernetes.NewForConfig(config)
        gomega.Expect(err).NotTo(gomega.HaveOccurred())

        dynamicClient, err = dynamic.NewForConfig(config)
        gomega.Expect(err).NotTo(gomega.HaveOccurred())
    })

    ginkgo.It("Should create ActiveMQArtemis resource and verify pods", func() {
        fileName := "ex-aao.yaml"
        namespace := "activemq-artemis-brokers"

        filePath, err := filepath.Abs(fileName)
        gomega.Expect(err).NotTo(gomega.HaveOccurred())

        fileBytes, err := ioutil.ReadFile(filePath)
        gomega.Expect(err).NotTo(gomega.HaveOccurred())

        decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
        obj := &unstructured.Unstructured{}
        _, gvk, err := decUnstructured.Decode(fileBytes, nil, obj)
        gomega.Expect(err).NotTo(gomega.HaveOccurred())

        obj.SetAPIVersion("broker.amq.io/v1beta1")
        obj.SetKind("ActiveMQArtemis")

        // Extract the replication factor from the YAML file
        replicationFactor, found, err := unstructured.NestedInt64(obj.Object, "spec", "deploymentPlan", "size")
        gomega.Expect(err).NotTo(gomega.HaveOccurred())
        gomega.Expect(found).To(gomega.BeTrue(), "Replication factor (size) not found in the YAML file")
        expectedPodCount := int(replicationFactor)

        resourceClient := dynamicClient.Resource(schema.GroupVersionResource{
            Group:    gvk.Group,    // or "broker.amq.io"
            Version:  gvk.Version,  // or "v1beta1"
            Resource: "activemqartemises", // Adjust according to the CRD definition
        }).Namespace(namespace)

        _, err = resourceClient.Create(context.TODO(), obj, metav1.CreateOptions{})
        gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Error creating ActiveMQArtemis resource")

        // Wait for the pods to be running
        labelSelector := fmt.Sprintf("app=%s", obj.GetLabels()["app"])
        err = wait.PollImmediate(5*time.Second, 2*time.Minute, func() (bool, error) {
            podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
                LabelSelector: labelSelector,
            })
            if err != nil {
                return false, err
            }
            runningPods := 0
            for _, pod := range podList.Items {
                if pod.Status.Phase == "Running" {
                    runningPods++
                }
            }
            return runningPods == expectedPodCount, nil
        })
        gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Error waiting for broker pods to be running")

        // Delete the ActiveMQArtemis resource
        err = resourceClient.Delete(context.TODO(), obj.GetName(), metav1.DeleteOptions{})
        gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Error deleting ActiveMQArtemis resource")
    })

    // ... [other tests] ...
})
