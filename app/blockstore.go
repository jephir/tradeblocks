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
