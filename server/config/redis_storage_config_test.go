package config_test

import (
	"context"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/config"
	. "github.com/saiya/dsps/server/testing"
)

func TestRedisMissingAddrs(t *testing.T) {
	configYaml := strings.ReplaceAll(`
storages:
	myRedis:
		redis:
			username: "user"
`, "\t", "  ")
	_, err := ParseConfig(context.Background(), Overrides{}, configYaml)
	assert.EqualError(t, err, "Storage configration problem: There is a configuration error on storage[myRedis].redis: Redis configration must have one of 'singleNode' and 'cluster' item")
}

func TestRedisAmbiguousAddrs(t *testing.T) {
	configYaml := strings.ReplaceAll(`
storages:
	myRedis:
		redis:
			singleNode: 'localhost:6379'
			cluster:
				- 'a-node-of-cluster-1:6379'
				- 'another-node-of-cluster-1:6379'
`, "\t", "  ")
	_, err := ParseConfig(context.Background(), Overrides{}, configYaml)
	assert.EqualError(t, err, "Storage configration problem: There is a configuration error on storage[myRedis].redis: Redis configration can have ONLY ONE of 'singleNode' and 'cluster' item, cannot specify both")
}

func TestRedisInvalidConfig(t *testing.T) {
	configYaml := strings.ReplaceAll(`
storages:
	myRedis:
		redis:
			singleNode: 'localhost:6379'
			db: -1
`, "\t", "  ")
	_, err := ParseConfig(context.Background(), Overrides{}, configYaml)
	assert.Contains(t, err.Error(), "Field validation for 'DBNumber' failed")
}

func TestRedisDefaultValues(t *testing.T) {
	configYaml := strings.ReplaceAll(`
storages:
	myRedis:
		redis:
			singleNode: 'localhost:6379'
`, "\t", "  ")
	config, err := ParseConfig(context.Background(), Overrides{}, configYaml)
	if err != nil {
		t.Error(err)
		return
	}

	if config.Storages["myRedis"].Redis == nil {
		t.Errorf("config.Storage.Redis missing")
		return
	}
	cfg := *config.Storages["myRedis"].Redis

	assert.Equal(t, MakeDurationPtr("5m"), cfg.ScriptReloadInterval)

	assert.Equal(t, MakeDurationPtr("5s"), cfg.Timeout.Connect)
	assert.Equal(t, MakeDurationPtr("5s"), cfg.Timeout.Read)
	assert.Equal(t, MakeDurationPtr("5s"), cfg.Timeout.Write)

	assert.Equal(t, MakeIntPtr(3), cfg.Retry.Count)
	assert.Equal(t, MakeDurationPtr("500ms"), cfg.Retry.Interval)
	assert.Equal(t, MakeDurationPtr("200ms"), cfg.Retry.IntervalJitter)

	assert.Equal(t, MakeIntPtr(1024), cfg.Connection.Max)
	assert.Equal(t, MakeIntPtr(runtime.NumCPU()*16), cfg.Connection.Min)
	assert.Equal(t, MakeDurationPtr("5m"), cfg.Connection.MaxIdleTime)
}

func TestRedisCustomValues(t *testing.T) {
	configYaml := strings.ReplaceAll(`
storages:
	myRedis:
		redis:
			singleNode: 'localhost:6379'
			timeout:
				connect: 1s500ms
				read: 3s
				write: 7s
			retry:
				count: 6
				interval: 321ms
				intervalJitter: 32ms
			connection:
				max: 521
				min: 256
				maxIdleTime: 90m
`, "\t", "  ")
	config, err := ParseConfig(context.Background(), Overrides{}, configYaml)
	if err != nil {
		t.Error(err)
		return
	}

	if config.Storages["myRedis"].Redis == nil {
		t.Errorf("config.Storage.Redis missing")
		return
	}
	cfg := *config.Storages["myRedis"].Redis

	assert.Equal(t, MakeDurationPtr("1s500ms"), cfg.Timeout.Connect)
	assert.Equal(t, MakeDurationPtr("3s"), cfg.Timeout.Read)
	assert.Equal(t, MakeDurationPtr("7s"), cfg.Timeout.Write)

	assert.Equal(t, MakeIntPtr(6), cfg.Retry.Count)
	assert.Equal(t, MakeDurationPtr("321ms"), cfg.Retry.Interval)
	assert.Equal(t, MakeDurationPtr("32ms"), cfg.Retry.IntervalJitter)

	assert.Equal(t, MakeIntPtr(521), cfg.Connection.Max)
	assert.Equal(t, MakeIntPtr(256), cfg.Connection.Min)
	assert.Equal(t, MakeDurationPtr("90m"), cfg.Connection.MaxIdleTime)
}
