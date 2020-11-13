package multiplex

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/saiya/dsps/server/domain"
)

type childResult struct {
	id    domain.StorageID
	value interface{}
}

// Note: channel must be closed.
func readAllChildResults(c chan childResult) map[domain.StorageID]interface{} {
	results := make(map[domain.StorageID]interface{})
	for result := range c {
		results[result.id] = result.value
	}
	return results
}

var errMultiplexSkipped = errors.New("storage.multiplex.skipped")

// Concurrently call given func for all children (even if one or more failed), then returns nil (success) if one or more succeeded
func (s *storageMultiplexer) parallelAtLeastOneSuccess(ctx context.Context, operationName string, f func(ctx context.Context, id domain.StorageID, s domain.Storage) (interface{}, error)) (map[domain.StorageID]interface{}, error) {
	wg := sync.WaitGroup{}
	successCh := make(chan childResult, len(s.children))
	errCh := make(chan childResult, len(s.children))
	for id, child := range s.children {
		wg.Add(1)
		id := id
		child := child

		go func() {
			defer wg.Done()
			result, err := f(ctx, id, child)
			if err != nil {
				errCh <- childResult{
					id:    id,
					value: err,
				}
			} else {
				successCh <- childResult{
					id:    id,
					value: result,
				}
			}
		}()
	}
	wg.Wait()
	close(successCh)
	close(errCh)

	result := readAllChildResults(successCh)

	var firstErr error
	for err := range errCh {
		id := err.id
		err := fmt.Errorf("%s failed on storage \"%s\": %w", operationName, id, err.value.(error))
		if errors.Is(err, errMultiplexSkipped) {
			continue
		}
		if firstErr == nil || domain.IsStorageNonFatalError(err) { // Report non-fatal error (business error) rather than fatal errors
			firstErr = err
		}
		if !domain.IsStorageNonFatalError(err) {
			fmt.Printf("%v\n", err) // TODO: Use logger
		}
	}
	if len(result) == 0 && firstErr != nil {
		return nil, firstErr
	}
	return result, nil
}
