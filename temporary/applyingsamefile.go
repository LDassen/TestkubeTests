package MessageMigration_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("ActiveMQ Artemis Deployment Test", func() {
	var dynamicClient dynamic.Interface
	var namespace string
	var resourceGVR schema.GroupVersionResource

	ginkgo.BeforeEach(func() {
		var err error
		config, err := rest.InClusterConfig()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		dynamicClient, err = dynamic.NewForConfig(config)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		namespace = "activemq-artemis-brokers"
		resourceGVR = schema.GroupVersionResource{
			Group:    "broker.amq.io",
			Version:  "v1beta1",
			Resource: "activemqartemises",
		}
	})

	ginkgo.It("Should create or update ActiveMQArtemis resource", func() {
		fileName := "ex-aaoMM.yaml"

		filePath, err := filepath.Abs(fileName)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		fileBytes, err := ioutil.ReadFile(filePath)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		obj := &unstructured.Unstructured{}
		_, _, err = decUnstructured.Decode(fileBytes, nil, obj)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		obj.SetAPIVersion("broker.amq.io/v1beta1")
		obj.SetKind("ActiveMQArtemis")

		resourceClient := dynamicClient.Resource(resourceGVR).Namespace(namespace)

		// Try to get the existing resource
		existingObj, err := resourceClient.Get(context.TODO(), obj.GetName(), metav1.GetOptions{})
		if err == nil {
			// Resource already exists, update it if needed
			if !reflect.DeepEqual(existingObj.Object, obj.Object) {
				fmt.Printf("ActiveMQArtemis resource already exists, updating configuration.\n")

				// Set the UID and ResourceVersion to perform an update
				obj.SetUID(existingObj.GetUID())
				obj.SetResourceVersion(existingObj.GetResourceVersion())

				_, err = resourceClient.Update(context.TODO(), obj, metav1.UpdateOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Error updating ActiveMQArtemis resource")
			} else {
				fmt.Printf("ActiveMQArtemis resource already exists and has the same configuration.\n")
			}
			return
		}

		// If the resource does not exist, create it
		createdObj, err := resourceClient.Create(context.TODO(), obj, metav1.CreateOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Error creating ActiveMQArtemis resource")

		// Confirm that the resource has been created
		fmt.Printf("Created ActiveMQArtemis resource: %s\n", createdObj.GetName())
	})
})
