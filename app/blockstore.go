package app

import (
	"github.com/jephir/tradeblocks"
)

// Keyed by hash
type accountBlocksMap map[string]*tradeblocks.AccountBlock

// BlockStore stores all of the local blockchains
type BlockStore struct {
	accountBlocks accountBlocksMap
}

// NewBlockStore allocates and returns a new BlockStore
func NewBlockStore() *BlockStore {
	return &BlockStore{
		accountBlocks: make(accountBlocksMap),
	}
}

// AddBlock verifies and adds the specified block to the store
func (s *BlockStore) AddBlock(b *tradeblocks.AccountBlock) error {
	// TODO Validate block
	// err := s.validator.ValidateAccountBlock(b)
	hash, err := AccountBlockHash(b)
	if err != nil {
		return err
	}
	s.accountBlocks[hash] = b
	return nil
}

// GetBlock returns the account block with the specified hash, or nil if it doesn't exist
func (s *BlockStore) GetBlock(hash string) *tradeblocks.AccountBlock {
	return s.accountBlocks[hash]
}
