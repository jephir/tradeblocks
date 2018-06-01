package app

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/jephir/tradeblocks"
)

// something to use for address that won't break the decrypter
const badAddress = "Ci0tLS0tQkVHSU4gUlNBIFBVQkxJQyBLRVktLS0tLQpNSUdmTUEwR0NTcUdTSWIzRFFFQkFRVUFBNEdOQURDQmlRS0JnUURWbGVZUStNT0doSFZ2a216Q2tKcmpJNUNMCjROTUh3TlJsN1NSbkVsRkkyK25XallNRXdTT2xwNXBUY0hCempSaEpPeDFTYkx0aUtSS0ZnMVE5d1Vldk5lV1MKUE1qQjFsK0xXbVVUUnFOVGNBUFFjMFZkZXVtanFzMVArZUhFUmZrOU13cU5zclB5dHZHd3ZOUUowNVBrZ0xTawpYdTU4a3I1aVh4TUFCSXVrYlFJREFRQUIKLS0tLS1FTkQgUlNBIFBVQkxJQyBLRVktLS0tLQo"

var key, err = rsa.GenerateKey(rand.Reader, 512)

func openSetup() (*tradeblocks.AccountBlock, *tradeblocks.AccountBlock, AccountBlockValidator, error) {
	// keys
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, nil, nil, err
	}
	p := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	pubKeyReader := bytes.NewReader(p)
	address, errKey := PublicKeyToAddress(pubKeyReader)
	if err != nil {
		return nil, nil, nil, errKey
	}

	s := NewBlockStore()
	i := tradeblocks.NewIssueBlock(address, 100.0)
	send := tradeblocks.NewSendBlock(i, address, 100.0)

	pubKeyReader = bytes.NewReader(p)
	open, errOpen := Open(pubKeyReader, send, 100.0)
	if errOpen != nil {
		return nil, nil, nil, errOpen
	}

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

	// test for success
	open, _, validator, errSetup := openSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	errValidate := validator.ValidateAccountBlock(open)
	if errValidate != nil {
		t.Error(errValidate)
	}

	// test for previous not null
	open, _, validator, errPrevNull := openSetup()
	if errPrevNull != nil {
		t.Error(errPrevNull)
	}
	open.Previous = open.Link

	errSignOpen := open.SignBlock(key)
	if errSignOpen != nil {
		t.Fatal(errSignOpen)
	}

	err := validator.ValidateAccountBlock(open)
	expectedError := "previous field was not null"
	if err == nil || err.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test for send not existing
	open, _, validator, errSendNull := openSetup()
	if errSendNull != nil {
		t.Error(errSendNull)
	}
	open.Link = ""
	errSignOpen = open.SignBlock(key)
	if errSignOpen != nil {
		t.Fatal(errSignOpen)
	}

	err = validator.ValidateAccountBlock(open)
	expectedError = "link field references invalid block"
	if err == nil || err.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test for send prev not existing
	open, send, validator, errSendPrevNull := openSetup()
	if errSendPrevNull != nil {
		t.Error(errSendPrevNull)
	}
	send.Previous = ""

	err = validator.ValidateAccountBlock(open)
	expectedError = "send has no previous"
	if err == nil || err.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test for send not having necessary balance
	open, send, validator, errSendBalance := openSetup()
	if errSendBalance != nil {
		t.Error(errSendBalance)
	}
	send.Balance = 51

	err = validator.ValidateAccountBlock(open)
	expectedError = "balance does not match"
	if err == nil || err.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test for send not referencing this open
	open, send, validator, errSendNoLink := openSetup()
	if errSendNoLink != nil {
		t.Error(errSendNoLink)
	}
	send.Link = "WRONG_ACCOUNT"

	err = validator.ValidateAccountBlock(open)
	expectedError = "send block does not reference this account"
	if err == nil || err.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}
}

