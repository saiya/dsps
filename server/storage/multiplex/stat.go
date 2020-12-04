package multiplex

import (
	"context"
	"fmt"
	"sync"

	"github.com/saiya/dsps/server/logger"
)

func (s *storageMultiplexer) Stat(ctx context.Context) (interface{}, error) {
	ch := make(chan childResult, len(s.children))
	wg := sync.WaitGroup{}
	for id, child := range s.children {
		wg.Add(1)

		id := id
		child := child
		go func() {
			defer wg.Done()

			stat, err := child.Stat(ctx)
			if err != nil {
				logger.Of(ctx).WarnError(fmt.Sprintf("Storage \"%s\" stat resulted in error", id), err)
				ch <- childResult{
					id: id,
					value: struct {
						Error string `json:"error"`
					}{Error: err.Error()},
				}
			} else {
				ch <- childResult{
					id:    id,
					value: stat,
				}
			}
		}()
	}
	wg.Wait()
	close(ch)

	snapshot := *s.stat
	return &storageMultiplexerStat{
		Multiplex: &snapshot,
		Children:  readAllChildResults(ch),
	}, nil
}
