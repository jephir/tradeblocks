package app

import (
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	tb "github.com/jephir/tradeblocks"
)

// ValidateAccountBlock returns an error if validation fails for the specified account block
func ValidateAccountBlock(c AccountBlockchain, b *tb.AccountBlock) error {
	var v AccountBlockValidator
	switch b.Action {
	case "open":
		v = NewOpenValidator(c)
	case "issue":
		v = NewIssueValidator(c)
	case "send":
		v = NewSendValidator(c)
	case "receive":
		v = NewReceiveValidator(c)
	default:
		return fmt.Errorf("blockvalidator: unknown action '%s'", b.Action)
	}
	return v.ValidateAccountBlock(b)
}

// ValidateSwapBlock returns an error if validation fails for the specified swap block
func ValidateSwapBlock(c AccountBlockchain, s SwapBlockchain, b *tb.SwapBlock) error {
	v := NewSwapValidator(c, s)
	return v.ValidateSwapBlock(b)
}

// ValidateOrderBlock returns an error if validation fails for the specified Order block
func ValidateOrderBlock(c AccountBlockchain, o OrderBlockchain, b *tb.OrderBlock) error {
	v := NewOrderValidator(c, o)
	return v.ValidateOrderBlock(b)
}

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

// ValidateAccountBlock Validates that an OpenBlock is correctly formatted
func (validator OpenBlockValidator) ValidateAccountBlock(block *tb.AccountBlock) error {
	//get the chain
	accountChain := validator.accountChain
	publicKey, err := addressToRsaKey(block.Account)
	if err != nil {
		return err
	}

	if err := block.VerifyBlock(publicKey); err != nil {
		return err
	}

	// check if the previous exists, don't care if it does
	_, errPrev := validatePrevious(block, accountChain)
	if errPrev == nil {
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
		//return errors.New("balance does not match")
		return fmt.Errorf("balance expected %f; got %f", block.Balance, sendBalance)
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
func (validator IssueBlockValidator) ValidateAccountBlock(block *tb.AccountBlock) error {
	// I don't think we need to validate this after creation, this should be spawned
	// by an account creation, most fields are generated there
	// No actionable fields to check on, besides signature
	publicKey, err := addressToRsaKey(block.Account)
	if err != nil {
		return err
	}

	if err := block.VerifyBlock(publicKey); err != nil {
		return err
	}

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
func (validator SendBlockValidator) ValidateAccountBlock(block *tb.AccountBlock) error {
	//get the chain
	accountChain := validator.accountChain

	publicKey, err := addressToRsaKey(block.Account)
	if err != nil {
		return err
	}

	if err := block.VerifyBlock(publicKey); err != nil {
		return err
	}

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

// ValidateAccountBlock Validates that a ReceiveBlock is correctly formatted
func (validator ReceiveBlockValidator) ValidateAccountBlock(block *tb.AccountBlock) error {
	//get the chain
	accountChain := validator.accountChain

	publicKey, err := addressToRsaKey(block.Account)
	if err != nil {
		return err
	}

	if err := block.VerifyBlock(publicKey); err != nil {
		return err
	}

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

// NewSwapValidator returns a new validator with the given chain
func NewSwapValidator(chain AccountBlockchain, swapChain SwapBlockchain) *SwapBlockValidator {
	return &SwapBlockValidator{
		accountChain: chain,
		swapChain:    swapChain,
		orderChain:   nil,
	}
}

// ValidateSwapBlock Validates that a SwapBlocks is correctly formatted
func (validator SwapBlockValidator) ValidateSwapBlock(block *tb.SwapBlock) error {
	//get the chain
	accountChain := validator.accountChain
	swapChain := validator.swapChain
	action := block.Action

	// check the signature, non executor case
	if block.Executor != "" && (action == "commit" || action == "refund-right") {
		executorKey, err := addressToRsaKey(block.Executor)
		if err != nil {
			return err
		}

		errVerify := block.VerifyBlock(executorKey)
		if errVerify != nil {
			return errVerify
		}
	} else {
		publicKey, err := addressToRsaKey(block.Account)
		if err != nil {
			return err
		}

		if err := block.VerifyBlock(publicKey); err != nil {
			return err
		}
	}

	// check if the previous block exists
	prevBlock, errPrev := swapChain.GetSwapBlock(block.Previous)

	// originating block of swap
	if action == "offer" {
		if prevBlock != nil {
			return errors.New("prev and right must be null together")
		}

		// check if the send block referenced exists, don't get it if it does
		left, errLeft := accountChain.GetBlock(block.Left)
		if errLeft != nil || left == nil || left.Action != "send" {
			return errors.New("link field references invalid block")
		}
	} else if action == "commit" { //counterparty block
		if errPrev != nil || prevBlock == nil {
			return errors.New("previous must be not null")
		}

		// check if swaps line up
		if swapCommitAlignment(block, prevBlock) {
			return errors.New("Counterparty swap has incorrect fields: must match originating swap")
		}

		// check if the send for the original swap exists
		ogSend, errSendOriginal := accountChain.GetBlock(prevBlock.Left)
		if errSendOriginal != nil || ogSend == nil {
			return errors.New("originating send not found")
		}

		// get the send for the second swap
		sendCounter, errSendCounter := accountChain.GetBlock(block.Right)
		if errSendCounter != nil || sendCounter == nil {
			return errors.New("counter send not found")
		}

		// get the sendCounter's prev to determine quantity sent
		sendCounterPrev, errSendCounterPrev := accountChain.GetBlock(sendCounter.Previous)
		if errSendCounterPrev != nil || sendCounterPrev == nil {
			return errors.New("counter send prev not found")
		}

		// check if the tokens sent line up
		requestedQty := prevBlock.Quantity
		requestedWant := prevBlock.Want
		counterQuantity := sendCounterPrev.Balance - sendCounter.Balance
		if requestedWant != sendCounter.Token || requestedQty != counterQuantity {
			return errors.New("amount/token requested not sent")
		}
	} else if action == "refund-left" {
		if errPrev != nil || prevBlock == nil {
			return errors.New("previous must be not null")
		}

		// check if swaps line up
		if swapRefundLeftAlignment(block, prevBlock) {
			return errors.New("Counterparty swap has incorrect fields: must match originating swap")
		}

		sendBlock, errSend := accountChain.GetBlock(prevBlock.Left)
		if errSend != nil || sendBlock == nil {
			return errors.New("Originating send is invalid or not found")
		}

		// make sure RefundLeft is the initiator's account
		if block.RefundLeft != sendBlock.Account {
			return errors.New("Refund must be to initiator's account")
		}

	} else if action == "refund-right" {
		if errPrev != nil || prevBlock == nil {
			return errors.New("previous must be not null")
		}

		// make sure the previous is actually a refund-left
		if prevBlock.Action != "refund-left" {
			return errors.New("Previous must be a refund-left")
		}

		// check if swaps line up
		if swapRefundRightAlignment(block, prevBlock) {
			return errors.New("Counterparty swap has incorrect fields: must match originating swap")
		}

		// get the counterparty send
		send, err := accountChain.GetBlock(block.Right)
		if err != nil || send == nil {
			return errors.New("counterparty send not found/invalid")
		}

		// check if the refund is going to right place
		if send.Account != block.RefundRight {
			return errors.New("Account for refund must be same as original send account")
		}
	}

	return nil
}

// check if all fields beside right and previous line up
// block is the counterparty, prevBlock is the originating
func swapCommitAlignment(block *tb.SwapBlock, prevBlock *tb.SwapBlock) bool {
	return prevBlock.Account != block.Account || prevBlock.Token != block.Token ||
		prevBlock.ID != block.ID || prevBlock.Left != block.Left ||
		prevBlock.RefundLeft != block.RefundLeft || prevBlock.RefundRight != block.RefundRight ||
		prevBlock.Counterparty != block.Counterparty || prevBlock.Want != block.Want ||
		prevBlock.Quantity != block.Quantity || prevBlock.Executor != block.Executor ||
		prevBlock.Fee != block.Fee
}

// check if all fields beside right and previous line up
// block is the refund left, prevBlock is the originating
func swapRefundLeftAlignment(block *tb.SwapBlock, prevBlock *tb.SwapBlock) bool {
	return prevBlock.Account != block.Account || prevBlock.Token != block.Token ||
		prevBlock.ID != block.ID || prevBlock.Left != block.Left || prevBlock.Right != block.Right ||
		prevBlock.RefundRight != block.RefundRight ||
		prevBlock.Counterparty != block.Counterparty || prevBlock.Want != block.Want ||
		prevBlock.Quantity != block.Quantity || prevBlock.Executor != block.Executor ||
		prevBlock.Fee != block.Fee
}

// check if all fields beside right and previous line up
// block is the refund right, prevBlock is the refund left
func swapRefundRightAlignment(block *tb.SwapBlock, prevBlock *tb.SwapBlock) bool {
	return prevBlock.Account != block.Account || prevBlock.Token != block.Token ||
		prevBlock.ID != block.ID || prevBlock.Left != block.Left ||
		prevBlock.RefundLeft != block.RefundLeft ||
		prevBlock.Counterparty != block.Counterparty || prevBlock.Want != block.Want ||
		prevBlock.Quantity != block.Quantity || prevBlock.Executor != block.Executor ||
		prevBlock.Fee != block.Fee
}

// OrderBlockValidator is a validator for SwapBlocks
type OrderBlockValidator struct {
	accountChain AccountBlockchain
	swapChain    SwapBlockchain
	orderChain   OrderBlockchain
}

// NewOrderValidator returns a new validator with the given chain
func NewOrderValidator(chain AccountBlockchain, orderChain OrderBlockchain) *OrderBlockValidator {
	return &OrderBlockValidator{
		accountChain: chain,
		swapChain:    nil,
		orderChain:   orderChain,
	}
}

// ValidateOrderBlock Validates that an OrderBlocks is correctly formatted
func (validator OrderBlockValidator) ValidateOrderBlock(block *tb.OrderBlock) error {
	//get the chain
	accountChain := validator.accountChain
	orderChain := validator.orderChain

	// check the signature
	if block.Executor != "" {
		executorKey, err := addressToRsaKey(block.Executor)
		if err != nil {
			return err
		}

		errVerify := block.VerifyBlock(executorKey)
		if errVerify != nil {
			return errVerify
		}
	} else {
		publicKey, err := addressToRsaKey(block.Account)
		if err != nil {
			return err
		}

		if err := block.VerifyBlock(publicKey); err != nil {
			return err
		}
	}

	// check if the previous block exists
	addressPair := addressPrefix + block.Account + ":" + block.ID
	_, errPrev := orderChain.GetOrderBlock(addressPair)
	if errPrev != nil {
		return errPrev
	}

	// check if the originating send exists
	ogSend, errSend := accountChain.GetBlock(block.Link)
	if errSend != nil || ogSend == nil {
		return errSend
	}

	// the rest of the checks are done when an actual swap offer is started
	return nil
}

func addressToRsaKey(hash string) (*rsa.PublicKey, error) {
	publicKeyBytes, err := AddressToPublicKey(hash)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(publicKeyBytes)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.New("failed to parse DER encoded public key: " + err.Error())
	}
	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	case *dsa.PublicKey:
		return nil, errors.New("wrong public key type, have dsa, want rsa")
	case *ecdsa.PublicKey:
		return nil, errors.New("wrong public key type, have ecdsa, want rsa")
	default:
		return nil, errors.New("wrong public key type, have unknown, want rsa")
	}
}
