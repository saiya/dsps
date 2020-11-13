package config

// OnmemoryStorageConfig is definition of "storage.onmemory" configuration
type OnmemoryStorageConfig struct {
	DisablePubSub bool `json:"__disablePubSub"`
	DisableJwt    bool `json:"__disableJwt"`

	RunGCOnShutdown bool `json:"__runGcOnShutdown"`
}
