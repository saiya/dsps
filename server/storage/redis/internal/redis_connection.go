package internal

import (
	"context"

	"github.com/go-redis/redis/extra/redisotel"
	"github.com/go-redis/redis/v8"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/logger"
)

// RedisConnection represents Redis connection system
type RedisConnection struct {
	RedisCmd RedisCmd
	Close    func() error

	IsSingleNode bool
	IsCluster    bool

	MaxConnections int
}

// NewRedisConnection establish connection pool to Redis server.
func NewRedisConnection(ctx context.Context, config *config.RedisStorageConfig) (RedisConnection, error) {
	var conn RedisConnection
	if config.SingleNode != nil {
		conn = createClientSingleNode(ctx, config)
	} else {
		conn = createClientCluster(ctx, config)
	}
	if err := conn.RedisCmd.Ping(ctx); err != nil {
		if err := conn.Close(); err != nil {
			logger.Of(ctx).InfoError(logger.CatStorage, "Failed to close Redis connection after initial ping failure", err)
		}
		return RedisConnection{}, err
	}
	return conn, nil
}

func createClientSingleNode(ctx context.Context, config *config.RedisStorageConfig) RedisConnection {
	c := redis.NewClient(&redis.Options{
		Addr: *config.SingleNode,

		DB:       config.DBNumber,
		Username: config.Username,
		Password: config.Password,

		DialTimeout:  config.Timeout.Connect.Duration,
		ReadTimeout:  config.Timeout.Read.Duration,
		WriteTimeout: config.Timeout.Write.Duration,

		MaxRetries:      *config.Retry.Count,
		MinRetryBackoff: config.Retry.Interval.Duration - config.Retry.IntervalJitter.Duration,
		MaxRetryBackoff: config.Retry.Interval.Duration + config.Retry.IntervalJitter.Duration,

		MinIdleConns: *config.Connection.Min,
		PoolSize:     *config.Connection.Max,
		IdleTimeout:  config.Connection.MaxIdleTime.Duration,
	})
	c.AddHook(redisotel.TracingHook{})
	return RedisConnection{
		RedisCmd: NewRedisCmd(c, func(ctx context.Context, channel RedisChannelID) *redis.PubSub {
			return c.PSubscribe(ctx, string(channel))
		}),
		Close: func() error {
			return c.Close()
		},
		IsSingleNode:   true,
		MaxConnections: *config.Connection.Max,
	}
}

func createClientCluster(ctx context.Context, config *config.RedisStorageConfig) RedisConnection {
	c := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: *config.Cluster,

		Username: config.Username,
		Password: config.Password,

		DialTimeout:  config.Timeout.Connect.Duration,
		ReadTimeout:  config.Timeout.Read.Duration,
		WriteTimeout: config.Timeout.Write.Duration,

		MaxRetries:      *config.Retry.Count,
		MinRetryBackoff: config.Retry.Interval.Duration - config.Retry.IntervalJitter.Duration,
		MaxRetryBackoff: config.Retry.Interval.Duration + config.Retry.IntervalJitter.Duration,

		MinIdleConns: *config.Connection.Min,
		PoolSize:     *config.Connection.Max,
		IdleTimeout:  config.Connection.MaxIdleTime.Duration,

		// --- Cluster specific config items ---
		MaxRedirects: 3,
		ReadOnly:     false,
	})
	c.AddHook(redisotel.TracingHook{})
	return RedisConnection{
		RedisCmd: NewRedisCmd(c, func(ctx context.Context, channel RedisChannelID) *redis.PubSub {
			return c.PSubscribe(ctx, string(channel))
		}),
		Close: func() error {
			return c.Close()
		},
		IsCluster:      true,
		MaxConnections: *config.Connection.Max,
	}
}
