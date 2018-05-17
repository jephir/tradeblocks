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
}

func validatePrevious(block *tb.AccountBlock, chain AccountBlockchain) (*tb.AccountBlock, error) {
	prevBlock, err := chain.GetBlock(block.Previous)
	if err != nil || prevBlock == nil {
		return nil, errors.New("previous field was invalid")
	}
	return prevBlock, nil
}

// OpenBlockValidator is a validator for OpenBlocks
type OpenBlockValidator struct {
	accountChain AccountBlockchain
	swapChain    SwapBlockchain
	orderChain   OrderBlockchain
}

// NewOpenValidator returns a new validator with the given chain
func NewOpenValidator(chain AccountBlockchain) *OpenBlockValidator {
	return &OpenBlockValidator{
		accountChain: chain,
		swapChain:    nil,
		orderChain:   nil,
	}
}

// ValidateAccountBlock Validates that an OoenBlock is correctly formatted
func (issue OpenBlockValidator) ValidateAccountBlock(block *tb.AccountBlock) error {
	//get the chain
	accountChain := issue.accountChain

	// check if the previous exists, don't care if it does
	_, err := validatePrevious(block, accountChain)
	if err == nil {
		return errors.New("previous field was not null")
	}

	// check if the send block referenced exists
	sendBlock, err := accountChain.GetBlock(block.Link)
	if err != nil || sendBlock == nil {
		return errors.New("link field references invalid block")
	}

	// get the previous of the send to get balance
	sendBlockPrev, err := accountChain.GetBlock(sendBlock.Previous)
	if err != nil || sendBlockPrev == nil {
		return errors.New("send has no previous")
	}

	// check if the balances match
	sendBalance := sendBlockPrev.Balance - sendBlock.Balance
	if sendBalance != block.Balance {
		return errors.New("balance does not match")
	}

	// check the send block references the right key pair
	if sendBlock.Link != block.Account {
		return errors.New("send block does not reference this account")
	}

	return nil
}

// IssueBlockValidator is a validator for IssueBlocks
type IssueBlockValidator struct {
	accountChain AccountBlockchain
	swapChain    SwapBlockchain
	orderChain   OrderBlockchain
}

// NewIssueValidator returns a new validator with the given chain
func NewIssueValidator(chain AccountBlockchain) *IssueBlockValidator {
	return &IssueBlockValidator{
		accountChain: chain,
		swapChain:    nil,
		orderChain:   nil,
	}
}

// ValidateAccountBlock Validates that an IssueBlock is correctly formatted
func (issue IssueBlockValidator) ValidateAccountBlock(block *tb.AccountBlock) error {
	// I don't think we need to validate this after creation, this should be spawned
	// by an account creation, most fields are generated there
	// No actionable fields to check on

	return nil
}

// SendBlockValidator is a validator for SendBlocks
type SendBlockValidator struct {
	accountChain AccountBlockchain
	swapChain    SwapBlockchain
	orderChain   OrderBlockchain
}

// NewSendValidator returns a new validator with the given chain
func NewSendValidator(chain AccountBlockchain) *SendBlockValidator {
	return &SendBlockValidator{
		accountChain: chain,
		swapChain:    nil,
		orderChain:   nil,
	}
}

// ValidateAccountBlock Validates that an SendBlocks is correctly formatted
func (issue SendBlockValidator) ValidateAccountBlock(block *tb.AccountBlock) error {
	//get the chain
	accountChain := issue.accountChain

	// check if the previous exists, get it if it does
	prevBlock, err := validatePrevious(block, accountChain)
	if err != nil {
		return err
	}

	// check if the balances are proper
	if prevBlock.Balance-block.Balance < 0 || block.Balance > prevBlock.Balance {
		return errors.New("invalid balance amount")
	}
	return nil
}

// ReceiveBlockValidator is a validator for ReceiveBlocks
type ReceiveBlockValidator struct {
	accountChain AccountBlockchain
	swapChain    SwapBlockchain
	orderChain   OrderBlockchain
}

