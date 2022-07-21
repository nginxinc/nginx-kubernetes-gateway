package events_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/config/configfakes"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/file/filefakes"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/runtime/runtimefakes"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/statefakes"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/status/statusfakes"
)

type unsupportedResource struct {
	metav1.ObjectMeta
}

func (r *unsupportedResource) GetObjectKind() schema.ObjectKind {
	return nil
}

func (r *unsupportedResource) DeepCopyObject() runtime.Object {
	return nil
}

var _ = Describe("EventLoop", func() {
	var (
		fakeProcessor           *statefakes.FakeChangeProcessor
		fakeServiceStore        *statefakes.FakeServiceStore
		fakeSecretStore         *statefakes.FakeSecretStore
		fakeSecretMemoryManager *statefakes.FakeSecretDiskMemoryManager
		fakeGenerator           *configfakes.FakeGenerator
		fakeNginxFimeMgr        *filefakes.FakeManager
		fakeNginxRuntimeMgr     *runtimefakes.FakeManager
		fakeStatusUpdater       *statusfakes.FakeUpdater
		cancel                  context.CancelFunc
		eventCh                 chan interface{}
		errorCh                 chan error
		start                   func()
	)

	BeforeEach(func() {
		fakeProcessor = &statefakes.FakeChangeProcessor{}
		eventCh = make(chan interface{})
		fakeServiceStore = &statefakes.FakeServiceStore{}
		fakeSecretMemoryManager = &statefakes.FakeSecretDiskMemoryManager{}
		fakeSecretStore = &statefakes.FakeSecretStore{}
		fakeGenerator = &configfakes.FakeGenerator{}
		fakeNginxFimeMgr = &filefakes.FakeManager{}
		fakeNginxRuntimeMgr = &runtimefakes.FakeManager{}
		fakeStatusUpdater = &statusfakes.FakeUpdater{}

		ctrl := events.NewEventLoop(events.EventLoopConfig{
			Processor:           fakeProcessor,
			ServiceStore:        fakeServiceStore,
			SecretStore:         fakeSecretStore,
			SecretMemoryManager: fakeSecretMemoryManager,
			Generator:           fakeGenerator,
			EventCh:             eventCh,
			Logger:              zap.New(),
			NginxFileMgr:        fakeNginxFimeMgr,
			NginxRuntimeMgr:     fakeNginxRuntimeMgr,
			StatusUpdater:       fakeStatusUpdater,
		})

		var ctx context.Context
		ctx, cancel = context.WithCancel(context.Background())
		errorCh = make(chan error)
		start = func() {
			errorCh <- ctrl.Start(ctx)
		}
	})

	Describe("Process Gateway API resource events", func() {
		BeforeEach(func() {
			go start()
		})

		AfterEach(func() {
			cancel()

			var err error
			Eventually(errorCh).Should(Receive(&err))
			Expect(err).To(BeNil())
		})

		DescribeTable("Upsert events",
			func(e *events.UpsertEvent) {
				fakeConf := state.Configuration{}
				changed := true
				fakeStatuses := state.Statuses{}
				fakeProcessor.ProcessReturns(changed, fakeConf, fakeStatuses)

				fakeCfg := []byte("fake")
				fakeGenerator.GenerateReturns(fakeCfg, config.Warnings{})

				eventCh <- e

				Eventually(fakeProcessor.CaptureUpsertChangeCallCount).Should(Equal(1))
				Expect(fakeProcessor.CaptureUpsertChangeArgsForCall(0)).Should(Equal(e.Resource))
				Eventually(fakeProcessor.ProcessCallCount).Should(Equal(1))

				Eventually(fakeGenerator.GenerateCallCount).Should(Equal(1))
				Expect(fakeGenerator.GenerateArgsForCall(0)).Should(Equal(fakeConf))

				Eventually(fakeNginxFimeMgr.WriteHTTPServersConfigCallCount).Should(Equal(1))
				name, cfg := fakeNginxFimeMgr.WriteHTTPServersConfigArgsForCall(0)
				Expect(name).Should(Equal("http-servers"))
				Expect(cfg).Should(Equal(fakeCfg))

				Eventually(fakeNginxRuntimeMgr.ReloadCallCount).Should(Equal(1))

				Eventually(fakeStatusUpdater.UpdateCallCount).Should(Equal(1))
				_, statuses := fakeStatusUpdater.UpdateArgsForCall(0)
				Expect(statuses).Should(Equal(fakeStatuses))
			},
			Entry("HTTPRoute", &events.UpsertEvent{Resource: &v1alpha2.HTTPRoute{}}),
			Entry("Gateway", &events.UpsertEvent{Resource: &v1alpha2.Gateway{}}),
			Entry("GatewayClass", &events.UpsertEvent{Resource: &v1alpha2.GatewayClass{}}),
		)

		DescribeTable("Delete events",
			func(e *events.DeleteEvent) {
				fakeConf := state.Configuration{}
				changed := true
				fakeProcessor.ProcessReturns(changed, fakeConf, state.Statuses{})

				fakeCfg := []byte("fake")
				fakeGenerator.GenerateReturns(fakeCfg, config.Warnings{})

				eventCh <- e

				Eventually(fakeProcessor.CaptureDeleteChangeCallCount).Should(Equal(1))
				passedObj, passedNsName := fakeProcessor.CaptureDeleteChangeArgsForCall(0)
				Expect(passedObj).Should(Equal(e.Type))
				Expect(passedNsName).Should(Equal(e.NamespacedName))

				Eventually(fakeProcessor.ProcessCallCount).Should(Equal(1))

				Eventually(fakeNginxFimeMgr.WriteHTTPServersConfigCallCount).Should(Equal(1))
				name, cfg := fakeNginxFimeMgr.WriteHTTPServersConfigArgsForCall(0)
				Expect(name).Should(Equal("http-servers"))
				Expect(cfg).Should(Equal(fakeCfg))

				Eventually(fakeNginxRuntimeMgr.ReloadCallCount).Should(Equal(1))
			},
			Entry("HTTPRoute", &events.DeleteEvent{Type: &v1alpha2.HTTPRoute{}, NamespacedName: types.NamespacedName{Namespace: "test", Name: "route"}}),
			Entry("Gateway", &events.DeleteEvent{Type: &v1alpha2.Gateway{}, NamespacedName: types.NamespacedName{Namespace: "test", Name: "gateway"}}),
			Entry("GatewayClass", &events.DeleteEvent{Type: &v1alpha2.GatewayClass{}, NamespacedName: types.NamespacedName{Name: "class"}}),
		)
	})

	Describe("Process Service events", func() {
		BeforeEach(func() {
			go start()
		})

		AfterEach(func() {
			cancel()

			var err error
			Eventually(errorCh).Should(Receive(&err))
			Expect(err).To(BeNil())
		})

		It("should process upsert event", func() {
			svc := &apiv1.Service{}

			eventCh <- &events.UpsertEvent{
				Resource: svc,
			}

			Eventually(fakeServiceStore.UpsertCallCount).Should(Equal(1))
			Expect(fakeServiceStore.UpsertArgsForCall(0)).Should(Equal(svc))

			Eventually(fakeProcessor.ProcessCallCount).Should(Equal(1))
		})

		It("should process delete event", func() {
			nsname := types.NamespacedName{Namespace: "test", Name: "service"}

			eventCh <- &events.DeleteEvent{
				NamespacedName: nsname,
				Type:           &apiv1.Service{},
			}

			Eventually(fakeServiceStore.DeleteCallCount).Should(Equal(1))
			Expect(fakeServiceStore.DeleteArgsForCall(0)).Should(Equal(nsname))

			Eventually(fakeProcessor.ProcessCallCount).Should(Equal(1))
		})
	})

	Describe("Process Secret events", func() {
		BeforeEach(func() {
			go start()
		})

		AfterEach(func() {
			cancel()

			var err error
			Eventually(errorCh).Should(Receive(&err))
			Expect(err).To(BeNil())
		})

		It("should process upsert event", func() {
			secret := &apiv1.Secret{}

			eventCh <- &events.UpsertEvent{
				Resource: secret,
			}

			Eventually(fakeSecretStore.UpsertCallCount).Should(Equal(1))
			Expect(fakeSecretStore.UpsertArgsForCall(0)).Should(Equal(secret))

			Eventually(fakeProcessor.ProcessCallCount).Should(Equal(1))
		})

		It("should process delete event", func() {
			nsname := types.NamespacedName{Namespace: "test", Name: "secret"}

			eventCh <- &events.DeleteEvent{
				NamespacedName: nsname,
				Type:           &apiv1.Secret{},
			}

			Eventually(fakeSecretStore.DeleteCallCount).Should(Equal(1))
			Expect(fakeSecretStore.DeleteArgsForCall(0)).Should(Equal(nsname))

			Eventually(fakeProcessor.ProcessCallCount).Should(Equal(1))
		})
	})

	Describe("Edge cases", func() {
		AfterEach(func() {
			cancel()
		})

		DescribeTable("Edge cases for events",
			func(e interface{}) {
				go func() {
					eventCh <- e
				}()

				Expect(start).Should(Panic())
			},
			Entry("should panic for an unknown event type",
				&struct{}{}),
			Entry("should panic for an unknown type of resource in upsert event",
				&events.UpsertEvent{
					Resource: &unsupportedResource{},
				}),
			Entry("should panic for an unknown type of resource in delete event",
				&events.DeleteEvent{
					Type: &unsupportedResource{},
				}),
		)
	})
})
