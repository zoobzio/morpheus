// Package main is the entry point for the admin API binary.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/zoobzio/aperture"
	"github.com/zoobzio/astql/postgres"
	"github.com/zoobzio/capitan"
	grubredis "github.com/zoobzio/grub/redis"
	"github.com/zoobzio/sum"
	"github.com/zoobzio/sumatra/admin/contracts"
	"github.com/zoobzio/sumatra/admin/handlers"
	"github.com/zoobzio/sumatra/config"
	"github.com/zoobzio/sumatra/events"
	intotel "github.com/zoobzio/sumatra/internal/otel"
	"github.com/zoobzio/sumatra/stores"

	_ "github.com/lib/pq"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	log.Println("admin: starting...")
	ctx := context.Background()

	// Initialize sum service and registry.
	svc := sum.New()
	k := sum.Start()

	// =========================================================================
	// 1. Load Configuration
	// =========================================================================

	if err := sum.Config[config.App](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load app config: %w", err)
	}
	if err := sum.Config[config.Database](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load database config: %w", err)
	}
	if err := sum.Config[config.Redis](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load redis config: %w", err)
	}
	if err := sum.Config[config.Encryption](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load encryption config: %w", err)
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
	log.Println("admin: database connected")
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
	log.Println("admin: redis connected")
	capitan.Emit(ctx, events.StartupRedisConnected)

	// Create grub Redis provider for session stores
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

	// Register admin contracts
	sum.Register[contracts.Users](k, allStores.Users)
	sum.Register[contracts.Sessions](k, allStores.Sessions)
	sum.Register[contracts.Providers](k, allStores.Providers)
	log.Println("admin: stores registered")

	// =========================================================================
	// 4. Register Boundaries
	// =========================================================================

	// No model or wire boundaries required for current implementation.

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
		serviceName = "sumatra-admin"
	}

	otelProviders, err := intotel.New(ctx, intotel.Config{
		Endpoint:    otelEndpoint,
		ServiceName: serviceName,
	})
	if err != nil {
		return fmt.Errorf("failed to create otel providers: %w", err)
	}
	defer func() { _ = otelProviders.Shutdown(ctx) }()
	log.Println("admin: observability initialized")
	capitan.Emit(ctx, events.StartupOTELReady)

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

	// =========================================================================
	// 7. Register Handlers and Run
	// =========================================================================

	svc.Handle(handlers.All()...)

	appCfg := sum.MustUse[config.App](ctx)
	capitan.Emit(ctx, events.StartupServerListening, events.StartupPortKey.Field(appCfg.Port))
	log.Printf("admin: starting server on port %d...", appCfg.Port)

	_ = ap // Remove when ap.Apply() is used.

	return svc.Run("", appCfg.Port)
}
