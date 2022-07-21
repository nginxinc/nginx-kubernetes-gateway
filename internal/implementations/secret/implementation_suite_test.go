package implementation_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSecretImplementation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Secret Implementation Suite")
}
