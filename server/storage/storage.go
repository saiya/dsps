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
)

// NewStorage initialize Storage instance as per given config
func NewStorage(ctx context.Context, config *config.StoragesConfig, systemClock domain.SystemClock, channelProvider domain.ChannelProvider) (domain.Storage, error) {
	children := map[domain.StorageID]domain.Storage{}
	for id, subConfig := range *config {
		storage, err := newSubStorage(ctx, id, subConfig, systemClock, channelProvider)
		if err != nil {
			return nil, fmt.Errorf("Failed to initialize storage \"%s\": %w", id, err)
		}
		children[id] = tracing.NewTracingStorage(storage, id)
	}

	storage, err := multiplex.NewStorageMultiplexer(children)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize storage multiplexer: %w", err)
	}
	return tracing.NewTracingStorage(storage, "#root"), nil
}

func newSubStorage(ctx context.Context, id domain.StorageID, config *config.StorageConfig, systemClock domain.SystemClock, channelProvider domain.ChannelProvider) (domain.Storage, error) {
	if config.Onmemory != nil {
		logger.Of(ctx).Warnf(logger.CatStorage, "Starting onmemory storage \"%s\", ** DO NOT USE onmemory storage on production environment **", id)
		return onmemory.NewOnmemoryStorage(ctx, config.Onmemory, systemClock, channelProvider)
	}
	if config.Redis != nil {
		logger.Of(ctx).Debugf(logger.CatStorage, "Starting Redis storage \"%s\"", id)
		return redis.NewRedisStorage(ctx, config.Redis, systemClock, channelProvider)
	}
	return nil, xerrors.New("Empty storage configuration given")
}
