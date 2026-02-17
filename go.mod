module github.com/zoobzio/sumatra

go 1.24.0

toolchain go1.25.3

require (
	github.com/jmoiron/sqlx v1.4.0
	github.com/lib/pq v1.10.9
	github.com/minio/minio-go/v7 v7.0.80
	github.com/zoobzio/aperture v0.1.0
	github.com/zoobzio/astql v1.0.5
	github.com/zoobzio/capitan v1.0.0
	github.com/zoobzio/cereal v0.1.1
	github.com/zoobzio/check v0.0.3
	github.com/zoobzio/grub v0.1.5
	github.com/zoobzio/rocco v0.1.10
	github.com/zoobzio/sum v0.1.0
	go.opentelemetry.io/otel v1.33.0
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.9.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.33.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.33.0
	go.opentelemetry.io/otel/log v0.9.0
	go.opentelemetry.io/otel/metric v1.33.0
	go.opentelemetry.io/otel/sdk v1.33.0
	go.opentelemetry.io/otel/sdk/log v0.9.0
	go.opentelemetry.io/otel/sdk/metric v1.33.0
	go.opentelemetry.io/otel/trace v1.33.0
)
