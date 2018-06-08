package app

import (
	"crypto/rsa"
	"strings"
	"testing"

	"github.com/jephir/tradeblocks"
)

// something to use for address that won't break the decrypter
const badAddress = "Ci0tLS0tQkVHSU4gUlNBIFBVQkxJQyBLRVktLS0tLQpNSUdmTUEwR0NTcUdTSWIzRFFFQkFRVUFBNEdOQURDQmlRS0JnUURWbGVZUStNT0doSFZ2a216Q2tKcmpJNUNMCjROTUh3TlJsN1NSbkVsRkkyK25XallNRXdTT2xwNXBUY0hCempSaEpPeDFTYkx0aUtSS0ZnMVE5d1Vldk5lV1MKUE1qQjFsK0xXbVVUUnFOVGNBUFFjMFZkZXVtanFzMVArZUhFUmZrOU13cU5zclB5dHZHd3ZOUUowNVBrZ0xTawpYdTU4a3I1aVh4TUFCSXVrYlFJREFRQUIKLS0tLS1FTkQgUlNBIFBVQkxJQyBLRVktLS0tLQo"

func openSetup(key *rsa.PrivateKey, address string) (*tradeblocks.AccountBlock, *tradeblocks.AccountBlock, AccountBlockValidator, error) {
	s := NewBlockStore()
	i := tradeblocks.NewIssueBlock(address, 100.0)
	send := tradeblocks.NewSendBlock(i, address, 100.0)

	open := tradeblocks.NewOpenBlockFromSend(address, send, 100.0)

	validator := NewOpenValidator(s)

	errSignIssue := i.SignBlock(key)
	if errSignIssue != nil {
		return nil, nil, nil, errSignIssue
	}

	errSignSend := send.SignBlock(key)
	if errSignSend != nil {
		return nil, nil, nil, errSignSend
	}

	errSignOpen := open.SignBlock(key)
	if errSignOpen != nil {
		return nil, nil, nil, errSignOpen
	}

	if _, err := s.AddBlock(i); err != nil {
		return nil, nil, nil, err
	}

	if _, err := s.AddBlock(send); err != nil {
		return nil, nil, nil, err
	}
	return open, send, validator, nil
}

func TestOpenBlockValidator(t *testing.T) {
	key, address := CreateAccount(t)
	// test for success
	open, _, validator, errSetup := openSetup(key, address)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	errValidate := validator.ValidateAccountBlock(open)
	if errValidate != nil {
		t.Fatal(errValidate)
	}

	// test for previous not null
	open, _, validator, errPrevNull := openSetup(key, address)
	if errPrevNull != nil {
		t.Fatal(errPrevNull)
	}
	open.Previous = open.Link

	errSignOpen := open.SignBlock(key)
	if errSignOpen != nil {
		t.Fatal(errSignOpen)
	}

	err := validator.ValidateAccountBlock(open)
	expectedError := "previous field was not null"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test for send not existing
	open, _, validator, errSendNull := openSetup(key, address)
	if errSendNull != nil {
		t.Fatal(errSendNull)
	}
	open.Link = ""
	errSignOpen = open.SignBlock(key)
	if errSignOpen != nil {
		t.Fatal(errSignOpen)
	}

	err = validator.ValidateAccountBlock(open)
	expectedError = "link field references invalid block"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test for send prev not existing
	open, send, validator, errSendPrevNull := openSetup(key, address)
	if errSendPrevNull != nil {
		t.Fatal(errSendPrevNull)
	}
	send.Previous = ""

	err = validator.ValidateAccountBlock(open)
	expectedError = "send has no previous"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test for send not having necessary balance
	open, send, validator, errSendBalance := openSetup(key, address)
	if errSendBalance != nil {
		t.Fatal(errSendBalance)
	}
	send.Balance = 51

	err = validator.ValidateAccountBlock(open)
	expectedError = "balance expected 100.000000; got 49.000000"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test for send not referencing this open
	open, send, validator, errSendNoLink := openSetup(key, address)
	if errSendNoLink != nil {
		t.Fatal(errSendNoLink)
	}
	send.Link = "WRONG_ACCOUNT"

	err = validator.ValidateAccountBlock(open)
	expectedError = "send link 'WRONG_ACCOUNT' does not reference account"
	if err == nil || !strings.Contains(err.Error(), expectedError) {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}
}