func issueSetup() (*tradeblocks.AccountBlock, AccountBlockValidator, error) {
	// keys
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	p := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	pubKeyReader := bytes.NewReader(p)

	issue, errIssue := Issue(pubKeyReader, 100)
	s := NewBlockStore()

	if errIssue != nil {
		return nil, nil, errIssue
	}

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
	// test for success
	issue, validator, errSetup := issueSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	errValidate := validator.ValidateAccountBlock(issue)
	if errValidate != nil {
		t.Error(errValidate)
	}
}

func sendSetup() (*tradeblocks.AccountBlock, AccountBlockValidator, error) {
	// keys
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return new(tradeblocks.AccountBlock), *new(AccountBlockValidator), err
	}
	p := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	pubKeyReader := bytes.NewReader(p)
	address, errKey := PublicKeyToAddress(pubKeyReader)
	if err != nil {
		return new(tradeblocks.AccountBlock), *new(AccountBlockValidator), errKey
	}

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
	// test for success
	send, validator, errSetup := sendSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	errValidate := validator.ValidateAccountBlock(send)
	if errValidate != nil {
		t.Error(errValidate)
	}

	// test for previous not null
	send, validator, errPrevNull := sendSetup()
	if errPrevNull != nil {
		t.Error(errPrevNull)
	}
	send.Previous = ""
	errSignSend := send.SignBlock(key)
	if errSignSend != nil {
		t.Fatal(errSignSend)
	}

	err := validator.ValidateAccountBlock(send)
	expectedError := "previous field was invalid"
	if err == nil || err.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}
}

func receiveSetup() (*tradeblocks.AccountBlock, *tradeblocks.AccountBlock, AccountBlockValidator, error) {
	// keys
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, nil, nil, err
	}
	p := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	pubKeyReader := bytes.NewReader(p)
	address, errKey := PublicKeyToAddress(pubKeyReader)
	if err != nil {
		return nil, nil, nil, errKey
	}

	s := NewBlockStore()

	i := tradeblocks.NewIssueBlock(address, 100.0)

	send := tradeblocks.NewSendBlock(i, address, 50.0)

	i2 := tradeblocks.NewIssueBlock(address, 100.0)

	receive := tradeblocks.NewReceiveBlock(i2, send, 50)

	validator := NewReceiveValidator(s)

	errSignIssue := i.SignBlock(key)
	if errSignIssue != nil {
		return nil, nil, nil, errSignIssue
	}

	errSignIssue = i2.SignBlock(key)
	if errSignIssue != nil {
		return nil, nil, nil, errSignIssue
	}

	errSignSend := send.SignBlock(key)
	if errSignSend != nil {
		return nil, nil, nil, errSignSend
	}

	errSignReceive := receive.SignBlock(key)
	if errSignReceive != nil {
		return nil, nil, nil, errSignReceive
	}

	if _, err := s.AddBlock(i); err != nil {
		return nil, nil, nil, err
	}

	if _, err := s.AddBlock(send); err != nil {
		return nil, nil, nil, err
	}
	if _, err := s.AddBlock(i2); err != nil {
		return nil, nil, nil, err
	}

	return receive, send, validator, nil
}

