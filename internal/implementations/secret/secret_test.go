package implementation_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
	implementation "github.com/nginxinc/nginx-kubernetes-gateway/internal/implementations/secret"
	"github.com/nginxinc/nginx-kubernetes-gateway/pkg/sdk"
)

var _ = Describe("SecretImplementation", func() {
	var (
		eventCh chan interface{}
		impl    sdk.SecretImpl
	)

	BeforeEach(func() {
		eventCh = make(chan interface{})

		impl = implementation.NewSecretImplementation(config.Config{
			Logger: zap.New(),
		}, eventCh)
	})

	const secretName = "my-secret"
	const secretNamespace = "test"

	Describe("Implementation processes Secret", func() {
		It("should process upsert", func() {
			secret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: secretNamespace,
				},
			}

			go func() {
				impl.Upsert(secret)
			}()

			Eventually(eventCh).Should(Receive(Equal(&events.UpsertEvent{Resource: secret})))
		})

		It("should process remove", func() {
			nsname := types.NamespacedName{Name: secretName, Namespace: secretNamespace}

			go func() {
				impl.Remove(nsname)
			}()

			Eventually(eventCh).Should(Receive(Equal(
				&events.DeleteEvent{
					NamespacedName: nsname,
					Type:           &v1.Secret{},
				})))
		})
	})
})
