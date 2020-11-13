package config

import (
	"runtime"

	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/domain"
)

// RedisStorageConfig is definition of "storage.redis" configuration
type RedisStorageConfig struct {
	SingleNode *string   `json:"singleNode"`
	Cluster    *[]string `json:"cluster"`

	DisablePubSub bool `json:"disablePubSub"`
	DisableJwt    bool `json:"disableJwt"`

	Username string `json:"username"`
	Password string `json:"password"`
	DBNumber int    `json:"db" validate:"min=0"`

	Timeout struct {
		Connect *domain.Duration `json:"connect"`
		Read    *domain.Duration `json:"read"`
		Write   *domain.Duration `json:"write"`
	} `json:"timeout"`

	Retry struct {
		Count          *int             `json:"count"`
		Interval       *domain.Duration `json:"interval"`
		IntervalJitter *domain.Duration `json:"intervalJitter"`
	} `json:"retry"`

	Connection struct {
		Max         *int             `json:"max"`
		Min         *int             `json:"min"`
		MaxIdleTime *domain.Duration `json:"maxIdleTime"`
	} `json:"connection"`
}

// IsSingleNode returns true only for single-node Redis
func (config RedisStorageConfig) IsSingleNode() bool {
	return config.SingleNode != nil && len(*config.SingleNode) > 0
}

// IsCluster returns true only for clustered Redis
func (config RedisStorageConfig) IsCluster() bool {
	return config.Cluster != nil && len(*config.Cluster) > 0
}

func postprocessRedisSubStorageConfig(config *RedisStorageConfig) error {
	if config.IsSingleNode() && config.IsCluster() {
		return xerrors.New("Redis configration can have ONLY ONE of 'singleNode' and 'cluster' item, cannot specify both")
	}
	if !(config.IsSingleNode() || config.IsCluster()) {
		return xerrors.New("Redis configration must have one of 'singleNode' and 'cluster' item")
	}

	if config.Timeout.Connect == nil {
		config.Timeout.Connect = makeDurationPtr("5s")
	}
	if config.Timeout.Read == nil {
		config.Timeout.Read = makeDurationPtr("5s")
	}
	if config.Timeout.Write == nil {
		config.Timeout.Write = makeDurationPtr("5s")
	}

	if config.Retry.Count == nil {
		config.Retry.Count = makeIntPtr(3)
	}
	if config.Retry.Interval == nil {
		config.Retry.Interval = makeDurationPtr("500ms")
	}
	if config.Retry.IntervalJitter == nil {
		config.Retry.IntervalJitter = makeDurationPtr("200ms")
	}

	if config.Connection.Max == nil {
		config.Connection.Max = makeIntPtr(runtime.NumCPU() * 64)
		if *config.Connection.Max < 1024 {
			config.Connection.Max = makeIntPtr(1024)
		}
	}
	if config.Connection.Min == nil {
		config.Connection.Min = makeIntPtr(runtime.NumCPU() * 16)
	}
	if config.Connection.MaxIdleTime == nil {
		config.Connection.MaxIdleTime = makeDurationPtr("5m")
	}

	return nil
}
