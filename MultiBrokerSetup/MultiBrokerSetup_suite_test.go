package MultiBrokerSetup_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMultiBrokerSetup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MultiBrokerSetup Suite")
}
