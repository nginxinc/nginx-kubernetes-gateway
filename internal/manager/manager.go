package manager

import (
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctlr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
	gw "github.com/nginxinc/nginx-kubernetes-gateway/internal/implementations/gateway"
	gc "github.com/nginxinc/nginx-kubernetes-gateway/internal/implementations/gatewayclass"
	hr "github.com/nginxinc/nginx-kubernetes-gateway/internal/implementations/httproute"
	secret "github.com/nginxinc/nginx-kubernetes-gateway/internal/implementations/secret"
	svc "github.com/nginxinc/nginx-kubernetes-gateway/internal/implementations/service"
	ngxcfg "github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/file"
	ngxruntime "github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/runtime"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/status"
	"github.com/nginxinc/nginx-kubernetes-gateway/pkg/sdk"
)

const (
	// clusterTimeout is a timeout for connections to the Kubernetes API
	clusterTimeout = 10 * time.Second
	// secretsFolder is the folder that holds all the secrets for NGINX servers.
	// nolint:gosec
	secretsFolder = "/etc/nginx/secrets"
)

var scheme = runtime.NewScheme()

func init() {
	// FIXME(pleshakov): handle errors returned by the calls bellow
	_ = gatewayv1alpha2.AddToScheme(scheme)
	_ = apiv1.AddToScheme(scheme)
}

func Start(cfg config.Config) error {
	logger := cfg.Logger

	options := manager.Options{
		Scheme: scheme,
	}

	eventCh := make(chan interface{})

	clusterCfg := ctlr.GetConfigOrDie()
	clusterCfg.Timeout = clusterTimeout

	mgr, err := manager.New(clusterCfg, options)
	if err != nil {
		return fmt.Errorf("cannot build runtime manager: %w", err)
	}

	err = sdk.RegisterGatewayClassController(mgr, gc.NewGatewayClassImplementation(cfg, eventCh))
	if err != nil {
		return fmt.Errorf("cannot register gatewayclass implementation: %w", err)
	}
	err = sdk.RegisterGatewayController(mgr, gw.NewGatewayImplementation(cfg, eventCh))
	if err != nil {
		return fmt.Errorf("cannot register gateway implementation: %w", err)
	}
	err = sdk.RegisterHTTPRouteController(mgr, hr.NewHTTPRouteImplementation(cfg, eventCh))
	if err != nil {
		return fmt.Errorf("cannot register httproute implementation: %w", err)
	}
	err = sdk.RegisterServiceController(mgr, svc.NewServiceImplementation(cfg, eventCh))
	if err != nil {
		return fmt.Errorf("cannot register service implementation: %w", err)
	}
	err = sdk.RegisterSecretController(mgr, secret.NewSecretImplementation(cfg, eventCh))
	if err != nil {
		return fmt.Errorf("cannot register secret implementation: %w", err)
	}

	secretStore := state.NewSecretStore()
	secretMemoryMgr := state.NewSecretDiskMemoryManager(secretsFolder, secretStore)

	processor := state.NewChangeProcessorImpl(state.ChangeProcessorConfig{
		GatewayCtlrName:     cfg.GatewayCtlrName,
		GatewayClassName:    cfg.GatewayClassName,
		SecretMemoryManager: secretMemoryMgr,
	})

	serviceStore := state.NewServiceStore()
	configGenerator := ngxcfg.NewGeneratorImpl(serviceStore)
	nginxFileMgr := file.NewManagerImpl()
	nginxRuntimeMgr := ngxruntime.NewManagerImpl()
	statusUpdater := status.NewUpdater(status.UpdaterConfig{
		GatewayCtlrName:  cfg.GatewayCtlrName,
		GatewayClassName: cfg.GatewayClassName,
		Client:           mgr.GetClient(),
		// FIXME(pleshakov) Make sure each component:
		// (1) Has a dedicated named logger.
		// (2) Get it from the Manager (the WithName is done here for all components).
		Logger: cfg.Logger.WithName("statusUpdater"),
		Clock:  status.NewRealClock(),
	})

	eventLoop := events.NewEventLoop(events.EventLoopConfig{
		Processor:           processor,
		ServiceStore:        serviceStore,
		SecretStore:         secretStore,
		SecretMemoryManager: secretMemoryMgr,
		Generator:           configGenerator,
		EventCh:             eventCh,
		Logger:              cfg.Logger.WithName("eventLoop"),
		NginxFileMgr:        nginxFileMgr,
		NginxRuntimeMgr:     nginxRuntimeMgr,
		StatusUpdater:       statusUpdater,
	})

	err = mgr.Add(eventLoop)
	if err != nil {
		return fmt.Errorf("cannot register event loop: %w", err)
	}

	ctx := ctlr.SetupSignalHandler()

	logger.Info("Starting manager")
	return mgr.Start(ctx)
}
