module github.com/nostosx/servex-examples

go 1.27.3

require (
	github.com/aws/aws-sdk-go-v2 v1.47.5
	github.com/aws/aws-sdk-go-v2/credentials v1.20.12
	github.com/aws/aws-sdk-go-v2/service/s3 v1.114.6
	github.com/hibiken/asynq v0.26.0
	github.com/jackc/pgx/v5 v5.11.0
	github.com/nostosx/servex v1.5.0
	github.com/nostosx/servex-contrib/auth/login v1.0.0
	github.com/nostosx/servex-contrib/auth/oauth2 v1.0.0
	github.com/nostosx/servex-contrib/auth/oauth2redis v1.0.0
	github.com/nostosx/servex-contrib/auth/oidc v1.0.0
	github.com/nostosx/servex-contrib/auth/password v1.0.0
	github.com/nostosx/servex-contrib/cache/memory v1.0.0
	github.com/nostosx/servex-contrib/cache/redis v1.1.0
	github.com/nostosx/servex-contrib/config/viper v1.0.1
	github.com/nostosx/servex-contrib/coordination/etcd v1.0.0
	github.com/nostosx/servex-contrib/database/goose v1.0.0
	github.com/nostosx/servex-contrib/database/gorm v1.0.0
	github.com/nostosx/servex-contrib/database/postgres v1.0.0
	github.com/nostosx/servex-contrib/di/do v1.2.0
	github.com/nostosx/servex-contrib/jobqueue/asynq v1.3.0
	github.com/nostosx/servex-contrib/jobqueue/pgmq v1.0.0
	github.com/nostosx/servex-contrib/lifecycle v1.0.0
	github.com/nostosx/servex-contrib/logger/zap v1.0.0
	github.com/nostosx/servex-contrib/metrics/prometheus v1.2.0
	github.com/nostosx/servex-contrib/middleware v1.5.0
	github.com/nostosx/servex-contrib/search/postgres v1.0.0
	github.com/nostosx/servex-contrib/session v1.0.0
	github.com/nostosx/servex-contrib/session/redis v1.0.0
	github.com/nostosx/servex-contrib/storage/s3 v1.0.0
	github.com/nostosx/servex-contrib/transport/cors v1.0.0
	github.com/nostosx/servex-contrib/transport/grpcx v1.3.0
	github.com/nostosx/servex-contrib/transport/httpx v1.1.0
	github.com/nostosx/servex-contrib/vector/pgvector v1.0.0
	github.com/pressly/goose/v3 v3.29.0
	github.com/prometheus/client_golang v1.25.0
	github.com/redis/go-redis/v9 v9.23.0
	github.com/rs/cors v1.11.1
	go.etcd.io/etcd/client/v3 v3.7.3
	go.uber.org/zap v1.27.1
	golang.org/x/oauth2 v0.37.0
	google.golang.org/grpc v1.84.0
	google.golang.org/protobuf v1.36.12
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gorm.io/gorm v1.31.2
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.25 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.5.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.8.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.5.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.11.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.14.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.20.8 // indirect
	github.com/aws/smithy-go v1.28.7 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/coreos/go-oidc/v3 v3.21.0 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.7.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-jose/go-jose/v4 v4.1.4 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nostosx/kitx v1.0.0 // indirect
	github.com/nostosx/servex-contrib/database/sql v1.0.0 // indirect
	github.com/nostosx/servex-contrib/encoding/json v1.0.0 // indirect
	github.com/nostosx/servex-contrib/metrics/noop v1.0.0 // indirect
	github.com/nostosx/servex-contrib/selector v1.2.0 // indirect
	github.com/nostosx/servex-contrib/tracing/noop v1.0.0 // indirect
	github.com/nostosx/servex-contrib/transport/client v1.0.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/prometheus/client_model v0.6.3 // indirect
	github.com/prometheus/common v0.72.0 // indirect
	github.com/prometheus/procfs v0.22.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/sagikazarmark/locafero v0.11.0 // indirect
	github.com/samber/do/v2 v2.1.0 // indirect
	github.com/samber/go-type-to-string v1.8.0 // indirect
	github.com/sethvargo/go-retry v0.5.0 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spf13/viper v1.21.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.etcd.io/etcd/api/v3 v3.7.3 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.7.3 // indirect
	go.uber.org/atomic v1.12.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/net v0.62.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20261005182115-fad411399dd8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261005182115-fad411399dd8 // indirect
	gorm.io/driver/postgres v1.6.3 // indirect
)
