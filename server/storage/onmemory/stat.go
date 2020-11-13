package onmemory

import (
	"context"

	"github.com/saiya/dsps/server/domain"
)

type onmemoryStorageStat struct {
	GC struct {
		TotalCount  int64       `json:"totalCount"`
		LastStartAt domain.Time `json:"lastStartAt"`
		LastGCSec   float64     `json:"lastGcSec"`

		Evicted struct {
			Subscribers    int64 `json:"subscribers"`
			Messages       int64 `json:"messages"`
			JwtRevocations int64 `json:"jwtRevocations"`
		} `json:"evicted"`
	} `json:"gc"`
}

func (s *onmemoryStorage) Stat(ctx context.Context) (interface{}, error) {
	snapshot := *s.stat
	return &snapshot, nil
}