func issueSetup(key *rsa.PrivateKey, address string) (*tradeblocks.AccountBlock, AccountBlockValidator, error) {
	issue := tradeblocks.NewIssueBlock(address, 100)
	s := NewBlockStore()

	validator := NewIssueValidator(s)

	errSignIssue := issue.SignBlock(key)
	if errSignIssue != nil {
		return nil, nil, errSignIssue
	}

	if _, err := s.AddBlock(issue); err != nil {
		return nil, nil, err
	}

	return issue, validator, nil
}

func TestIssueBlockValidator(t *testing.T) {
	key, address := CreateAccount(t)
	// test for success
	issue, validator, errSetup := issueSetup(key, address)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	errValidate := validator.ValidateAccountBlock(issue)
	if errValidate != nil {
		t.Fatal(errValidate)
	}
}

func sendSetup(key *rsa.PrivateKey, address string) (*tradeblocks.AccountBlock, AccountBlockValidator, error) {
	s := NewBlockStore()
	issue := tradeblocks.NewIssueBlock(address, 100.0)

	send := tradeblocks.NewSendBlock(issue, address, 100.0)

	validator := NewSendValidator(s)

	errSignSend := send.SignBlock(key)
	if errSignSend != nil {
		return nil, nil, errSignSend
	}

	errSignIssue := issue.SignBlock(key)
	if errSignIssue != nil {
		return nil, nil, errSignIssue
	}

	if _, err := s.AddBlock(issue); err != nil {
		return nil, nil, err
	}

	if _, err := s.AddBlock(send); err != nil {
		return nil, nil, err
	}

	return send, validator, nil
}

func TestSendBlockValidator(t *testing.T) {
	key, address := CreateAccount(t)
	// test for success
	send, validator, errSetup := sendSetup(key, address)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	errValidate := validator.ValidateAccountBlock(send)
	if errValidate != nil {
		t.Fatal(errValidate)
	}

	// test for previous not null
	send, validator, errPrevNull := sendSetup(key, address)
	if errPrevNull != nil {
		t.Fatal(errPrevNull)
	}
	send.Previous = ""
	errSignSend := send.SignBlock(key)
	if errSignSend != nil {
		t.Fatal(errSignSend)
	}

	err := validator.ValidateAccountBlock(send)
	expectedError := "previous field was invalid"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}
}

func receiveSetup(key *rsa.PrivateKey, address string, t *testing.T) (*tradeblocks.AccountBlock, *tradeblocks.AccountBlock, AccountBlockValidator, error) {
	key2, address2 := CreateAccount(t)
	s := NewBlockStore()

	i := tradeblocks.NewIssueBlock(address2, 100.0)

	send := tradeblocks.NewSendBlock(i, address, 50.0)

	i2 := tradeblocks.NewIssueBlock(address, 100.0)

	receive := tradeblocks.NewReceiveBlockFromSend(i2, send, 50)

	validator := NewReceiveValidator(s)

	errSignIssue := i.SignBlock(key2)
	if errSignIssue != nil {
		t.Fatal(errSignIssue)
	}

	errSignIssue = i2.SignBlock(key)
	if errSignIssue != nil {
		t.Fatal(errSignIssue)
	}

	errSignSend := send.SignBlock(key2)
	if errSignSend != nil {
		t.Fatal(errSignSend)
	}

	errSignReceive := receive.SignBlock(key)
	if errSignReceive != nil {
		t.Fatal(errSignReceive)
	}

	if _, err := s.AddBlock(i); err != nil {
		t.Fatal(err)
	}

	if _, err := s.AddBlock(send); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddBlock(i2); err != nil {
		t.Fatal(err)
	}

	return receive, send, validator, nil
}

