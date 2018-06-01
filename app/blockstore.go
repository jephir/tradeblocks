package app

import (
	"github.com/jephir/tradeblocks"
)

// AccountBlocksMap maps block hashes to account blocks
type AccountBlocksMap map[string]*tradeblocks.AccountBlock

// AccountChangeListener is called whenever an account block is added or changed
type AccountChangeListener func(hash string, b *tradeblocks.AccountBlock)

// BlockStore stores all of the local blockchains
type BlockStore struct {
	AccountChangeListener AccountChangeListener
	AccountBlocks         AccountBlocksMap
}

// NewBlockStore allocates and returns a new BlockStore
func NewBlockStore() *BlockStore {
	return &BlockStore{
		AccountBlocks: make(AccountBlocksMap),
	}
}

// AddBlock verifies and adds the specified block to the store, and returns the hash of the added block
func (s *BlockStore) AddBlock(b *tradeblocks.AccountBlock) (string, error) {
	err := ValidateAccountBlock(s, b)
	if err != nil {
		return "", err
	}
	h := b.Hash()
	s.AccountBlocks[h] = b
	if s.AccountChangeListener != nil {
		s.AccountChangeListener(h, b)
	}
	return h, nil
}

// GetBlock returns the account block with the specified hash, or nil if it doesn't exist
// error return added for future proofing
func (s *BlockStore) GetBlock(hash string) (*tradeblocks.AccountBlock, error) {
	return s.AccountBlocks[hash], nil
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
	err := ValidateSwapBlock(c, s, b)
	if err != nil {
		return "", err
	}
	h := b.Hash()
	s.SwapBlocks[h] = b
	if s.SwapChangeListener != nil {
		s.SwapChangeListener(h, b)
	}
	return h, nil
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
func (s *OrderBlockStore) AddBlock(b *tradeblocks.OrderBlock, c AccountBlockchain) (string, error) {
	err := ValidateOrderBlock(c, s, b)
	if err != nil {
		return "", err
	}
	h := b.Hash()
	s.OrderBlocks[h] = b
	if s.OrderChangeListener != nil {
		s.OrderChangeListener(h, b)
	}
	return h, nil
}

// GetOrderBlock returns the Order block with the specified hash, or nil if it doesn't exist
// error return added for future proofing
func (s *OrderBlockStore) GetOrderBlock(hash string) (*tradeblocks.OrderBlock, error) {
	return s.OrderBlocks[hash], nil
}