// NewReceiveValidator returns a new validator with the given chain
func NewReceiveValidator(chain AccountBlockchain) *ReceiveBlockValidator {
	return &ReceiveBlockValidator{
		accountChain: chain,
		swapChain:    nil,
		orderChain:   nil,
	}
}

// ValidateAccountBlock Validates that an ReceiveBlocks is correctly formatted
func (issue ReceiveBlockValidator) ValidateAccountBlock(block *tb.AccountBlock) error {
	//get the chain
	accountChain := issue.accountChain

	// check if the previous block exists, get it if it does
	prevBlock, err := validatePrevious(block, accountChain)
	if err != nil {
		return errors.New("previous field was invalid")
	}

	// check if the send block referenced exists, get it if it does
	sendBlock, err := accountChain.GetBlock(block.Link)
	if err != nil || sendBlock == nil {
		return errors.New("link field references invalid block")
	}

	// now need to get the send previous
	sendPrevBlock, err := accountChain.GetBlock(sendBlock.Previous)
	if err != nil || sendPrevBlock == nil {
		return errors.New("link field's previous references invalid block")
	}

	// check if the balances match
	balSent := sendPrevBlock.Balance - sendBlock.Balance
	balRec := block.Balance - prevBlock.Balance
	if balRec != balSent {
		return errors.New("mismatched balances")
	}

	// check if this is the intended recipient
	if sendBlock.Link != block.Account {
		return errors.New("sendBlock does not reference this account")
	}

	return nil
}

// SwapBlockValidator is a validator for SwapBlocks
type SwapBlockValidator struct {
	accountChain AccountBlockchain
	swapChain    SwapBlockchain
	orderChain   OrderBlockchain
}

/* WORK IN PROGRESS
// ValidateAccountBlock Validates that an SwapBlocks is correctly formatted
func (issue SwapBlockValidator) ValidateAccountBlock(block *tb.SwapBlock) error {
	//get the chain
	accountChain := issue.accountChain
	swapChain := issue.swapChain

	// check if the previous block doesn't exist
	prevBlock, errPrev := swapChain.GetBlock(block.Account, block.ID)

	// if right is null this is the initializing swap
	if block.Right == "" {
		if errPrev == nil {
			return errors.New("prev and right must be null together")
		}

		// check if the send block referenced exists, don't get it if it does
		_, errLeft := accountChain.GetBlock(block.Left)
		if errLeft != nil {
			return errors.New("link field references invalid block")
		}
	} else { // non null right means this is the follow up, prevBlock points to the original
		if errPrev != nil || errPrev == nil {
			return errors.New("previous must be not null")
		}

		// check if swaps line up
		if prevBlock.ID != block.ID || prevBlock.Account != block.Account ||
			prevBlock.Left != block.Left || prevBlock.Counterparty != block.Counterparty ||
			prevBlock.Want != block.Want || prevBlock.Quantity != block.Quantity {
			return errors.New("swaps do not match")
		}

		// check if the send for the original swap exists
		_, errSendOriginal := accountChain.GetBlock(prevBlock.Left)
		if errSendOriginal != nil {
			return errors.New("original send not found")
		}

		// get the send for the second swap
		sendCounter, errSendCounter := accountChain.GetBlock(block.Right)
		if errSendCounter != nil || sendCounter == nil {
			return errors.New("counter send not found")
		}

		// get the sendCounter's prev to determine quantity sent
		sendCounterPrev, errSendCounterPrev := accountChain.GetBlock(block.Right)
		if errSendCounterPrev != nil {
			return errors.New("counter send prev not found")
		}

		// check if the tokens sent line up
		requestedQty := prevBlock.Quantity
		requestedWant := prevBlock.Want
		counterQuantity := sendCounterPrev.Balance - sendCounter.Balance
		if requestedWant != sendCounter.Token || requestedQty != counterQuantity {
			return errors.New("amount requested not sent")
		}
	}

	return nil
}
*/