func TestReceiveBlockValidator(t *testing.T) {
	key, address := CreateAccount(t)

	// test for success
	receive, _, validator, errSetup := receiveSetup(key, address, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	errValidate := validator.ValidateAccountBlock(receive)
	if errValidate != nil {
		t.Fatal(errValidate)
	}

	// test for previous not null
	receive, _, validator, errPrevNull := receiveSetup(key, address, t)
	if errPrevNull != nil {
		t.Fatal(errPrevNull)
	}

	receive.Previous = ""
	errSignReceive := receive.SignBlock(key)
	if errSignReceive != nil {
		t.Fatal(errSignReceive)
	}

	err := validator.ValidateAccountBlock(receive)
	expectedError := "previous field was invalid"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test for link invalid
	receive, _, validator, errLink := receiveSetup(key, address, t)
	if errLink != nil {
		t.Fatal(errLink)
	}
	receive.Link = ""
	errSignReceive = receive.SignBlock(key)
	if errSignReceive != nil {
		t.Fatal(errSignReceive)
	}

	err = validator.ValidateAccountBlock(receive)
	expectedError = "link field references invalid block"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test for link previous invalid
	receive, send, validator, errLinkPrev := receiveSetup(key, address, t)
	if errLinkPrev != nil {
		t.Fatal(errLinkPrev)
	}
	send.Previous = ""
	errSignReceive = receive.SignBlock(key)
	if errSignReceive != nil {
		t.Fatal(errSignReceive)
	}

	err = validator.ValidateAccountBlock(receive)
	expectedError = "link field's previous references invalid block"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test for send linked to this
	receive, send, validator, errSendLink := receiveSetup(key, address, t)
	if errSendLink != nil {
		t.Fatal(errSendLink)
	}
	send.Previous = ""
	errSignSend := send.SignBlock(key)
	if errSignSend != nil {
		t.Fatal(errSignSend)
	}

	err = validator.ValidateAccountBlock(receive)
	expectedError = "link field's previous references invalid block"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}
}

func TestUnkownAction(t *testing.T) {
	err := ValidateAccountBlock(nil, &tradeblocks.AccountBlock{})
	if err.Error() != "blockvalidator: unknown action ''" {
		t.Fatalf("expected an error but got %s", err.Error())
	}
}

func swapOfferSetup(keys []*rsa.PrivateKey, address []string, t *testing.T) (*tradeblocks.SwapBlock, *tradeblocks.SwapBlock, *tradeblocks.AccountBlock, *SwapBlockValidator, error) {
	accountStore := NewBlockStore()
	swapStore := NewSwapBlockStore()

	i := tradeblocks.NewIssueBlock(address[0], 100.0)
	send := tradeblocks.NewSendBlock(i, address[2], 50.0)

	i2 := tradeblocks.NewIssueBlock(address[1], 50.0)
	send2 := tradeblocks.NewSendBlock(i2, address[2], 10.0)

	swap := tradeblocks.NewOfferBlock(address[2], send, "test-ID", "counterparty", address[1], 10.0, "", 0.0)
	swap2 := tradeblocks.NewCommitBlock(swap, send2)

	err := i.SignBlock(keys[0])
	if err != nil {
		t.Fatal(err)
	}

	err = i2.SignBlock(keys[1])
	if err != nil {
		t.Fatal(err)
	}

	err = send.SignBlock(keys[0])
	if err != nil {
		t.Fatal(err)
	}

	err = send2.SignBlock(keys[1])
	if err != nil {
		t.Fatal(err)
	}

	err = swap.SignBlock(keys[2])
	if err != nil {
		t.Fatal(err)
	}

	err = swap2.SignBlock(keys[2])
	if err != nil {
		t.Fatal(err)
	}

	_, err = accountStore.AddBlock(i)
	if err != nil {
		t.Fatal(err)
	}
	_, err = accountStore.AddBlock(send)
	if err != nil {
		t.Fatal(err)
	}
	_, err = accountStore.AddBlock(i2)
	if err != nil {
		t.Fatal(err)
	}

	_, err = accountStore.AddBlock(send2)
	if err != nil {
		t.Fatal(err)
	}

	_, err = swapStore.AddBlock(swap, accountStore)
	if err != nil {
		t.Fatal(err)
	}

	_, err = swapStore.AddBlock(swap2, accountStore)
	if err != nil {
		t.Fatal(err)
	}

	validator := NewSwapValidator(accountStore, swapStore)

	return swap, swap2, send, validator, nil
}

