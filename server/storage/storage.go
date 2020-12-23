package storage

import (
	"context"
	"fmt"

	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/logger"
	"github.com/saiya/dsps/server/storage/multiplex"
	"github.com/saiya/dsps/server/storage/onmemory"
	"github.com/saiya/dsps/server/storage/redis"
	"github.com/saiya/dsps/server/storage/tracing"
	"github.com/saiya/dsps/server/telemetry"
)

// NewStorage initialize Storage instance as per given config
func NewStorage(ctx context.Context, config *config.StoragesConfig, systemClock domain.SystemClock, channelProvider domain.ChannelProvider, telemetry *telemetry.Telemetry) (domain.Storage, error) {
	children := map[domain.StorageID]domain.Storage{}
	for id, subConfig := range *config {
		storage, err := newSubStorage(ctx, id, subConfig, systemClock, channelProvider, telemetry)
		if err != nil {
			return nil, fmt.Errorf("Failed to initialize storage \"%s\": %w", id, err)
		}
		children[id] = tracing.NewTracingStorage(storage, id, telemetry)
	}

	storage, err := multiplex.NewStorageMultiplexer(children)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize storage multiplexer: %w", err)
	}
	return tracing.NewTracingStorage(storage, "#root", telemetry), nil
}

func newSubStorage(ctx context.Context, id domain.StorageID, config *config.StorageConfig, systemClock domain.SystemClock, channelProvider domain.ChannelProvider, telemetry *telemetry.Telemetry) (domain.Storage, error) {
	if config.Onmemory != nil {
		logger.Of(ctx).Warnf(logger.CatStorage, "Starting onmemory storage \"%s\", ** DO NOT USE onmemory storage on production environment **", id)
		return onmemory.NewOnmemoryStorage(ctx, config.Onmemory, systemClock, channelProvider, telemetry)
	}
	if config.Redis != nil {
		logger.Of(ctx).Debugf(logger.CatStorage, "Starting Redis storage \"%s\"", id)
		return redis.NewRedisStorage(ctx, config.Redis, systemClock, channelProvider, telemetry)
	}
	return nil, xerrors.New("Empty storage configuration given")
}
