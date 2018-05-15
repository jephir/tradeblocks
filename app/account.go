package app

import tb "github.com/jephir/tradeblocks"

// AccountBlockchain exists to get specific AccountBlocks
type AccountBlockchain interface {
	GetBlock(hash string) (*tb.AccountBlock, error)
	GetPublicKey() string
}
