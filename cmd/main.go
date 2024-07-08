package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"

	"github.com/crt379/registerdiscovery"
	"github.com/crt379/svc-collector-grpc-gw/internal/config"
	"github.com/crt379/svc-collector-grpc-gw/internal/logging"
	"github.com/crt379/svc-collector-grpc-gw/internal/middleware"
	"github.com/crt379/svc-collector-grpc-gw/internal/storage"
	appapigw "github.com/crt379/svc-collector-grpc-proto/appapi"
	appgw "github.com/crt379/svc-collector-grpc-proto/application"
	appprocgw "github.com/crt379/svc-collector-grpc-proto/appproc"
	appsvcgw "github.com/crt379/svc-collector-grpc-proto/appsvc"
	processorgw "github.com/crt379/svc-collector-grpc-proto/processor"
	registergw "github.com/crt379/svc-collector-grpc-proto/register"
	servicegw "github.com/crt379/svc-collector-grpc-proto/service"
	svcapigw "github.com/crt379/svc-collector-grpc-proto/svcapi"
	svcapieggw "github.com/crt379/svc-collector-grpc-proto/svcapieg"
	tenantgw "github.com/crt379/svc-collector-grpc-proto/tenant"
)

func main() {
	defer logging.LoggerSync()

	logger := zap.L()
	logger.Info("starting")

	zlogger := logging.NewZapLogger(logger)
	grpclog.SetLoggerV2(zlogger)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	listenAddr := fmt.Sprintf("%s:%s", config.AppConfig.Listen.Host, config.AppConfig.Listen.Port)

	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			switch key {
			case "X-Access-Trace-Id":
				return key, true
			case "X-Access-Tenant":
				return key, true
			default:
				return runtime.DefaultHeaderMatcher(key)
			}
		}),
	)

	registerdiscovery.RegisterResolver(storage.EtcdClient)

	roundrobinConn, err := grpc.NewClient(
		fmt.Sprintf("%s:///%s", "etcd", config.AppConfig.Service.Name),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}

	rs := []func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error{
		tenantgw.RegisterTenantHandler,
		servicegw.RegisterServiceHandler,
		svcapigw.RegisterSvcapiHandler,
		svcapieggw.RegisterSvcapiegHandler,
		appgw.RegisterApplicationHandler,
		appsvcgw.RegisterAppsvcHandler,
		registergw.RegisterRegisterHandler,
		processorgw.RegisterProcessorHandler,
		appapigw.RegisterAppapiHandler,
		appprocgw.RegisterAppprocHandler,
	}
	for _, r := range rs {
		err = r(ctx, mux, roundrobinConn)
		if err != nil {
			panic(err)
		}
	}

	srv := &http.Server{
		Addr:    listenAddr,
		Handler: middleware.WithRecover(middleware.WithLogger(mux)),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Info("listen err", zap.String("error", err.Error()))
		}
	}()

	registerdiscovery.RegisterService(
		context.Background(),
		config.AppConfig.Register.Name,
		config.AppConfig.Addr,
		storage.EtcdClient,
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	logger.Info("Shutdown Server")
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Info("Server Shutdown Error", zap.String("error", err.Error()))
	}
	logger.Info("Server exiting")
}
