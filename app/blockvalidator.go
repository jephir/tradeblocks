package app

import (
	"errors"

	tb "github.com/jephir/tradeblocks"
)

// AccountBlockValidator to do server validation of each AccountBlock sent in
// see ../blockgraph.go for details on AccountBlock types
type AccountBlockValidator interface {
	// validates the given AccountBlock
	ValidateAccountBlock(block *tb.AccountBlock) error
	// constructs a validator with the given AccountBlock
	initializeValidator(chain *AccountBlockchain)
}

// OpenBlockValidator is a validator for OpenBlocks
type OpenBlockValidator struct {
	chain *AccountBlockchain
}

// ValidateAccountBlock Validates that an OoenBlock is correctly formatted
func (issue OpenBlockValidator) ValidateAccountBlock(block *tb.AccountBlock) error {
	//get the chain
	chain := *issue.chain

	// check if the send block referenced exists
	sendTx, err := chain.GetBlock(block.Link)
	if err != nil {
		return errors.New("link field references invalid block")
	}

	// check if the balances match
	if sendTx.Balance != block.Balance {
		return errors.New("balance does not match")
	}

	// check the send block references the right key pair
	if sendTx.Link != chain.GetPublicKey() {
		return errors.New("sendTx does not reference this account")
	}

	return nil
}

// IssueBlockValidator is a validator for IssueBlocks
type IssueBlockValidator struct {
	//*AccountBlockchain chain TODO
}

// ValidateAccountBlock Validates that an IssueBlock is correctly formatted
func (issue IssueBlockValidator) ValidateAccountBlock(block *tb.AccountBlock) error {
	str := block.Action

	return nil
}
