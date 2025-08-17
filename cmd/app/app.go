package app

import (
	"DeBlockTest/internal/config"
	"DeBlockTest/pkg/addresses"
	"DeBlockTest/pkg/httpserver"
	"DeBlockTest/pkg/monitoring"
	"DeBlockTest/pkg/processing"
	"DeBlockTest/pkg/storage/postgres"
	"DeBlockTest/pkg/storage/redis"
	"DeBlockTest/pkg/transport"
	"context"

	"github.com/caarlos0/env/v6"
	"github.com/pkg/errors"
	"github.com/tel-io/tel/v2"
	"golang.org/x/sync/errgroup"
)

type Server struct{}

func New() *Server { return &Server{} }

func (s *Server) Run(ctx context.Context) error {
	return s.Action(ctx)
}

func errHandle(str string, err error) {
	if err == nil {
		return
	}
	tel.Global().Fatal(str, tel.Error(err))
}

func (s *Server) Action(ctx context.Context) error {

	cfg := &config.Config{}
	errHandle("config load error", env.Parse(cfg))

	tel.Global().Info("starting DeBlock monitoring service",
		tel.String("stage", cfg.Stage),
		tel.String("instance_id", cfg.InstanceID),
		tel.Int("worker_count", cfg.WorkerCount))

	postgresClient, err := postgres.Create(ctx, &cfg.Database)
	errHandle("postgres connection error", err)
	defer postgresClient.Close()

	redisClient, err := redis.Create(ctx, &cfg.Redis)
	errHandle("redis connection error", err)
	defer redisClient.Close()

	transportModule, err := transport.NewTransportModule(ctx, &cfg.Kafka, &cfg.Ethereum)
	errHandle("transport initialization error", err)
	defer transportModule.Close()

	addressModule, err := addresses.NewAddressModule(ctx, postgresClient, redisClient)
	errHandle("address module initialization error", err)

	processingModule, err := processing.NewProcessingModule(ctx, postgresClient, cfg.InstanceID)
	errHandle("processing module initialization error", err)

	tel.Global().Info("all modules initialized successfully",
		tel.Int("monitored_addresses", addressModule.GetAddressCount()))

	httpSrv := httpserver.NewHTTPServer(&cfg.HTTP, addressModule, processingModule)

	wgroup, _ := errgroup.WithContext(ctx)

	wgroup.Go(func() error {
		tel.Global().Info("starting HTTP server")
		return httpSrv.Start(ctx)
	})

	wgroup.Go(func() error {
		tel.Global().Info("starting blockchain monitor")
		return s.startMonitoring(ctx, transportModule, addressModule, processingModule, cfg)
	})

	return errors.WithStack(wgroup.Wait())
}

func (s *Server) startMonitoring(
	ctx context.Context,
	transport *transport.TransportModule,
	addresses *addresses.AddressModule,
	processing *processing.ProcessingModule,
	cfg *config.Config,
) error {
	tel.Global().Info("starting blockchain monitoring service",
		tel.String("instance_id", cfg.InstanceID))

	return monitoring.NewMonitoringModule(transport, addresses, processing, cfg.InstanceID).StartMonitoring(ctx)
}
