package config

import (
	"fmt"

	"github.com/saiya/dsps/server/domain"
)

// StoragesConfig is list of storage configs
type StoragesConfig map[domain.StorageID]*StorageConfig

// StorageConfig is an item of "storage" configuration section
type StorageConfig struct {
	Onmemory *OnmemoryStorageConfig `json:"onmemory"`
	Redis    *RedisStorageConfig    `json:"redis"`
}

// DefaultStoragesConfig returns default configuration of storage backends
func DefaultStoragesConfig() StoragesConfig {
	return StoragesConfig{"default": &StorageConfig{Onmemory: &OnmemoryStorageConfig{}}}
}

// PostprocessStorageConfig fixup given configurations
func PostprocessStorageConfig(config *StoragesConfig) error {
	if len(*config) == 0 {
		*config = DefaultStoragesConfig()
		return nil
	}

	for id, s := range *config {
		types := 0
		if s.Onmemory != nil {
			types++
		}
		if s.Redis != nil {
			types++
			if err := postprocessRedisSubStorageConfig(s.Redis); err != nil {
				return fmt.Errorf("There is a configuration error on storage[%s].redis: %w", id, err)
			}
		}
		switch types {
		case 0:
			return fmt.Errorf("there is a configuration error on storage[%s]: no storage type under the item", id)
		case 1: // Correct
		default:
			return fmt.Errorf("there is a configuration error on storage[%s]: found multiple storage type under single item. To configure multiple storages, write separate storage definitions", id)
		}
	}
	return nil
}
