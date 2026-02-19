// Package main is the entry point for the application.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/zoobzio/aegis"
	"github.com/zoobzio/aegis/proto/identity"
	"github.com/zoobzio/aperture"
	"github.com/zoobzio/astql/postgres"
	"github.com/zoobzio/capitan"
	grubredis "github.com/zoobzio/grub/redis"
	"github.com/zoobzio/sum"
	"github.com/zoobzio/sumatra/api/contracts"
	"github.com/zoobzio/sumatra/api/handlers"
	"github.com/zoobzio/sumatra/config"
	"github.com/zoobzio/sumatra/events"
	intidentity "github.com/zoobzio/sumatra/internal/identity"
	intotel "github.com/zoobzio/sumatra/internal/otel"
	"github.com/zoobzio/sumatra/stores"
	"google.golang.org/grpc"

	_ "github.com/lib/pq"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	log.Println("starting...")
	ctx := context.Background()

	// Initialize sum service and registry.
	svc := sum.New()
	k := sum.Start()

	// =========================================================================
	// 1. Load Configuration
	// =========================================================================

	// Load all configs via sum.Config[T]().
	if err := sum.Config[config.App](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load app config: %w", err)
	}
	if err := sum.Config[config.Database](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load database config: %w", err)
	}
	if err := sum.Config[config.Redis](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load redis config: %w", err)
	}
	if err := sum.Config[config.GitHub](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load github config: %w", err)
	}
	if err := sum.Config[config.Google](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load google config: %w", err)
	}
	if err := sum.Config[config.Session](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load session config: %w", err)
	}
	if err := sum.Config[config.Encryption](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load encryption config: %w", err)
	}
	if err := sum.Config[config.Postmark](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load postmark config: %w", err)
	}
	if err := sum.Config[config.Tokens](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load tokens config: %w", err)
	}
	if err := sum.Config[config.Mesh](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load mesh config: %w", err)
	}

	// =========================================================================
	// 2. Connect to Infrastructure
	// =========================================================================

	// Database (PostgreSQL)
	dbCfg := sum.MustUse[config.Database](ctx)
	db, err := sqlx.Connect("postgres", dbCfg.DSN())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()
	log.Println("database connected")
	capitan.Emit(ctx, events.StartupDatabaseConnected)

	// Redis
	redisCfg := sum.MustUse[config.Redis](ctx)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisCfg.Addr(),
		Password: redisCfg.Password,
		DB:       redisCfg.DB,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}
	defer func() { _ = redisClient.Close() }()
	log.Println("redis connected")
	capitan.Emit(ctx, events.StartupRedisConnected)

	// Create grub Redis provider for session/token stores
	redisProvider := grubredis.New(redisClient)

	// =========================================================================
	// 3. Create and Register Stores
	// =========================================================================

	// Create SQL renderer for database stores
	renderer := postgres.New()

	// Create all stores
	allStores, err := stores.New(db, renderer, redisProvider)
	if err != nil {
		return fmt.Errorf("failed to create stores: %w", err)
	}

	// Register contracts
	sum.Register[contracts.Users](k, allStores.Users)
	sum.Register[contracts.Providers](k, allStores.Providers)
	sum.Register[contracts.Sessions](k, allStores.Sessions)
	sum.Register[contracts.VerificationTokens](k, allStores.VerificationTokens)
	log.Println("stores registered")

	// =========================================================================
	// 4. Register Boundaries
	// =========================================================================

	// No model or wire boundaries required for current implementation.
	// Provider model has encryption boundaries but those are handled via
	// lifecycle hooks (BeforeSave/AfterLoad) which pull the boundary from context.

	// =========================================================================
	// 5. Freeze Registry
	// =========================================================================

	sum.Freeze(k)
	capitan.Emit(ctx, events.StartupServicesReady)

	// =========================================================================
	// 6. Initialize Observability (OTEL + Aperture)
	// =========================================================================

	otelEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelEndpoint == "" {
		otelEndpoint = "localhost:4318"
	}
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "sumatra"
	}

	otelProviders, err := intotel.New(ctx, intotel.Config{
		Endpoint:    otelEndpoint,
		ServiceName: serviceName,
	})
	if err != nil {
		return fmt.Errorf("failed to create otel providers: %w", err)
	}
	defer func() { _ = otelProviders.Shutdown(ctx) }()
	log.Println("observability initialized")
	capitan.Emit(ctx, events.StartupOTELReady)

	// Initialize aperture to bridge capitan events → OTEL.
	ap, err := aperture.New(
		capitan.Default(),
		otelProviders.Log,
		otelProviders.Metric,
		otelProviders.Trace,
	)
	if err != nil {
		return fmt.Errorf("failed to create aperture: %w", err)
	}
	defer ap.Close()
	capitan.Emit(ctx, events.StartupApertureReady)

	// Optional: Apply aperture schema for metrics/traces configuration.
	// schema, err := aperture.LoadSchemaFromYAML(schemaBytes)
	// if err != nil {
	// 	return fmt.Errorf("failed to load aperture schema: %w", err)
	// }
	// if err := ap.Apply(schema); err != nil {
	// 	return fmt.Errorf("failed to apply aperture schema: %w", err)
	// }

	// =========================================================================
	// 7. Start Mesh Node
	// =========================================================================

	meshCfg := sum.MustUse[config.Mesh](ctx)
	identityServer := intidentity.New(allStores.Users, allStores.Sessions, allStores.Providers)

	node, err := aegis.NewNodeBuilder().
		WithID(meshCfg.ID).
		WithName(meshCfg.Name).
		WithAddress(meshCfg.Addr()).
		WithServices(aegis.ServiceInfo{Name: "identity", Version: "v1"}).
		WithServiceRegistration(func(s *grpc.Server) {
			identity.RegisterIdentityServiceServer(s, identityServer)
		}).
		WithCertDir(meshCfg.CertDir).
		Build()
	if err != nil {
		return fmt.Errorf("failed to build mesh node: %w", err)
	}

	if err := node.StartServer(); err != nil {
		return fmt.Errorf("failed to start mesh server: %w", err)
	}
	defer node.Shutdown()
	log.Printf("mesh node started on %s", meshCfg.Addr())
	capitan.Emit(ctx, events.StartupMeshReady)

	// =========================================================================
	// 8. Register Handlers and Run
	// =========================================================================

	svc.Handle(handlers.All()...)

	appCfg := sum.MustUse[config.App](ctx)
	capitan.Emit(ctx, events.StartupServerListening, events.StartupPortKey.Field(appCfg.Port))
	log.Printf("starting server on port %d...", appCfg.Port)

	_ = ap // Remove when using ap.Apply() for aperture schema.

	return svc.Run("", appCfg.Port)
}
