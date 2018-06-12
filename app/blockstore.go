package app

import (
	"sync"

	"github.com/jephir/tradeblocks"
)

// Blocks maps hashes to blocks
type Blocks map[string]tradeblocks.Block

// AccountBlocks maps a context-specific identifier to an account block
type AccountBlocks map[string]*tradeblocks.AccountBlock

// SwapBlocks maps a context-specific identifier to a swap block
type SwapBlocks map[string]*tradeblocks.SwapBlock

// OrderBlocks maps a context-specific identifier to an order block
type OrderBlocks map[string]*tradeblocks.OrderBlock

// BlockStore is a concurrency-safe block store
type BlockStore struct {
	mu sync.RWMutex

	// Keyed by hash
	blocks        Blocks
	blockSequence map[string]int
	sequence      int

	// Keyed by hash
	accountBlocks AccountBlocks
	// Keyed by account:token
	accountHeads AccountBlocks

	// Keyed by hash
	swapBlocks SwapBlocks
	// Keyed by account:id
	swapHeads SwapBlocks

	// Keyed by hash
	orderBlocks OrderBlocks
	// Keyed by account:id
	orderHeads OrderBlocks
}

// NewBlockStore allocates and returns a new BlockStore
func NewBlockStore() *BlockStore {
	return &BlockStore{
		blocks:        make(Blocks),
		blockSequence: make(map[string]int),
		accountBlocks: make(AccountBlocks),
		accountHeads:  make(AccountBlocks),
		swapBlocks:    make(SwapBlocks),
		swapHeads:     make(SwapBlocks),
		orderBlocks:   make(OrderBlocks),
		orderHeads:    make(OrderBlocks),
	}
}

// AddAccountBlock verifies and adds the specified account block to this store
func (s *BlockStore) AddAccountBlock(b *tradeblocks.AccountBlock) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ValidateAccountBlock(s, b); err != nil {
		return err
	}
	h := b.Hash()
	s.blocks[h] = b
	s.accountBlocks[h] = b
	s.accountHeads[accountHeadKey(b.Account, b.Token)] = b
	s.setBlockSequence(h)
	return nil
}

// AddSwapBlock verifies and adds the specified swap block to this store
func (s *BlockStore) AddSwapBlock(b *tradeblocks.SwapBlock) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// if err := ValidateSwapBlock(s, b); err != nil {
	// 	return err
	// }
	h := b.Hash()
	s.blocks[h] = b
	s.swapBlocks[h] = b
	s.swapHeads[swapHeadKey(b.Account, b.ID)] = b
	s.setBlockSequence(h)
	return nil
}

// AddOrderBlock verifies and adds the specified order block to this store
func (s *BlockStore) AddOrderBlock(b *tradeblocks.OrderBlock) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// if err := ValidateOrderBlock(s, b); err != nil {
	// 	return err
	// }
	h := b.Hash()
	s.blocks[h] = b
	s.orderBlocks[h] = b
	s.orderHeads[orderHeadKey(b.Account, b.ID)] = b
	s.setBlockSequence(h)
	return nil
}

func (s *BlockStore) setBlockSequence(hash string) {
	s.blockSequence[hash] = s.sequence
	s.sequence++
}

// Sequence returns the sequence number of the block with the specified hash
func (s *BlockStore) Sequence(hash string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.blockSequence[hash]
}

// SequenceLess returns true if the sequence number of block i is less than block j
func (s *BlockStore) SequenceLess(i, j string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.blockSequence[i] < s.blockSequence[j]
}

// AccountBlocks calls the specified function with every block in this store. Return false to stop iteration.
func (s *BlockStore) AccountBlocks(f func(sequence int, b *tradeblocks.AccountBlock) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for h, b := range s.accountBlocks {
		seq := s.blockSequence[h]
		if !f(seq, b) {
			return
		}
	}
}

// SwapBlocks calls the specified function with every block in this store. Return false to stop iteration.
func (s *BlockStore) SwapBlocks(f func(sequence int, b *tradeblocks.SwapBlock) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for h, b := range s.swapBlocks {
		seq := s.blockSequence[h]
		if !f(seq, b) {
			return
		}
	}
}

// OrderBlocks calls the specified function with every block in this store. Return false to stop iteration.
func (s *BlockStore) OrderBlocks(f func(sequence int, b *tradeblocks.OrderBlock) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for h, b := range s.orderBlocks {
		seq := s.blockSequence[h]
		if !f(seq, b) {
			return
		}
	}
}

// Blocks calls the specified function with every block in this store. Return false to stop iteration.
func (s *BlockStore) Blocks(f func(sequence int, b tradeblocks.Block) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for h, b := range s.blocks {
		seq := s.blockSequence[h]
		if !f(seq, b) {
			return
		}
	}
}

// Block returns the block with the specified hash or nil if it's not found
func (s *BlockStore) Block(hash string) tradeblocks.Block {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.blocks[hash]
}

// GetAccountBlock returns the account block for the specified hash or nil if it's not found
func (s *BlockStore) GetAccountBlock(hash string) *tradeblocks.AccountBlock {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.accountBlocks[hash]
}

// GetSwapBlock returns the swap block for the specified hash or nil if it's not found
func (s *BlockStore) GetSwapBlock(hash string) *tradeblocks.SwapBlock {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.swapBlocks[hash]
}

// GetOrderBlock returns the order block for the specified hash or nil if it's not found
func (s *BlockStore) GetOrderBlock(hash string) *tradeblocks.OrderBlock {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.orderBlocks[hash]
}

// GetAccountHead returns the head block for the specified account-token pair
func (s *BlockStore) GetAccountHead(account, token string) *tradeblocks.AccountBlock {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.accountHeads[accountHeadKey(account, token)]
}

// GetSwapHead returns the head block for the specified account-id pair
func (s *BlockStore) GetSwapHead(account, id string) *tradeblocks.SwapBlock {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.swapHeads[swapHeadKey(account, id)]
}

// GetOrderHead returns the head block for the specified account-id pair
func (s *BlockStore) GetOrderHead(account, id string) *tradeblocks.OrderBlock {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.orderHeads[orderHeadKey(account, id)]
}

func accountHeadKey(account, token string) string {
	return account + ":" + token
}

func swapHeadKey(account, id string) string {
	return account + ":" + id
}

func orderHeadKey(account, id string) string {
	return account + ":" + id
}