func TestReceiveBlockValidator(t *testing.T) {
	// test for success
	receive, _, validator, errSetup := receiveSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	errValidate := validator.ValidateAccountBlock(receive)
	if errValidate != nil {
		t.Error(errValidate)
	}

	// test for previous not null
	receive, _, validator, errPrevNull := receiveSetup()
	if errPrevNull != nil {
		t.Error(errPrevNull)
	}

	receive.Previous = ""
	errSignReceive := receive.SignBlock(key)
	if errSignReceive != nil {
		t.Fatal(errSignReceive)
	}

	err := validator.ValidateAccountBlock(receive)
	expectedError := "previous field was invalid"
	if err == nil || err.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test for link invalid
	receive, _, validator, errLink := receiveSetup()
	if errLink != nil {
		t.Error(errLink)
	}
	receive.Link = ""
	errSignReceive = receive.SignBlock(key)
	if errSignReceive != nil {
		t.Fatal(errSignReceive)
	}

	err = validator.ValidateAccountBlock(receive)
	expectedError = "link field references invalid block"
	if err == nil || err.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test for link previous invalid
	receive, send, validator, errLinkPrev := receiveSetup()
	if errLinkPrev != nil {
		t.Error(errLinkPrev)
	}
	send.Previous = ""
	errSignReceive = receive.SignBlock(key)
	if errSignReceive != nil {
		t.Fatal(errSignReceive)
	}

	err = validator.ValidateAccountBlock(receive)
	expectedError = "link field's previous references invalid block"
	if err == nil || err.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test for send linked to this
	receive, send, validator, errSendLink := receiveSetup()
	if errSendLink != nil {
		t.Error(errSendLink)
	}
	send.Previous = ""
	errSignSend := send.SignBlock(key)
	if errSignSend != nil {
		t.Fatal(errSignSend)
	}

	err = validator.ValidateAccountBlock(receive)
	expectedError = "link field's previous references invalid block"
	if err == nil || err.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}
}

func TestUnkownAction(t *testing.T) {
	err := ValidateAccountBlock(nil, &tradeblocks.AccountBlock{})
	if err.Error() != "blockvalidator: unknown action ''" {
		t.Fatalf("expected an error but got %s", err.Error())
	}
}

func swapOfferSetup() (*tradeblocks.SwapBlock, *tradeblocks.SwapBlock, *tradeblocks.AccountBlock, *SwapBlockValidator, error) {
	// keys
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	p := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	pubKeyReader := bytes.NewReader(p)
	address, err := PublicKeyToAddress(pubKeyReader)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	accountStore := NewBlockStore()
	swapStore := NewSwapBlockStore()

	i := tradeblocks.NewIssueBlock(address, 100.0)
	send := tradeblocks.NewSendBlock(i, address, 50.0)
	i2 := tradeblocks.NewIssueBlock(address, 50.0)
	send2 := tradeblocks.NewSendBlock(i2, address, 10.0)
	swap := tradeblocks.NewOfferBlock(address, send, "test-ID", "counterparty", address, 10.0, "", 0.0)
	swap2 := tradeblocks.NewCommitBlock(send2, swap)

	err = i.SignBlock(key)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	err = i2.SignBlock(key)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	err = send.SignBlock(key)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	err = send2.SignBlock(key)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	err = swap.SignBlock(key)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	err = swap2.SignBlock(key)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	_, err = accountStore.AddBlock(i)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	_, err = accountStore.AddBlock(send)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	_, err = accountStore.AddBlock(i2)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	_, err = accountStore.AddBlock(send2)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	_, err = swapStore.AddBlock(swap, accountStore)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	_, err = swapStore.AddBlock(swap2, accountStore)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	validator := NewSwapValidator(accountStore, swapStore)

	return swap, swap2, send, validator, nil
}

// only for offer, each action has different test
func TestSwapOfferValidation(t *testing.T) {
	swap, _, _, validator, errSetup := swapOfferSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	// base success
	errVerify := validator.ValidateSwapBlock(swap)
	if errVerify != nil {
		t.Fatal(errVerify)
	}

	// signature verifying tests
	// Note: can't do executors on offers
	swap, _, _, validator, errSetup = swapOfferSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	// random signature
	swap.Signature = "garbage"
	errVerify = validator.ValidateSwapBlock(swap)
	expectedError := "illegal base64 data at input byte 4"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// different key signature
	swap, _, _, validator, errSetup = swapOfferSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}
	newKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		t.Fatal(err)
	}
	errSign := swap.SignBlock(newKey)
	if errSign != nil {
		t.Fatal(errSign)
	}
	errVerify = validator.ValidateSwapBlock(swap)
	expectedError = "crypto/rsa: verification error"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// is there a previous? Hint: there shouldn't be
	swap, swap2, _, validator, errSetup := swapOfferSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	swap.Previous = swap2.Hash()
	errSign = swap.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}
	errVerify = validator.ValidateSwapBlock(swap)
	expectedError = "prev and right must be null together"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// make sure there's an account block from the left field
	swap, _, _, validator, errSetup = swapOfferSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	// no block
	swap.Left = "not a block"
	errSign = swap.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}
	errVerify = validator.ValidateSwapBlock(swap)
	expectedError = "link field references invalid block"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	//not a send
	tempI := tradeblocks.NewIssueBlock("address", 100.0)
	swap.Left = tempI.Hash()
	errSign = swap.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}
	errVerify = validator.ValidateSwapBlock(swap)
	expectedError = "link field references invalid block"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}
}

