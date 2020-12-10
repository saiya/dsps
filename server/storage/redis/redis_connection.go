package redis

import (
	"context"

	"github.com/go-redis/redis/v8"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/logger"
)

type redisConnection struct {
	redisCmd redisCmd
	close    func() error

	isSingleNode bool
	isCluster    bool

	maxConnections int
}

func connect(ctx context.Context, config *config.RedisStorageConfig) (redisConnection, error) {
	var conn redisConnection
	if config.SingleNode != nil {
		conn = createClientSingleNode(ctx, config)
	} else {
		conn = createClientCluster(ctx, config)
	}
	if err := conn.redisCmd.Ping(ctx); err != nil {
		if err := conn.close(); err != nil {
			logger.Of(ctx).InfoError("Failed to close Redis connection after initial ping failure", err)
		}
		return redisConnection{}, err
	}
	return conn, nil
}

func createClientSingleNode(ctx context.Context, config *config.RedisStorageConfig) redisConnection {
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
	return redisConnection{
		redisCmd: newRedisCmd(c, func(ctx context.Context, channel string) *redis.PubSub {
			return c.Subscribe(ctx, channel)
		}),
		close: func() error {
			return c.Close()
		},
		isSingleNode:   true,
		maxConnections: *config.Connection.Max,
	}
}

func createClientCluster(ctx context.Context, config *config.RedisStorageConfig) redisConnection {
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
	return redisConnection{
		redisCmd: newRedisCmd(c, func(ctx context.Context, channel string) *redis.PubSub {
			return c.Subscribe(ctx, channel)
		}),
		close: func() error {
			return c.Close()
		},
		isCluster:      true,
		maxConnections: *config.Connection.Max,
	}
}
