package app

import (
	"fmt"

	"github.com/jephir/tradeblocks"
)

// AccountBlocksMap maps block hashes to account blocks
type AccountBlocksMap map[string]*tradeblocks.AccountBlock

// AccountTokenHeadMap maps an 'account:token' string to their head block
type AccountTokenHeadMap map[string]*tradeblocks.AccountBlock

// AccountChangeListener is called whenever an account block is added or changed
type AccountChangeListener func(hash string, b *tradeblocks.AccountBlock)

// VoteBlocksMap maps block hashes to vote blocks
type VoteBlocksMap map[string]*tradeblocks.VoteBlock

// BlockStore stores all of the local blockchains
type BlockStore struct {
	AccountChangeListener AccountChangeListener
	AccountBlocks         AccountBlocksMap
	AccountHeads          AccountTokenHeadMap
	VoteBlocks            VoteBlocksMap
}

// NewBlockStore allocates and returns a new BlockStore
func NewBlockStore() *BlockStore {
	return &BlockStore{
		AccountBlocks: make(AccountBlocksMap),
		AccountHeads:  make(AccountTokenHeadMap),
		VoteBlocks:    make(VoteBlocksMap),
	}
}

// AddBlock verifies and adds the specified block to the store, and returns the hash of the added block
func (s *BlockStore) AddBlock(b *tradeblocks.AccountBlock) (string, error) {
	if err := ValidateAccountBlock(s, b); err != nil {
		return "", err
	}
	if err := s.checkConflict(b); err != nil {
		return "", err
	}
	h := b.Hash()
	s.AccountBlocks[h] = b
	s.AccountHeads[s.accountHeadKey(b.Account, b.Token)] = b
	if s.AccountChangeListener != nil {
		s.AccountChangeListener(h, b)
	}
	return h, nil
}

func (s *BlockStore) checkConflict(b *tradeblocks.AccountBlock) error {
	// open or issue case
	// TODO Fix
	if b.Previous == "" {
		for _, block := range s.AccountBlocks {
			if block.Previous == "" && block.Account == b.Account {
				return &BlockConflictError{block}
			}
		}
		return nil
	}
	for _, block := range s.AccountBlocks {
		if block.Previous == b.Previous {
			return &BlockConflictError{block}
		}
	}
	return nil
}

// GetBlock returns the account block with the specified hash, or nil if it doesn't exist
// error return added for future proofing
func (s *BlockStore) GetBlock(hash string) (*tradeblocks.AccountBlock, error) {
	return s.AccountBlocks[hash], nil
}

// GetHeadBlock returns the head block for the specified account-token blockchain
func (s *BlockStore) GetHeadBlock(account, token string) (*tradeblocks.AccountBlock, error) {
	return s.AccountHeads[s.accountHeadKey(account, token)], nil
}

func (s *BlockStore) accountHeadKey(account, token string) string {
	return account + ":" + token
}

// SwapBlocksMap maps block hashes to Swap blocks
type SwapBlocksMap map[string]*tradeblocks.SwapBlock

// SwapChangeListener is called whenever an Swap block is added or changed
type SwapChangeListener func(hash string, b *tradeblocks.SwapBlock)

// SwapBlockStore stores all of the local blockchains
type SwapBlockStore struct {
	SwapChangeListener SwapChangeListener
	SwapBlocks         SwapBlocksMap
}

// NewSwapBlockStore allocates and returns a new BlockStore
func NewSwapBlockStore() *SwapBlockStore {
	return &SwapBlockStore{
		SwapBlocks: make(SwapBlocksMap),
	}
}

// AddBlock verifies and adds the specified block to the store, and returns the hash of the added block
func (s *SwapBlockStore) AddBlock(b *tradeblocks.SwapBlock, c AccountBlockchain) (string, error) {
	if err := ValidateSwapBlock(c, s, b); err != nil {
		return "", err
	}
	if err := s.checkConflict(b); err != nil {
		return "", err
	}
	h := b.Hash()
	s.SwapBlocks[h] = b
	if s.SwapChangeListener != nil {
		s.SwapChangeListener(h, b)
	}
	return h, nil
}

func (s *SwapBlockStore) checkConflict(b *tradeblocks.SwapBlock) error {
	if b.Previous == "" {
		return nil
	}
	for _, block := range s.SwapBlocks {
		if block.Previous == b.Previous {
			return &swapConflictError{block}
		}
	}
	return nil
}

// GetSwapBlock returns the Swap block with the specified hash, or nil if it doesn't exist
// error return added for future proofing
func (s *SwapBlockStore) GetSwapBlock(hash string) (*tradeblocks.SwapBlock, error) {
	return s.SwapBlocks[hash], nil
}

// OrderBlocksMap maps block hashes to Order blocks
type OrderBlocksMap map[string]*tradeblocks.OrderBlock

// OrderChangeListener is called whenever an Order block is added or changed
type OrderChangeListener func(hash string, b *tradeblocks.OrderBlock)

// OrderBlockStore stores all of the local blockchains
type OrderBlockStore struct {
	OrderChangeListener OrderChangeListener
	OrderBlocks         OrderBlocksMap
}

// NewOrderBlockStore allocates and returns a new BlockStore
func NewOrderBlockStore() *OrderBlockStore {
	return &OrderBlockStore{
		OrderBlocks: make(OrderBlocksMap),
	}
}

// AddBlock verifies and adds the specified block to the store, and returns the hash of the added block
func (o *OrderBlockStore) AddBlock(b *tradeblocks.OrderBlock, s SwapBlockchain, c AccountBlockchain) (string, error) {
	if err := ValidateOrderBlock(c, s, o, b); err != nil {
		return "", err
	}
	if err := o.checkConflict(b); err != nil {
		return "", err
	}
	h := b.Hash()
	o.OrderBlocks[h] = b
	if o.OrderChangeListener != nil {
		o.OrderChangeListener(h, b)
	}
	return h, nil
}

func (o *OrderBlockStore) checkConflict(b *tradeblocks.OrderBlock) error {
	if b.Previous == "" {
		return nil
	}
	for _, block := range o.OrderBlocks {
		if block.Previous == b.Previous {
			return &orderConflictError{block}
		}
	}
	return nil
}

// GetOrderBlock returns the Order block with the specified hash, or nil if it doesn't exist
// error return added for future proofing
func (o *OrderBlockStore) GetOrderBlock(hash string) (*tradeblocks.OrderBlock, error) {
	return o.OrderBlocks[hash], nil
}

// BlockConflictError represents a conflict (multiple parent claim)
type BlockConflictError struct {
	existing *tradeblocks.AccountBlock
}

func (e *BlockConflictError) Error() string {
	return fmt.Sprintf("blockstore: conflict with existing block '%s'", e.existing.Hash())
}

type swapConflictError struct {
	existing *tradeblocks.SwapBlock
}

func (e *swapConflictError) Error() string {
	return fmt.Sprintf("blockstore: conflict with existing block '%s'", e.existing.Hash())
}

type orderConflictError struct {
	existing *tradeblocks.OrderBlock
}

func (e *orderConflictError) Error() string {
	return fmt.Sprintf("blockstore: conflict with existing block '%s'", e.existing.Hash())
}
