package user

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	asynqsdk "github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/nostosx/servex-contrib/auth/oauth2"
	"github.com/nostosx/servex-contrib/auth/oauth2redis"
	cachememory "github.com/nostosx/servex-contrib/cache/memory"
	cacheredis "github.com/nostosx/servex-contrib/cache/redis"
	etcdx "github.com/nostosx/servex-contrib/coordination/etcd"
	gormx "github.com/nostosx/servex-contrib/database/gorm"
	"github.com/nostosx/servex-contrib/jobqueue/asynq"
	"github.com/nostosx/servex-contrib/lifecycle"
	"github.com/nostosx/servex-contrib/session"
	sessionredis "github.com/nostosx/servex-contrib/session/redis"
	s3x "github.com/nostosx/servex-contrib/storage/s3"
	"github.com/nostosx/servex/app"
	"github.com/nostosx/servex/observability/metrics"
	"github.com/nostosx/servex/storage/cache"

	"github.com/nostosx/servex-examples/application/notes"
	"github.com/nostosx/servex-examples/infrastructure/config"
	"github.com/nostosx/servex-examples/infrastructure/store"
)

type dependencies struct {
	repository  notes.Repository
	sessions    session.Store
	states      oauth2.StateStore
	cache       cache.Cache
	queue       *asynq.Queue
	schedule    app.Hook
	objects     *s3x.Store
	coordinator *etcdx.Coordinator
}

func buildResources(ctx context.Context, cfg config.Composition, scope *lifecycle.Scope, meter metrics.Meter) (*dependencies, error) {
	result := &dependencies{repository: notes.NewMemoryRepository(), cache: cachememory.New()}
	sessions, err := session.NewMemoryStore(10000)
	if err != nil {
		return nil, err
	}
	result.sessions = sessions
	states, err := oauth2.NewMemoryStore(1000)
	if err != nil {
		return nil, err
	}
	result.states = states
	namespace := cfg.Namespace
	if namespace == "" {
		namespace = "servex-example:"
	}
	if cfg.PostgresDSN != "" {
		db, err := gormx.OpenPostgres("postgres", gormx.DefaultPostgresConfig(cfg.PostgresDSN), nil, meter)
		if err != nil {
			return nil, err
		}
		if err := scope.Add(db); err != nil {
			return nil, err
		}
		result.repository = store.NewNoteRepository(db.GORM())
	}
	if cfg.RedisAddress != "" {
		client := redis.NewClient(&redis.Options{Addr: cfg.RedisAddress, DialTimeout: 5 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second, ContextTimeoutEnabled: true, MaxRetries: -1})
		resource, err := lifecycle.New("redis", client, lifecycle.Operations[*redis.Client]{Check: func(ctx context.Context, c *redis.Client) error { return c.Ping(ctx).Err() }, Close: func(_ context.Context, c *redis.Client) error { return c.Close() }})
		if err != nil {
			return nil, err
		}
		if err := scope.Add(resource); err != nil {
			return nil, err
		}
		result.sessions, err = sessionredis.New(client, sessionredis.WithPrefix(namespace+"sessions:"))
		if err != nil {
			return nil, err
		}
		result.states, err = oauth2redis.New(client, namespace+"oauth-state:")
		if err != nil {
			return nil, err
		}
		result.cache = cacheredis.New(client, cacheredis.WithPrefix(namespace+"cache:"))
		queueName := namespace + "jobs"
		result.queue = asynq.NewFromClient(client, asynq.WithQueue(queueName), asynq.WithConcurrency(2))
		queueResource, err := lifecycle.New("queue", result.queue, lifecycle.Operations[*asynq.Queue]{Close: func(_ context.Context, q *asynq.Queue) error { return q.Close() }})
		if err != nil {
			return nil, err
		}
		if err := scope.Add(queueResource); err != nil {
			return nil, err
		}
		if cfg.JobSchedule != "" {
			result.schedule, err = asynq.NewScheduler(client, nil, []asynq.Schedule{{Spec: cfg.JobSchedule, Type: "example.refresh", Options: []asynqsdk.Option{asynqsdk.Queue(queueName), asynqsdk.MaxRetry(3)}}})
			if err != nil {
				return nil, err
			}
			// The schedule is an App consumer hook, not a resource owned by do or scope.
		}
	}
	if cfg.S3.Bucket != "" {
		region := cfg.S3.Region
		if region == "" {
			region = "us-east-1"
		}
		client := awss3.New(awss3.Options{Region: region, Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(cfg.S3.AccessKey, cfg.S3.SecretKey, "")), HTTPClient: &http.Client{Timeout: 10 * time.Second}, BaseEndpoint: optionalString(cfg.S3.Endpoint), UsePathStyle: cfg.S3.Endpoint != ""})
		result.objects, err = s3x.New(client, cfg.S3.Bucket, namespace+"objects/")
		if err != nil {
			return nil, err
		}
	}
	if len(cfg.EtcdEndpoints) > 0 {
		client, err := clientv3.New(clientv3.Config{Endpoints: cfg.EtcdEndpoints, DialTimeout: 5 * time.Second, Context: context.WithoutCancel(ctx)})
		if err != nil {
			return nil, err
		}
		resource, err := lifecycle.New("etcd", client, lifecycle.Operations[*clientv3.Client]{Check: func(ctx context.Context, c *clientv3.Client) error {
			_, err := c.Get(ctx, namespace+"health", clientv3.WithCountOnly())
			return err
		}, Close: func(_ context.Context, c *clientv3.Client) error { return c.Close() }})
		if err != nil {
			return nil, err
		}
		if err := scope.Add(resource); err != nil {
			return nil, err
		}
		result.coordinator, err = etcdx.New(client, "/"+strings.Trim(namespace, "/")+"/", 10)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
