package implementation_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGatewayClassImplementation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gateway Class Implementation Suite")
}
