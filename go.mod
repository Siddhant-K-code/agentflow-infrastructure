module github.com/agentflow/infrastructure

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/google/uuid v1.4.0
	github.com/lib/pq v1.10.9
	github.com/ClickHouse/clickhouse-go/v2 v2.15.0
	github.com/redis/go-redis/v9 v9.3.0
	github.com/nats-io/nats.go v1.31.0
	github.com/nats-io/jetstream v0.0.0-20231025203626-ad8b3e8b8e8e
	github.com/golang-migrate/migrate/v4 v4.16.2
	github.com/spf13/cobra v1.8.0
	github.com/spf13/viper v1.17.0
	github.com/stretchr/testify v1.8.4
	github.com/prometheus/client_golang v1.17.0
	github.com/gorilla/websocket v1.5.1
	github.com/bytedance/sonic v1.10.2
	github.com/openfga/go-sdk v0.3.5
	github.com/wasmtime-go v4.0.0+incompatible
	google.golang.org/grpc v1.59.0
	google.golang.org/protobuf v1.31.0
	gopkg.in/yaml.v3 v3.0.1
)