// only for offer, each action has different test
func TestSwapOfferValidation(t *testing.T) {
	key, address := CreateAccount(t)
	key2, address2 := CreateAccount(t)
	key3, address3 := CreateAccount(t)

	keyList := []*rsa.PrivateKey{key, key2, key3}
	addressList := []string{address, address2, address3}

	swap, _, _, validator, errSetup := swapOfferSetup(keyList, addressList, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	// base success
	errVerify := validator.ValidateSwapBlock(swap)
	if errVerify != nil {
		t.Fatal(errVerify)
	}

	// signature verifying tests
	// Note: can't do executors on offers
	swap, _, _, validator, errSetup = swapOfferSetup(keyList, addressList, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	// random signature
	swap.Signature = "garbage"
	errVerify = validator.ValidateSwapBlock(swap)
	expectedError := "illegal base64 data at input byte 4"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// different key signature
	swap, _, _, validator, errSetup = swapOfferSetup(keyList, addressList, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	errSign := swap.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}
	errVerify = validator.ValidateSwapBlock(swap)
	expectedError = "crypto/rsa: verification error"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// is there a previous? Hint: there shouldn't be
	swap, swap2, _, validator, err := swapOfferSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	swap.Previous = swap2.Hash()
	errSign = swap.SignBlock(key3)
	if errSign != nil {
		t.Fatal(errSign)
	}
	errVerify = validator.ValidateSwapBlock(swap)
	expectedError = "prev and right must be null together"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// make sure there's an account block from the left field
	swap, _, _, validator, errSetup = swapOfferSetup(keyList, addressList, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	// no block
	swap.Left = "not a block"
	errSign = swap.SignBlock(key3)
	if errSign != nil {
		t.Fatal(errSign)
	}
	errVerify = validator.ValidateSwapBlock(swap)
	expectedError = "link field references invalid block"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	//not a send
	tempI := tradeblocks.NewIssueBlock("address", 100.0)
	swap.Left = tempI.Hash()
	errSign = swap.SignBlock(key3)
	if errSign != nil {
		t.Fatal(errSign)
	}
	errVerify = validator.ValidateSwapBlock(swap)
	expectedError = "link field references invalid block"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}
}

// only for commit, each action has different test
func TestSwapCommitValidation(t *testing.T) {
	key, address := CreateAccount(t)
	key2, address2 := CreateAccount(t)
	key3, address3 := CreateAccount(t)

	keyList := []*rsa.PrivateKey{key, key2, key3}
	addressList := []string{address, address2, address3}

	_, swap2, _, validator, errSetup := swapOfferSetup(keyList, addressList, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	// base success
	errVerify := validator.ValidateSwapBlock(swap2)
	if errVerify != nil {
		t.Fatal(errVerify)
	}

	// signature verifying tests
	// Note: can't do executors on offers
	_, swap2, send, validator, errSetup := swapOfferSetup(keyList, addressList, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	// random signature
	swap2.Signature = "garbage"
	errVerify = validator.ValidateSwapBlock(swap2)
	expectedError := "illegal base64 data at input byte 4"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	swap2.Previous = send.Hash()
	errSign := swap2.SignBlock(key3)
	if errSign != nil {
		t.Fatal(errSign)
	}
	errVerify = validator.ValidateSwapBlock(swap2)
	expectedError = "previous must be not null"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// bad account decoding
	swap2.Account = "help"
	errSign = swap2.SignBlock(key3)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(swap2)
	expectedError = "failed to parse PEM block containing the public key"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// bad executor
	swap2.Executor = "help"
	errSign = swap2.SignBlock(key3)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(swap2)
	expectedError = "failed to parse PEM block containing the public key"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// swaps not same fields
	_, swap2, _, validator, errSetup = swapOfferSetup(keyList, addressList, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}
	// long list of fields that could trigger this:
	// Account, Token, ID, Left, RefundLeft, RefundRight,
	// Counterparty, Want, Quantity, Executor, Fee
	// Note on account: used as address to get publicKey, changing will cause
	// a pem parsing error
	swap2.ID = "help"
	errSign = swap2.SignBlock(key3)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(swap2)
	expectedError = "Counterparty swap has incorrect fields: must match originating swap"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// original send doesn't exist (linked to bad swap)
	swap, swap2, _, validator, errSetup := swapOfferSetup(keyList, addressList, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	swap.Left = badAddress
	swap2.Left = badAddress
	errSign = swap2.SignBlock(key3)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(swap2)
	expectedError = "originating send not found"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// counter send doesn't exist
	swap, swap2, _, validator, errSetup = swapOfferSetup(keyList, addressList, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	swap2.Right = badAddress
	errSign = swap2.SignBlock(key3)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(swap2)
	expectedError = "counter send not found"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// counter send doesn't exist
	swap, swap2, _, validator, errSetup = swapOfferSetup(keyList, addressList, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	send2, err := validator.accountChain.GetBlock(swap2.Right)
	if err != nil || send2 == nil {
		t.Fatal(err)
	}
	send2.Previous = badAddress
	errSign = send2.SignBlock(key3)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(swap2)
	expectedError = "counter send prev not found"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// counter send doesn't exist
	swap, swap2, _, validator, errSetup = swapOfferSetup(keyList, addressList, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	send2, err = validator.accountChain.GetBlock(swap2.Right)
	if err != nil || send2 == nil {
		t.Fatal(err)
	}
	send2.Balance = 0
	errSign = send2.SignBlock(key3)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(swap2)
	expectedError = "amount/token requested not sent"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}
}

func swapRefundLeftSetup(key []*rsa.PrivateKey, address []string, t *testing.T) (*tradeblocks.SwapBlock, *tradeblocks.SwapBlock, *tradeblocks.AccountBlock, *SwapBlockValidator, error) {
	accountStore := NewBlockStore()
	swapStore := NewSwapBlockStore()

	i := tradeblocks.NewIssueBlock(address[0], 100.0)
	send := tradeblocks.NewSendBlock(i, address[2], 50.0)

	i2 := tradeblocks.NewIssueBlock(address[1], 50.0)
	swap := tradeblocks.NewOfferBlock(address[2], send, "test-ID", "counterparty", address[1], 10.0, "", 0.0)

	refundLeft := tradeblocks.NewRefundLeftBlock(swap, address[0])

	errSignIssue := i.SignBlock(key[0])
	if errSignIssue != nil {
		t.Fatal(errSignIssue)
	}

	errSignIssue = i2.SignBlock(key[1])
	if errSignIssue != nil {
		t.Fatal(errSignIssue)
	}

	errSignSend := send.SignBlock(key[0])
	if errSignSend != nil {
		t.Fatal(errSignSend)
	}

	errSignSwap := swap.SignBlock(key[2])
	if errSignSwap != nil {
		t.Fatal(errSignSwap)
	}

	errSignRefund := refundLeft.SignBlock(key[2])
	if errSignRefund != nil {
		t.Fatal(errSignRefund)
	}

	accountStore.AddBlock(i)
	accountStore.AddBlock(send)
	accountStore.AddBlock(i2)

	_, err := swapStore.AddBlock(swap, accountStore)
	if err != nil {
		t.Fatal(err)
	}
	_, err = swapStore.AddBlock(refundLeft, accountStore)
	if err != nil {
		t.Fatal(err)
	}

	validator := NewSwapValidator(accountStore, swapStore)

	return swap, refundLeft, send, validator, nil
}

// only for refund-left, each action has different test
func TestSwapRefundLeftValidation(t *testing.T) {
	key, address := CreateAccount(t)
	key2, address2 := CreateAccount(t)
	key3, address3 := CreateAccount(t)

	keyList := []*rsa.PrivateKey{key, key2, key3}
	addressList := []string{address, address2, address3}

	_, refundLeft, _, validator, errSetup := swapRefundLeftSetup(keyList, addressList, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	// base success
	errVerify := validator.ValidateSwapBlock(refundLeft)
	if errVerify != nil {
		t.Fatal(errVerify)
	}

	// signature verifying tests
	// Note: can't do executors on offers
	_, refundLeft, send, validator, errSetup := swapRefundLeftSetup(keyList, addressList, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	// random signature
	refundLeft.Signature = "garbage"
	errVerify = validator.ValidateSwapBlock(refundLeft)
	expectedError := "illegal base64 data at input byte 4"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// invalid previous
	refundLeft.Previous = send.Hash()
	errSign := refundLeft.SignBlock(key3)
	if errSign != nil {
		t.Fatal(errSign)
	}
	errVerify = validator.ValidateSwapBlock(refundLeft)
	expectedError = "previous must be not null"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// long list of fields that could trigger this:
	// Account, Token, ID, Left, Right, RefundRight,
	// Counterparty, Want, Quantity, Executor, Fee
	// Note on account: used as address to get publicKey, changing will cause
	// a pem parsing error
	swap, refundLeft, send, validator, errSetup := swapRefundLeftSetup(keyList, addressList, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	refundLeft.Fee = 1.0
	errSign = refundLeft.SignBlock(key3)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(refundLeft)
	expectedError = "Counterparty swap has incorrect fields: must match originating swap"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// originating send invalid
	swap, refundLeft, send, validator, errSetup = swapRefundLeftSetup(keyList, addressList, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	swap.Left = badAddress
	refundLeft.Left = badAddress
	errSign = refundLeft.SignBlock(key3)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(refundLeft)
	expectedError = "Originating send is invalid or not found"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// refundleft is not the initiators account
	swap, refundLeft, send, validator, errSetup = swapRefundLeftSetup(keyList, addressList, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	refundLeft.RefundLeft = badAddress
	errSign = refundLeft.SignBlock(key3)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(refundLeft)
	expectedError = "Refund must be to initiator's account"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}
}

func swapRefundRightSetup(key []*rsa.PrivateKey, address []string, t *testing.T) (*tradeblocks.SwapBlock, *tradeblocks.SwapBlock, *tradeblocks.SwapBlock, *tradeblocks.AccountBlock, *SwapBlockValidator, error) {
	accountStore := NewBlockStore()
	swapStore := NewSwapBlockStore()

	i := tradeblocks.NewIssueBlock(address[0], 100.0)
	send := tradeblocks.NewSendBlock(i, address[0], 50.0)
	i2 := tradeblocks.NewIssueBlock(address[1], 50)
	send2 := tradeblocks.NewSendBlock(i2, address[2], 10.0)
	swap := tradeblocks.NewOfferBlock(address[2], send, "test-ID", "counterparty", address[1], 10.0, "", 0.0)
	refundLeft := tradeblocks.NewRefundLeftBlock(swap, address[0])
	refundRight := tradeblocks.NewRefundRightBlock(refundLeft, send2, address[1])

	err := i.SignBlock(key[0])
	if err != nil {
		t.Fatal(err)
	}

	err = i2.SignBlock(key[1])
	if err != nil {
		t.Fatal(err)
	}

	err = send.SignBlock(key[0])
	if err != nil {
		t.Fatal(err)
	}

	err = send2.SignBlock(key[1])
	if err != nil {
		t.Fatal(err)
	}

	err = swap.SignBlock(key[2])
	if err != nil {
		t.Fatal(err)
	}

	err = refundLeft.SignBlock(key[2])
	if err != nil {
		t.Fatal(err)
	}

	err = refundRight.SignBlock(key[2])
	if err != nil {
		t.Fatal(err)
	}

	accountStore.AddBlock(i)
	accountStore.AddBlock(send)
	accountStore.AddBlock(i2)
	accountStore.AddBlock(send2)

	_, err = swapStore.AddBlock(swap, accountStore)
	if err != nil {
		t.Fatal(err)
	}
	_, err = swapStore.AddBlock(refundLeft, accountStore)
	if err != nil {
		t.Fatal(err)
	}
	_, err = swapStore.AddBlock(refundRight, accountStore)
	if err != nil {
		t.Fatal(err)
	}

	validator := NewSwapValidator(accountStore, swapStore)

	return swap, refundLeft, refundRight, send2, validator, nil
}

// only for refund-left, each action has different test
func TestSwapRefundRightValidation(t *testing.T) {
	key, address := CreateAccount(t)
	key2, address2 := CreateAccount(t)
	key3, address3 := CreateAccount(t)

	keyList := []*rsa.PrivateKey{key, key2, key3}
	addressList := []string{address, address2, address3}

	_, _, refundRight, send2, validator, errSetup := swapRefundRightSetup(keyList, addressList, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	// base success
	errVerify := validator.ValidateSwapBlock(refundRight)
	if errVerify != nil {
		t.Fatal(errVerify)
	}

	// signature verifying tests
	// Note: can't do executors on offers
	_, _, refundRight, _, validator, errSetup = swapRefundRightSetup(keyList, addressList, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	// random signature
	refundRight.Signature = "garbage"
	errVerify = validator.ValidateSwapBlock(refundRight)
	expectedError := "illegal base64 data at input byte 4"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// invalid previous
	refundRight.Previous = send2.Hash()
	errSign := refundRight.SignBlock(key3)
	if errSign != nil {
		t.Fatal(errSign)
	}
	errVerify = validator.ValidateSwapBlock(refundRight)
	expectedError = "previous must be not null"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// previous must be refund-left
	_, refundLeft, refundRight, _, validator, errSetup := swapRefundRightSetup(keyList, addressList, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	refundLeft.Action = "commit"
	errSign = refundRight.SignBlock(key3)
	if errSign != nil {
		t.Fatal(errSign)
	}
	errVerify = validator.ValidateSwapBlock(refundRight)
	expectedError = "Previous must be a refund-left"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// swap alignment
	// long list of fields that could trigger this:
	// Account, Token, ID, Left, Right, RefundLeft,
	// Counterparty, Want, Quantity, Executor, Fee
	// Note on account: used as address to get publicKey, changing will cause
	// a pem parsing error
	_, refundLeft, refundRight, send2, validator, errSetup = swapRefundRightSetup(keyList, addressList, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	refundRight.Fee = 1.0
	errSign = refundRight.SignBlock(key3)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(refundRight)
	expectedError = "Refund Right must match Refund Left fields"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// invalid counterparty send
	_, refundLeft, refundRight, send2, validator, errSetup = swapRefundRightSetup(keyList, addressList, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	refundRight.Right = badAddress
	errSign = refundRight.SignBlock(key3)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(refundRight)
	expectedError = "counterparty send not found/invalid"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// wrong address to refund
	// changing block.Account will break the decrypter, make the send Account wrong
	_, refundLeft, refundRight, send2, validator, errSetup = swapRefundRightSetup(keyList, addressList, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	send2.Account = badAddress
	errSign = refundRight.SignBlock(key3)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(refundRight)
	expectedError = "Account for refund must be same as original send account"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// wrong address to refund
	// changing block.Account will break the decrypter, make the send Account wrong
	_, refundLeft, refundRight, send2, validator, errSetup = swapRefundRightSetup(keyList, addressList, t)
	if errSetup != nil {
		t.Fatal(errSetup)
	}

	send2.Account = badAddress
	errSign = refundRight.SignBlock(key3)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(refundRight)
	expectedError = "Account for refund must be same as original send account"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}
}
