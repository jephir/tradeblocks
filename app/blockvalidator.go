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
func ValidateAccountBlock(c *BlockStore, b *tb.AccountBlock) error {
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
func ValidateSwapBlock(c *BlockStore, b *tb.SwapBlock) error {
	v := NewSwapValidator(c)
	return v.ValidateSwapBlock(b)
}

// ValidateOrderBlock returns an error if validation fails for the specified Order block
func ValidateOrderBlock(c *BlockStore, b *tb.OrderBlock) error {
	v := NewOrderValidator(c)
	return v.ValidateOrderBlock(b)
}

// AccountBlockValidator to do server validation of each AccountBlock sent in
// see ../blockgraph.go for details on AccountBlock types
type AccountBlockValidator interface {
	// validates the given AccountBlock
	ValidateAccountBlock(block *tb.AccountBlock) error
}

// OpenBlockValidator is a validator for OpenBlocks
type OpenBlockValidator struct {
	blockStore *BlockStore
}

// NewOpenValidator returns a new validator with the given chain
func NewOpenValidator(chain *BlockStore) *OpenBlockValidator {
	return &OpenBlockValidator{
		blockStore: chain,
	}
}

// ValidateAccountBlock Validates that an OpenBlock is correctly formatted
func (validator OpenBlockValidator) ValidateAccountBlock(block *tb.AccountBlock) error {
	//get the chain
	blockStore := validator.blockStore
	publicKey, err := addressToRsaKey(block.Account)
	if err != nil {
		return err
	}

	if err := block.VerifyBlock(publicKey); err != nil {
		return err
	}

	// check if the previous exists, don't care if it does
	_, errPrev := getAndVerifyAccount(block.Previous, blockStore)
	if errPrev == nil {
		return errors.New("previous field was not null")
	}

	// check if the send block referenced exists
	sendBlock, err := getAndVerifyAccount(block.Link, blockStore)
	if err != nil || sendBlock == nil {
		return errors.New("link field references invalid block")
	}

	// get the previous of the send to get balance
	sendBlockPrev, err := getAndVerifyAccount(sendBlock.Previous, blockStore)
	if err != nil || sendBlockPrev == nil {
		return errors.New("send has no previous")
	}

	// check if the balances match
	if block.Balance < 0 {
		return errors.New("Balance must be positive")
	}
	sendBalance := sendBlockPrev.Balance - sendBlock.Balance
	if sendBalance != block.Balance {
		//return errors.New("balance does not match")
		return fmt.Errorf("balance expected %f; got %f", block.Balance, sendBalance)
	}

	// check the send block references the right key pair
	if sendBlock.Link != block.Account {
		return fmt.Errorf("send link '%s' does not reference account '%s'", sendBlock.Link, block.Account)
	}

	return nil
}

// IssueBlockValidator is a validator for IssueBlocks
type IssueBlockValidator struct {
	blockStore *BlockStore
}

// NewIssueValidator returns a new validator with the given chain
func NewIssueValidator(blockStore *BlockStore) *IssueBlockValidator {
	return &IssueBlockValidator{
		blockStore: blockStore,
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
	blockStore *BlockStore
}

// NewSendValidator returns a new validator with the given chain
func NewSendValidator(blockStore *BlockStore) *SendBlockValidator {
	return &SendBlockValidator{
		blockStore: blockStore,
	}
}

// ValidateAccountBlock Validates that an SendBlocks is correctly formatted
func (validator SendBlockValidator) ValidateAccountBlock(block *tb.AccountBlock) error {
	//get the chain
	blockStore := validator.blockStore

	publicKey, err := addressToRsaKey(block.Account)
	if err != nil {
		return err
	}

	if err := block.VerifyBlock(publicKey); err != nil {
		return err
	}

	// check if the previous exists, get it if it does
	prevBlock, err := getAndVerifyAccount(block.Previous, blockStore)
	if err != nil {
		return errors.New("Previous block invalid")
	}

	// check if the balances are proper
	if block.Balance < 0 {
		return errors.New("Balance must be positive")
	}
	if block.Balance > prevBlock.Balance {
		return errors.New("invalid balance amount")
	}
	return nil
}

// ReceiveBlockValidator is a validator for ReceiveBlocks
type ReceiveBlockValidator struct {
	blockStore *BlockStore
}

// NewReceiveValidator returns a new validator with the given chain
func NewReceiveValidator(blockStore *BlockStore) *ReceiveBlockValidator {
	return &ReceiveBlockValidator{
		blockStore: blockStore,
	}
}

// ValidateAccountBlock Validates that a ReceiveBlock is correctly formatted
func (validator ReceiveBlockValidator) ValidateAccountBlock(block *tb.AccountBlock) error {
	//get the chain
	blockStore := validator.blockStore

	publicKey, err := addressToRsaKey(block.Account)
	if err != nil {
		return err
	}

	if err := block.VerifyBlock(publicKey); err != nil {
		return err
	}

	// check if the previous block exists, get it if it does
	prevBlock, err := getAndVerifyAccount(block.Previous, blockStore)
	if err != nil {
		return errors.New("previous field was invalid")
	}

	// check if the send block referenced exists, get it if it does
	sendBlock, err := getAndVerifyAccount(block.Link, blockStore)
	if err != nil || sendBlock == nil {
		return errors.New("link field references invalid block")
	}

	// now need to get the send previous
	sendPrevBlock, err := getAndVerifyAccount(sendBlock.Previous, blockStore)
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
	blockStore *BlockStore
}

// NewSwapValidator returns a new validator with the given chain
func NewSwapValidator(blockStore *BlockStore) *SwapBlockValidator {
	return &SwapBlockValidator{
		blockStore: blockStore,
	}
}

// ValidateSwapBlock Validates that a SwapBlocks is correctly formatted
func (validator SwapBlockValidator) ValidateSwapBlock(block *tb.SwapBlock) error {
	//get the chain
	blockStore := validator.blockStore
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
	prevBlock, errPrev := getAndVerifySwap(block.Previous, blockStore)

	// originating block of swap
	if action == "offer" {
		if prevBlock != nil {
			return errors.New("prev and right must be null together")
		}

		// check if the send block referenced exists, don't get it if it does
		left, errLeft := getAndVerifyAccount(block.Left, blockStore)
		if errLeft != nil || left == nil || left.Action != "send" {
			return errors.New("link field references invalid block")
		}

		// check to see if the send (left) is pointed at this block
		if left.Link != block.Account+":swap:"+block.ID {
			return errors.New("Linked left block does not send to this swap")
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
		ogSend, errSendOriginal := getAndVerifyAccount(prevBlock.Left, blockStore)
		if errSendOriginal != nil || ogSend == nil {
			return errors.New("originating send not found")
		}

		// get the send (right) for the second swap
		sendCounter, errSendCounter := getAndVerifyAccount(block.Right, blockStore)
		if errSendCounter != nil || sendCounter == nil {
			return errors.New("counter send not found")
		}

		// check to see if the send (left) is pointed at this block
		if sendCounter.Link != block.Account+":swap:"+block.ID {
			return errors.New("Linked right block does not send to this swap")
		}

		// get the sendCounter's prev to determine quantity sent
		sendCounterPrev, errSendCounterPrev := getAndVerifyAccount(sendCounter.Previous, blockStore)
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

		sendBlock, errSend := getAndVerifyAccount(prevBlock.Left, blockStore)
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
			return errors.New("Refund Right must match Refund Left fields")
		}

		// get the counterparty send
		send, err := getAndVerifyAccount(block.Right, blockStore)
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
	blockStore *BlockStore
}

// NewOrderValidator returns a new validator with the given chain
func NewOrderValidator(blockStore *BlockStore) *OrderBlockValidator {
	return &OrderBlockValidator{
		blockStore: blockStore,
	}
}

// ValidateOrderBlock Validates that an OrderBlocks is correctly formatted
func (validator OrderBlockValidator) ValidateOrderBlock(block *tb.OrderBlock) error {
	//get the chain
	blockStore := validator.blockStore
	action := block.Action

	switch action {
	case "create-order":
		// check the signature
		publicKey, err := addressToRsaKey(block.Account)
		if err != nil {
			return err
		}

		if err := block.VerifyBlock(publicKey); err != nil {
			return err
		}

		// previous should be null
		if block.Previous != "" {
			return errors.New("Previous should be null")
		}

		// get the originating send
		ogSend, err := getAndVerifyAccount(block.Link, blockStore)
		if err != nil || ogSend == nil {
			return errors.New("Order linked send not found")
		}

		// check to see if the send is pointed at this order
		if ogSend.Link != block.Account+":order:"+block.ID {
			return errors.New("Linked send block does not send to this order")
		}

		// get the previous of the send
		ogPrevSend, err := getAndVerifyAccount(ogSend.Previous, blockStore)
		if err != nil || ogPrevSend == nil {
			return errors.New("Linked send block does not have a valid previous")
		}

		// check if the balances line up
		balanceSent := ogPrevSend.Balance - ogSend.Balance
		if balanceSent != block.Balance {
			return errors.New("Balance sent and Balance created do not match up")
		}

	case "accept-order":
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
		prevBlock, err := getAndVerifyOrder(block.Previous, blockStore)
		if err != nil || prevBlock == nil {
			return errors.New("Previous block invalid")
		}

		// check if fields beside balance, link, and previous line up
		if orderAcceptAlignment(block, prevBlock) {
			return errors.New("Fields did not line up with order creation")
		}

		// get the linked swap
		swapBlock, err := getAndVerifySwap(block.Link, blockStore)
		if err != nil || swapBlock == nil {
			return errors.New("Linked swap not found")
		}

		// get the linked swap's send
		swapSendBlock, err := getAndVerifyAccount(swapBlock.Left, blockStore)
		if err != nil || swapSendBlock == nil {
			return errors.New("Linked swap not found")
		}

		// get the linked swap's send previous
		swapSendPrevBlock, err := getAndVerifyAccount(swapSendBlock.Previous, blockStore)
		if err != nil || swapSendPrevBlock == nil {
			return errors.New("Linked swap not found")
		}

		// check the swap's counterparty is the order's account
		if swapBlock.Counterparty != block.Account {
			return errors.New("The swap must have counterparty point to this order")
		}

		// check the ID is the same for swap and order
		if swapBlock.ID != block.ID {
			return errors.New("The swap must have same ID as the order")
		}

		// check if the token type lines up
		if swapBlock.Want != block.Token || swapBlock.Token != block.Quote {
			return errors.New("Swap and Order token mismatch")
		}

		// Balances check
		swapWant := swapBlock.Quantity
		orderBalance := prevBlock.Balance - block.Balance
		orderPrice := block.Price
		swapSendQuantity := swapSendPrevBlock.Balance - swapSendBlock.Balance
		// valid block balance
		if block.Balance < 0 {
			return errors.New("Invalid block balance, must be greater than zero")
		}
		// check if allowed to not fill the whole order
		if !block.Partial {
			if block.Balance != 0 {
				return errors.New("Balance must be paid in full for blocks with Partial = false")
			}
		}
		// check to see if order gets what it wants
		if orderPrice*swapSendQuantity != orderBalance {
			return errors.New("Balance sent to order is invalid")
		}

		// check if swap gets what it wants
		if orderBalance != swapWant {
			return errors.New("Balance sent to swap is invalid")
		}

	case "refund-order":
		publicKey, err := addressToRsaKey(block.Account)
		if err != nil {
			return err
		}

		if err := block.VerifyBlock(publicKey); err != nil {
			return err
		}

		// check if the previous block exists
		prevBlock, errPrev := getAndVerifyOrder(block.Previous, blockStore)
		if errPrev != nil || prevBlock == nil {
			return errors.New("Previous block invalid")
		}

		// check if fields beside link and previous line up
		if orderRefundAlignment(block, prevBlock) {
			return errors.New("Fields did not line up with head order block")
		}

		// make sure the link is to the originating send account
		if block.Account != block.Link {
			return errors.New("Must refund to the original sender")
		}

	default:
		return errors.New("undefined action type: " + action)
	}

	return nil
}

// check if all fields beside link, balance, and previous
// block is the accept-order, prevBlock is the create-order
func orderAcceptAlignment(block *tb.OrderBlock, prevBlock *tb.OrderBlock) bool {
	return block.Account != prevBlock.Account || block.Token != prevBlock.Token ||
		block.ID != prevBlock.ID || block.Quote != prevBlock.Quote || block.Price != prevBlock.Price ||
		block.Partial != prevBlock.Partial || block.Executor != prevBlock.Executor || block.Fee != prevBlock.Fee
}

// check if all fields beside link and previous
// block is the refund-order, prevBlock is the current head orderblock.
func orderRefundAlignment(block *tb.OrderBlock, prevBlock *tb.OrderBlock) bool {
	return block.Account != prevBlock.Account || block.Token != prevBlock.Token || block.Balance != prevBlock.Balance ||
		block.ID != prevBlock.ID || block.Quote != prevBlock.Quote || block.Price != prevBlock.Price ||
		block.Partial != prevBlock.Partial || block.Executor != prevBlock.Executor || block.Fee != prevBlock.Fee
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

func getAndVerifyAccount(hash string, chain *BlockStore) (*tb.AccountBlock, error) {
	// check if the previous block exists
	block := chain.GetAccountBlock(hash)
	if block == nil {
		return nil, errors.New("Getting block failed for hash: " + hash)
	}

	publicKey, err := addressToRsaKey(block.Account)
	if err != nil {
		return nil, err
	}

	err = block.VerifyBlock(publicKey)
	if err != nil {
		return nil, errors.New("Verification of block failed")
	}

	return block, nil
}

func getAndVerifySwap(hash string, chain *BlockStore) (*tb.SwapBlock, error) {
	// check if the previous block exists
	block := chain.GetSwapBlock(hash)
	if block == nil {
		return nil, errors.New("Getting block failed for hash: " + hash)
	}

	address := block.Account
	if (block.Action == "commit" || block.Action == "refund-right") && block.Executor != "" {
		address = block.Executor
	}

	publicKey, err := addressToRsaKey(address)
	if err != nil {
		return nil, err
	}

	err = block.VerifyBlock(publicKey)
	if err != nil {
		return nil, errors.New("Verification of block failed")
	}

	return block, nil
}

func getAndVerifyOrder(hash string, chain *BlockStore) (*tb.OrderBlock, error) {
	// check if the previous block exists
	block := chain.GetOrderBlock(hash)
	if block == nil {
		return nil, errors.New("Getting block failed for hash: " + hash)
	}

	address := block.Account
	if block.Action == "accept-order" && block.Executor != "" {
		address = block.Executor
	}

	publicKey, err := addressToRsaKey(address)
	if err != nil {
		return nil, err
	}

	err = block.VerifyBlock(publicKey)
	if err != nil {
		return nil, errors.New("Verification of block failed")
	}

	return block, nil
}