// only for commit, each action has different test
func TestSwapCommitValidation(t *testing.T) {
	_, swap2, _, validator, errSetup := swapOfferSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	// base success
	errVerify := validator.ValidateSwapBlock(swap2)
	if errVerify != nil {
		t.Fatal(errVerify)
	}

	// signature verifying tests
	// Note: can't do executors on offers
	_, swap2, send, validator, errSetup := swapOfferSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	// random signature
	swap2.Signature = "garbage"
	errVerify = validator.ValidateSwapBlock(swap2)
	expectedError := "illegal base64 data at input byte 4"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	swap2.Previous = send.Hash()
	errSign := swap2.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}
	errVerify = validator.ValidateSwapBlock(swap2)
	expectedError = "previous must be not null"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// bad account decoding
	swap2.Account = "help"
	errSign = swap2.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(swap2)
	expectedError = "failed to parse PEM block containing the public key"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// bad executor
	swap2.Executor = "help"
	errSign = swap2.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(swap2)
	expectedError = "failed to parse PEM block containing the public key"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// swaps not same fields
	_, swap2, _, validator, errSetup = swapOfferSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}
	// long list of fields that could trigger this:
	// Account, Token, ID, Left, RefundLeft, RefundRight,
	// Counterparty, Want, Quantity, Executor, Fee
	// Note on account: used as address to get publicKey, changing will cause
	// a pem parsing error
	swap2.ID = "help"
	errSign = swap2.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(swap2)
	expectedError = "Counterparty swap has incorrect fields: must match originating swap"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// original send doesn't exist (linked to bad swap)
	swap, swap2, _, validator, errSetup := swapOfferSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	swap.Left = badAddress
	swap2.Left = badAddress
	errSign = swap2.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(swap2)
	expectedError = "originating send not found"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// counter send doesn't exist
	swap, swap2, _, validator, errSetup = swapOfferSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	swap2.Right = badAddress
	errSign = swap2.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(swap2)
	expectedError = "counter send not found"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// counter send doesn't exist
	swap, swap2, _, validator, errSetup = swapOfferSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	send2, err := validator.accountChain.GetBlock(swap2.Right)
	if err != nil || send2 == nil {
		t.Fatal(err)
	}
	send2.Previous = badAddress
	errSign = send2.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(swap2)
	expectedError = "counter send prev not found"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// counter send doesn't exist
	swap, swap2, _, validator, errSetup = swapOfferSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	send2, err = validator.accountChain.GetBlock(swap2.Right)
	if err != nil || send2 == nil {
		t.Fatal(err)
	}
	send2.Balance = 0
	errSign = send2.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(swap2)
	expectedError = "amount/token requested not sent"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}
}

