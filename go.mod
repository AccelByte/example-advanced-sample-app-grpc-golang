module matchmaking-function-grpc-plugin-server-go

go 1.18

require (
	github.com/AccelByte/accelbyte-go-sdk v0.36.0
	github.com/AccelByte/go-jose v2.1.4+incompatible
	github.com/elliotchance/pie/v2 v2.4.0
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.14.0
	github.com/sirupsen/logrus v1.9.0
	github.com/stretchr/testify v1.8.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.36.1
	go.opentelemetry.io/contrib/propagators/b3 v1.11.1
	go.opentelemetry.io/otel v1.11.1
	go.opentelemetry.io/otel/exporters/zipkin v1.11.1
	go.opentelemetry.io/otel/sdk v1.11.1
	golang.org/x/exp v0.0.0-20220321173239-a90fa8a75705
	google.golang.org/grpc v1.50.1
	google.golang.org/protobuf v1.28.1
)

require golang.org/x/sys v0.2.0 // indirect
