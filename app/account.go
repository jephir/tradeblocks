package app

import tb "github.com/jephir/tradeblocks"

// AccountBlockchain exists to get specific AccountBlocks
type AccountBlockchain interface {
	GetBlock(hash string) (*tb.AccountBlock, error)
}

// SwapBlockchain exists to get specific AccountBlocks
type SwapBlockchain interface {
	GetBlock(account string, id string) (*tb.SwapBlock, error)
}

// OrderBlockchain exists to get specific AccountBlocks
type OrderBlockchain interface {
	GetBlock(account string, id string) (*tb.OrderBlock, error)
}