func swapRefundLeftSetup() (*tradeblocks.SwapBlock, *tradeblocks.SwapBlock, *tradeblocks.AccountBlock, *SwapBlockValidator, error) {
	// keys
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	p := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	pubKeyReader := bytes.NewReader(p)
	address, errKey := PublicKeyToAddress(pubKeyReader)
	if err != nil {
		return nil, nil, nil, nil, errKey
	}

	accountStore := NewBlockStore()
	swapStore := NewSwapBlockStore()

	i := tradeblocks.NewIssueBlock(address, 100.0)
	send := tradeblocks.NewSendBlock(i, address, 50.0)
	i2 := tradeblocks.NewIssueBlock(address, 50)
	swap := tradeblocks.NewOfferBlock(address, send, "test-ID", "counterparty", address, 10.0, "", 0.0)
	refundLeft := tradeblocks.NewRefundLeftBlock(swap, address)

	errSignIssue := i.SignBlock(key)
	if errSignIssue != nil {
		return nil, nil, nil, nil, errSignIssue
	}

	errSignIssue = i2.SignBlock(key)
	if errSignIssue != nil {
		return nil, nil, nil, nil, errSignIssue
	}

	errSignSend := send.SignBlock(key)
	if errSignSend != nil {
		return nil, nil, nil, nil, errSignSend
	}

	errSignSwap := swap.SignBlock(key)
	if errSignSwap != nil {
		return nil, nil, nil, nil, errSignSwap
	}

	errSignRefund := refundLeft.SignBlock(key)
	if errSignRefund != nil {
		return nil, nil, nil, nil, errSignRefund
	}

	accountStore.AddBlock(i)
	accountStore.AddBlock(send)
	accountStore.AddBlock(i2)

	_, err = swapStore.AddBlock(swap, accountStore)
	if err != nil {
		panic(err)
	}
	_, err = swapStore.AddBlock(refundLeft, accountStore)
	if err != nil {
		panic(err)
	}

	validator := NewSwapValidator(accountStore, swapStore)

	return swap, refundLeft, send, validator, nil
}

// only for refund-left, each action has different test
func TestSwapRefundLeftValidation(t *testing.T) {
	_, refundLeft, _, validator, errSetup := swapRefundLeftSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	// base success
	errVerify := validator.ValidateSwapBlock(refundLeft)
	if errVerify != nil {
		t.Fatal(errVerify)
	}

	// signature verifying tests
	// Note: can't do executors on offers
	_, refundLeft, send, validator, errSetup := swapRefundLeftSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	// random signature
	refundLeft.Signature = "garbage"
	errVerify = validator.ValidateSwapBlock(refundLeft)
	expectedError := "illegal base64 data at input byte 4"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// invalid previous
	refundLeft.Previous = send.Hash()
	errSign := refundLeft.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}
	errVerify = validator.ValidateSwapBlock(refundLeft)
	expectedError = "previous must be not null"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// long list of fields that could trigger this:
	// Account, Token, ID, Left, Right, RefundRight,
	// Counterparty, Want, Quantity, Executor, Fee
	// Note on account: used as address to get publicKey, changing will cause
	// a pem parsing error
	swap, refundLeft, send, validator, errSetup := swapRefundLeftSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	refundLeft.Fee = 1.0
	errSign = refundLeft.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(refundLeft)
	expectedError = "Counterparty swap has incorrect fields: must match originating swap"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// originating send invalid
	swap, refundLeft, send, validator, errSetup = swapRefundLeftSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	swap.Left = badAddress
	refundLeft.Left = badAddress
	errSign = refundLeft.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(refundLeft)
	expectedError = "Originating send is invalid or not found"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// refundleft is not the initiators account
	swap, refundLeft, send, validator, errSetup = swapRefundLeftSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	refundLeft.RefundLeft = badAddress
	errSign = refundLeft.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(refundLeft)
	expectedError = "Refund must be to initiator's account"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}
}

