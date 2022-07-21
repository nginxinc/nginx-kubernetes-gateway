package implementation_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGatewayImplementation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gateway Implementation Suite")
}
