package multiplex

import (
	"context"
	"fmt"
	"sync"

	"github.com/saiya/dsps/server/domain"
	storageinternal "github.com/saiya/dsps/server/storage/internal"
)

func (s *storageMultiplexer) PublishMessages(ctx context.Context, msgs []domain.Message) error {
	_, err := s.parallelAtLeastOneSuccess(ctx, "PublishMessages", func(ctx context.Context, _ domain.StorageID, child domain.Storage) (interface{}, error) {
		if child := child.AsPubSubStorage(); child != nil {
			return nil, child.PublishMessages(ctx, msgs)
		}
		return nil, errMultiplexSkipped
	})
	return err
}

func (s *storageMultiplexer) FetchMessages(ctx context.Context, sl domain.SubscriberLocator, max int, waituntil domain.Duration) (messages []domain.Message, moreMessages bool, ackHandle domain.AckHandle, err error) {
	type fetchResult struct {
		msgs         []domain.Message
		moreMessages bool
		ackHandle    domain.AckHandle
	}
	results, err := s.parallelAtLeastOneSuccess(ctx, "FetchMessages", func(ctx context.Context, _ domain.StorageID, child domain.Storage) (interface{}, error) {
		if child := child.AsPubSubStorage(); child != nil {
			msgs, moreMsgs, ackHandle, err := child.FetchMessages(ctx, sl, max, waituntil)
			if err != nil {
				return nil, err
			}
			return fetchResult{msgs: msgs, moreMessages: moreMsgs, ackHandle: ackHandle}, nil
		}
		return nil, errMultiplexSkipped
	})
	if err != nil {
		return nil, false, domain.AckHandle{}, err
	}

	// Note that this merge logic honors message ordering as possible.
	// Only exception is that storages returns messages by different ordering, possible cause of such case is that client retry to publish messages.
	// If client retried publish, no need to guarantee ordering of the messages sent concurrently with the retry.
	msgRedudancies := map[domain.MessageLocator]int{}
	msgsFromChildren := make([]domain.Message, 0, max)
	ackHandles := map[domain.StorageID]domain.AckHandle{}
	for storageID, result := range results {
		result := result.(fetchResult)
		moreMessages = moreMessages || result.moreMessages
		for _, msg := range result.msgs {
			redundancy := msgRedudancies[msg.MessageLocator] + 1
			msgRedudancies[msg.MessageLocator] = redundancy
			if redundancy == 1 {
				msgsFromChildren = append(msgsFromChildren, msg)
			}
		}
		if len(result.msgs) != 0 { // If zero, the ackHandle is not valid
			ackHandles[storageID] = result.ackHandle
		}
	}

	staleSuspectMsgs := make([]domain.MessageLocator, 0, len(msgsFromChildren))
	for _, msg := range msgsFromChildren {
		if msgRedudancies[msg.MessageLocator] != len(results) {
			// The message might be old (acknowledged) message in some storages; should ignore the message if it already acknowledged.
			staleSuspectMsgs = append(staleSuspectMsgs, msg.MessageLocator)
		}
	}
	staleMegs, err := s.IsOldMessages(ctx, sl, staleSuspectMsgs)
	if err != nil {
		return nil, false, domain.AckHandle{}, err
	}
	messages = make([]domain.Message, 0, len(msgsFromChildren))
	for _, msg := range msgsFromChildren {
		if !staleMegs[msg.MessageLocator] {
			messages = append(messages, msg)
		}
	}

	ackHandle, err = storageinternal.EncodeMultiplexAckHandle(ackHandles)
	if err != nil {
		return nil, false, domain.AckHandle{}, err
	}
	return
}

func (s *storageMultiplexer) AcknowledgeMessages(ctx context.Context, handle domain.AckHandle) error {
	h, err := storageinternal.DecodeMultiplexAckHandle(handle)
	if err != nil {
		return err
	}
	_, err = s.parallelAtLeastOneSuccess(ctx, "AcknowledgeMessages", func(ctx context.Context, id domain.StorageID, child domain.Storage) (interface{}, error) {
		if child := child.AsPubSubStorage(); child != nil {
			if handle, ok := h[id]; ok {
				return nil, child.AcknowledgeMessages(ctx, handle)
			}
			return nil, errMultiplexSkipped // This storage added after creation of the handle
		}
		return nil, errMultiplexSkipped
	})
	return err
}

// This method does not return error even if all storage backend returns error (consistent with what storageMultiplexer.FetchMessages does).
// Because Storage.IsOldMessages can return false for "unsure" messages, it is okay to return false when storage error occurs.
func (s *storageMultiplexer) IsOldMessages(ctx context.Context, sl domain.SubscriberLocator, msgs []domain.MessageLocator) (map[domain.MessageLocator]bool, error) {
	if len(msgs) == 0 { // optimization for storageMultiplexer.FetchMessages
		return map[domain.MessageLocator]bool{}, nil
	}

	ch := make(chan map[domain.MessageLocator]bool, len(s.children))
	wg := sync.WaitGroup{}
	for id, child := range s.children {
		id := id
		if child := child.AsPubSubStorage(); child != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				staleMap, err := child.IsOldMessages(ctx, sl, msgs)
				if err != nil {
					if !domain.IsStorageNonFatalError(err) {
						fmt.Printf("IsOldMessages of \"%s\" failed: %v\n", id, err) // TODO: Use logger
					}
					return
				}
				ch <- staleMap
			}()
		}
	}
	wg.Wait()
	close(ch)

	result := map[domain.MessageLocator]bool{}
	for _, msgLoc := range msgs {
		result[msgLoc] = false
	}
	for m := range ch {
		for msgLoc, isOld := range m {
			if isOld {
				result[msgLoc] = true
			}
		}
	}
	return result, nil
}
