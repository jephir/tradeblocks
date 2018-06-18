package app

import (
	"crypto/rand"
	"fmt"

	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/db"
)

// BlockStore is a concurrency-safe block store
type BlockStore struct {
	db  *db.DB
	err error
}

// NewBlockStore allocates and returns a new BlockStore
func NewBlockStore() *BlockStore {
	x := randString(16)
	s := fmt.Sprintf("file:%s?mode=memory&cache=shared&_foreign_keys=true", x)
	db, err := db.NewDB(s)
	if err != nil {
		panic(err)
	}
	return &BlockStore{
		db: db,
	}
}

// https://stackoverflow.com/a/12772666
func randString(n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

// Err returns the current error or nil
func (s *BlockStore) Err() error {
	return s.err
}

// AddAccountBlock verifies and adds the specified account block to this store
func (s *BlockStore) AddAccountBlock(b *tradeblocks.AccountBlock) error {
	if err := ValidateAccountBlock(s, b); err != nil {
		return err
	}
	tx, err := s.db.NewTransaction()
	if err != nil {
		return err
	}
	defer tx.Commit()
	if err := tx.InsertAccountBlock(b); err != nil {
		return err
	}
	return tx.Commit()
}

// AddSwapBlock verifies and adds the specified swap block to this store
func (s *BlockStore) AddSwapBlock(b *tradeblocks.SwapBlock) error {
	if err := ValidateSwapBlock(s, b); err != nil {
		return err
	}
	tx, err := s.db.NewTransaction()
	if err != nil {
		return err
	}
	defer tx.Commit()
	if err := tx.InsertSwapBlock(b); err != nil {
		return err
	}
	return tx.Commit()
}

// AddOrderBlock verifies and adds the specified order block to this store
func (s *BlockStore) AddOrderBlock(b *tradeblocks.OrderBlock) error {
	if err := ValidateOrderBlock(s, b); err != nil {
		return err
	}
	tx, err := s.db.NewTransaction()
	if err != nil {
		return err
	}
	defer tx.Commit()
	if err := tx.InsertOrderBlock(b); err != nil {
		return err
	}
	return tx.Commit()
}

// AddConfirmBlock verifies and adds the specified confirm block to this store
func (s *BlockStore) AddConfirmBlock(b *tradeblocks.ConfirmBlock) error {
	// if err := ValidateConfirmBlock(s, b); err != nil {
	// 	return err
	// }
	tx, err := s.db.NewTransaction()
	if err != nil {
		return err
	}
	defer tx.Commit()
	if err := tx.InsertConfirmBlock(b); err != nil {
		return err
	}
	return tx.Commit()
}

// AccountBlocks calls the specified function with every block in this store. Return false to stop iteration.
func (s *BlockStore) AccountBlocks(f func(sequence int, b *tradeblocks.AccountBlock) bool) error {
	tx, err := s.db.NewTransaction()
	if err != nil {
		return err
	}
	defer tx.Commit()
	blocks, err := tx.GetAccountBlocks()
	if err != nil {
		return err
	}
	for i, b := range blocks {
		if !f(i, b) {
			return nil
		}
	}
	return nil
}

// SwapBlocks calls the specified function with every block in this store. Return false to stop iteration.
func (s *BlockStore) SwapBlocks(f func(sequence int, b *tradeblocks.SwapBlock) bool) error {
	tx, err := s.db.NewTransaction()
	if err != nil {
		return err
	}
	defer tx.Commit()
	blocks, err := tx.GetSwapBlocks()
	if err != nil {
		return err
	}
	for i, b := range blocks {
		if !f(i, b) {
			return nil
		}
	}
	return nil
}

// OrderBlocks calls the specified function with every block in this store. Return false to stop iteration.
func (s *BlockStore) OrderBlocks(f func(sequence int, b *tradeblocks.OrderBlock) bool) error {
	tx, err := s.db.NewTransaction()
	if err != nil {
		return err
	}
	defer tx.Commit()
	blocks, err := tx.GetOrderBlocks()
	if err != nil {
		return err
	}
	for i, b := range blocks {
		if !f(i, b) {
			return nil
		}
	}
	return nil
}

// Blocks calls the specified function with every block in this store. Return false to stop iteration.
func (s *BlockStore) Blocks(f func(sequence int, b tradeblocks.Block) bool) error {
	tx, err := s.db.NewTransaction()
	if err != nil {
		return err
	}
	defer tx.Commit()
	blocks, err := tx.GetBlocks()
	if err != nil {
		return err
	}
	for i, b := range blocks {
		if !f(i, b) {
			return nil
		}
	}
	return nil
}

// Block returns the block with the specified hash or nil if it's not found
func (s *BlockStore) Block(hash string) tradeblocks.Block {
	var tx *db.Transaction
	tx, s.err = s.db.NewTransaction()
	if s.err != nil {
		return nil
	}
	defer tx.Commit()
	var b tradeblocks.Block
	b, s.err = tx.GetBlock(hash)
	if s.err != nil {
		return nil
	}
	return b
}

// GetAccountBlock returns the account block for the specified hash or nil if it's not found
func (s *BlockStore) GetAccountBlock(hash string) *tradeblocks.AccountBlock {
	var tx *db.Transaction
	tx, s.err = s.db.NewTransaction()
	if s.err != nil {
		return nil
	}
	defer tx.Commit()
	var b *tradeblocks.AccountBlock
	b, s.err = tx.GetAccountBlock(hash)
	if s.err != nil {
		return nil
	}
	return b
}

// GetSwapBlock returns the swap block for the specified hash or nil if it's not found
func (s *BlockStore) GetSwapBlock(hash string) *tradeblocks.SwapBlock {
	var tx *db.Transaction
	tx, s.err = s.db.NewTransaction()
	if s.err != nil {
		return nil
	}
	defer tx.Commit()
	var b *tradeblocks.SwapBlock
	b, s.err = tx.GetSwapBlock(hash)
	if s.err != nil {
		return nil
	}
	return b
}

// GetOrderBlock returns the order block for the specified hash or nil if it's not found
func (s *BlockStore) GetOrderBlock(hash string) *tradeblocks.OrderBlock {
	var tx *db.Transaction
	tx, s.err = s.db.NewTransaction()
	if s.err != nil {
		return nil
	}
	defer tx.Commit()
	var b *tradeblocks.OrderBlock
	b, s.err = tx.GetOrderBlock(hash)
	if s.err != nil {
		return nil
	}
	return b
}

// GetAccountHead returns the head block for the specified account-token pair
func (s *BlockStore) GetAccountHead(account, token string) *tradeblocks.AccountBlock {
	var tx *db.Transaction
	tx, s.err = s.db.NewTransaction()
	if s.err != nil {
		return nil
	}
	defer tx.Commit()
	var b *tradeblocks.AccountBlock
	b, s.err = tx.GetAccountHead(account, token)
	if s.err != nil {
		return nil
	}
	return b
}

// GetSwapHead returns the head block for the specified account-id pair
func (s *BlockStore) GetSwapHead(account, id string) *tradeblocks.SwapBlock {
	var tx *db.Transaction
	tx, s.err = s.db.NewTransaction()
	if s.err != nil {
		return nil
	}
	defer tx.Commit()
	var b *tradeblocks.SwapBlock
	b, s.err = tx.GetSwapHead(account, id)
	if s.err != nil {
		return nil
	}
	return b
}

// GetOrderHead returns the head block for the specified account-id pair
func (s *BlockStore) GetOrderHead(account, id string) *tradeblocks.OrderBlock {
	var tx *db.Transaction
	tx, s.err = s.db.NewTransaction()
	if s.err != nil {
		return nil
	}
	defer tx.Commit()
	var b *tradeblocks.OrderBlock
	b, s.err = tx.GetOrderHead(account, id)
	if s.err != nil {
		return nil
	}
	return b
}

// GetConfirmHead returns the head block for the specified account-address pair
func (s *BlockStore) GetConfirmHead(account, address string) *tradeblocks.ConfirmBlock {
	var tx *db.Transaction
	tx, s.err = s.db.NewTransaction()
	if s.err != nil {
		return nil
	}
	defer tx.Commit()
	var b *tradeblocks.ConfirmBlock
	b, s.err = tx.GetConfirmHead(account, address)
	if s.err != nil {
		return nil
	}
	return b
}

// MatchOrdersForBuy returns orders that meet the specified criteria
func (s *BlockStore) MatchOrdersForBuy(base string, ppu float64, quote string, f func(b *tradeblocks.OrderBlock)) error {
	tx, err := s.db.NewTransaction()
	if err != nil {
		return err
	}
	defer tx.Commit()
	blocks, err := tx.GetLimitOrders(base, "<=", ppu, quote)
	if err != nil {
		return err
	}
	for _, b := range blocks {
		f(b)
	}
	return nil
}

// MatchOrdersForSell returns orders that meet the specified criteria
func (s *BlockStore) MatchOrdersForSell(base string, ppu float64, quote string, f func(b *tradeblocks.OrderBlock)) error {
	tx, err := s.db.NewTransaction()
	if err != nil {
		return err
	}
	defer tx.Commit()
	blocks, err := tx.GetLimitOrders(base, ">=", ppu, quote)
	if err != nil {
		return err
	}
	for _, b := range blocks {
		f(b)
	}
	return nil
}

// GetVariableBlock returns a block of any block type. Used currently for receive Links
// which can link to sendor commit swap
func (s *BlockStore) GetVariableBlock(hash string) tradeblocks.Block {
	return s.Block(hash)
}
