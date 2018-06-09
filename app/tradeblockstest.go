package app

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/jephir/tradeblocks"
)

// BlockTestTable represents a series of block values for testing
type BlockTestTable struct {
	t             *testing.T
	AccountBlocks []*tradeblocks.AccountBlock
	SwapBlocks    []*tradeblocks.SwapBlock
	OrderBlocks   []*tradeblocks.OrderBlock
}

// NewBlockTestTable returns an initialized block test table
func NewBlockTestTable(t *testing.T) *BlockTestTable {
	return &BlockTestTable{
		t: t,
	}
}

// AddAccountBlock signs and adds an account block to the test table and returns the block
func (tt *BlockTestTable) AddAccountBlock(priv *rsa.PrivateKey, b *tradeblocks.AccountBlock) *tradeblocks.AccountBlock {
	signBlock(tt.t, priv, b)
	tt.AccountBlocks = append(tt.AccountBlocks, b)
	return b
}

// AddSwapBlock signs and adds a swap block to the test table and returns the block
func (tt *BlockTestTable) AddSwapBlock(priv *rsa.PrivateKey, b *tradeblocks.SwapBlock) *tradeblocks.SwapBlock {
	signBlock(tt.t, priv, b)
	tt.SwapBlocks = append(tt.SwapBlocks, b)
	return b
}

// AddOrderBlock signs and adds an order block to the test table and returns the block
func (tt *BlockTestTable) AddOrderBlock(priv *rsa.PrivateKey, b *tradeblocks.OrderBlock) *tradeblocks.OrderBlock {
	signBlock(tt.t, priv, b)
	tt.OrderBlocks = append(tt.OrderBlocks, b)
	return b
}

// TypedBlock represents a block with type information
type TypedBlock struct {
	*tradeblocks.AccountBlock
	*tradeblocks.SwapBlock
	*tradeblocks.OrderBlock
	T string
}

// GetAll returns all the blocks in the test table
func (tt *BlockTestTable) GetAll() []TypedBlock {
	var result []TypedBlock
	for _, b := range tt.AccountBlocks {
		result = append(result, TypedBlock{
			AccountBlock: b,
			T:     "account",
		})
	}
	for _, b := range tt.SwapBlocks {
		result = append(result, TypedBlock{
			SwapBlock: b,
			T:     "swap",
		})
	}
	for _, b := range tt.OrderBlocks {
		result = append(result, TypedBlock{
			OrderBlock: b,
			T:     "order",
		})
	}
	return result
}

func signBlock(t *testing.T, priv *rsa.PrivateKey, b tradeblocks.Block) {
	if err := b.SignBlock(priv); err != nil {
		t.Fatal(err)
	}
}

// CreateAccount returns a private key and address for a new account
func CreateAccount(t *testing.T) (priv *rsa.PrivateKey, address string) {
	var err error
	priv, err = rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	address, err = PrivateKeyToAddress(priv)
	if err != nil {
		t.Fatal(err)
	}
	return
}