func swapRefundRightSetup() (*tradeblocks.SwapBlock, *tradeblocks.SwapBlock, *tradeblocks.SwapBlock, *tradeblocks.AccountBlock, *SwapBlockValidator, error) {
	// keys
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	p := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	pubKeyReader := bytes.NewReader(p)
	address, err := PublicKeyToAddress(pubKeyReader)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	accountStore := NewBlockStore()
	swapStore := NewSwapBlockStore()

	i := tradeblocks.NewIssueBlock(address, 100.0)
	send := tradeblocks.NewSendBlock(i, address, 50.0)
	i2 := tradeblocks.NewIssueBlock(address, 50)
	send2 := tradeblocks.NewSendBlock(i2, address, 10.0)
	swap := tradeblocks.NewOfferBlock(address, send, "test-ID", "counterparty", address, 10.0, "", 0.0)
	refundLeft := tradeblocks.NewRefundLeftBlock(swap, address)
	refundRight := tradeblocks.NewRefundRightBlock(refundLeft, send2, address)

	err = i.SignBlock(key)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	err = i2.SignBlock(key)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	err = send.SignBlock(key)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	err = send2.SignBlock(key)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	err = swap.SignBlock(key)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	err = refundLeft.SignBlock(key)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	err = refundRight.SignBlock(key)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	accountStore.AddBlock(i)
	accountStore.AddBlock(send)
	accountStore.AddBlock(i2)
	accountStore.AddBlock(send2)

	_, err = swapStore.AddBlock(swap, accountStore)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	_, err = swapStore.AddBlock(refundLeft, accountStore)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	_, err = swapStore.AddBlock(refundRight, accountStore)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	validator := NewSwapValidator(accountStore, swapStore)

	return swap, refundLeft, refundRight, send2, validator, nil
}

// only for refund-left, each action has different test
func TestSwapRefundRightValidation(t *testing.T) {
	_, _, refundRight, send2, validator, errSetup := swapRefundRightSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	// base success
	errVerify := validator.ValidateSwapBlock(refundRight)
	if errVerify != nil {
		t.Fatal(errVerify)
	}

	// signature verifying tests
	// Note: can't do executors on offers
	_, _, refundRight, _, validator, errSetup = swapRefundRightSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	// random signature
	refundRight.Signature = "garbage"
	errVerify = validator.ValidateSwapBlock(refundRight)
	expectedError := "illegal base64 data at input byte 4"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// invalid previous
	refundRight.Previous = send2.Hash()
	errSign := refundRight.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}
	errVerify = validator.ValidateSwapBlock(refundRight)
	expectedError = "previous must be not null"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// previous must be refund-left
	_, refundLeft, refundRight, _, validator, errSetup := swapRefundRightSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	refundLeft.Action = "commit"
	errSign = refundRight.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}
	errVerify = validator.ValidateSwapBlock(refundRight)
	expectedError = "Previous must be a refund-left"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// swap alignment
	// long list of fields that could trigger this:
	// Account, Token, ID, Left, Right, RefundLeft,
	// Counterparty, Want, Quantity, Executor, Fee
	// Note on account: used as address to get publicKey, changing will cause
	// a pem parsing error
	_, refundLeft, refundRight, send2, validator, errSetup = swapRefundRightSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	refundRight.Fee = 1.0
	errSign = refundRight.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(refundRight)
	expectedError = "Counterparty swap has incorrect fields: must match originating swap"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// invalid counterparty send
	_, refundLeft, refundRight, send2, validator, errSetup = swapRefundRightSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	refundRight.Right = badAddress
	errSign = refundRight.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(refundRight)
	expectedError = "counterparty send not found/invalid"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// wrong address to refund
	// changing block.Account will break the decrypter, make the send Account wrong
	_, refundLeft, refundRight, send2, validator, errSetup = swapRefundRightSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	send2.Account = badAddress
	errSign = refundRight.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(refundRight)
	expectedError = "Account for refund must be same as original send account"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}

	// wrong address to refund
	// changing block.Account will break the decrypter, make the send Account wrong
	_, refundLeft, refundRight, send2, validator, errSetup = swapRefundRightSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	send2.Account = badAddress
	errSign = refundRight.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify = validator.ValidateSwapBlock(refundRight)
	expectedError = "Account for refund must be same as original send account"
	if errVerify == nil || errVerify.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", errVerify, expectedError)
	}
}
