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

var key, err = rsa.GenerateKey(rand.Reader, 512)

func openSetup() (*tradeblocks.AccountBlock, *tradeblocks.AccountBlock, AccountBlockValidator, error) {
	// keys
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return new(tradeblocks.AccountBlock), new(tradeblocks.AccountBlock), *new(AccountBlockValidator), err
	}
	p := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	pubKeyReader := bytes.NewReader(p)
	address, errKey := PublicKeyToAddress(pubKeyReader)
	if err != nil {
		return new(tradeblocks.AccountBlock), new(tradeblocks.AccountBlock), *new(AccountBlockValidator), errKey
	}

	s := NewBlockStore()
	i := tradeblocks.NewIssueBlock(address, 100.0)
	s.AddBlock(i)
	send := tradeblocks.NewSendBlock(i, address, 100.0)
	s.AddBlock(send)

	pubKeyReader = bytes.NewReader(p)
	open, errOpen := Open(pubKeyReader, send, 100.0)
	if errOpen != nil {
		return new(tradeblocks.AccountBlock), new(tradeblocks.AccountBlock), *new(AccountBlockValidator), errOpen
	}

	validator := NewOpenValidator(s)

	errSignIssue := i.SignBlock(key)
	if errSignIssue != nil {
		return new(tradeblocks.AccountBlock), new(tradeblocks.AccountBlock), *new(AccountBlockValidator), errSignIssue
	}

	errSignSend := send.SignBlock(key)
	if errSignSend != nil {
		return new(tradeblocks.AccountBlock), new(tradeblocks.AccountBlock), *new(AccountBlockValidator), errSignSend
	}

	errSignOpen := open.SignBlock(key)
	if errSignOpen != nil {
		return new(tradeblocks.AccountBlock), new(tradeblocks.AccountBlock), *new(AccountBlockValidator), errSignOpen
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
		return new(tradeblocks.AccountBlock), *new(AccountBlockValidator), err
	}
	p := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	pubKeyReader := bytes.NewReader(p)

	issue, errIssue := Issue(pubKeyReader, 100)
	s := NewBlockStore()
	s.AddBlock(issue)

	if errIssue != nil {
		return new(tradeblocks.AccountBlock), *new(AccountBlockValidator), errIssue
	}

	validator := NewIssueValidator(s)

	errSignIssue := issue.SignBlock(key)
	if errSignIssue != nil {
		return new(tradeblocks.AccountBlock), *new(AccountBlockValidator), errSignIssue
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
	i := tradeblocks.NewIssueBlock(address, 100.0)
	s.AddBlock(i)
	send := tradeblocks.NewSendBlock(i, address, 100.0)
	s.AddBlock(send)

	validator := NewSendValidator(s)

	errSignSend := send.SignBlock(key)
	if errSignSend != nil {
		return new(tradeblocks.AccountBlock), *new(AccountBlockValidator), errSignSend
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
		return new(tradeblocks.AccountBlock), new(tradeblocks.AccountBlock), *new(AccountBlockValidator), err
	}
	p := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	pubKeyReader := bytes.NewReader(p)
	address, errKey := PublicKeyToAddress(pubKeyReader)
	if err != nil {
		return new(tradeblocks.AccountBlock), new(tradeblocks.AccountBlock), *new(AccountBlockValidator), errKey
	}

	s := NewBlockStore()

	i := tradeblocks.NewIssueBlock(address, 100.0)
	s.AddBlock(i)
	send := tradeblocks.NewSendBlock(i, address, 50.0)
	s.AddBlock(send)

	i2 := tradeblocks.NewIssueBlock(address, 100.0)
	s.AddBlock(i2)

	receive := tradeblocks.NewReceiveBlock(i2, send, 50)

	validator := NewReceiveValidator(s)

	errSignIssue := i.SignBlock(key)
	if errSignIssue != nil {
		return new(tradeblocks.AccountBlock), new(tradeblocks.AccountBlock), *new(AccountBlockValidator), errSignIssue
	}

	errSignIssue = i2.SignBlock(key)
	if errSignIssue != nil {
		return new(tradeblocks.AccountBlock), new(tradeblocks.AccountBlock), *new(AccountBlockValidator), errSignIssue
	}

	errSignSend := send.SignBlock(key)
	if errSignSend != nil {
		return new(tradeblocks.AccountBlock), new(tradeblocks.AccountBlock), *new(AccountBlockValidator), errSignSend
	}

	errSignReceive := receive.SignBlock(key)
	if errSignReceive != nil {
		return new(tradeblocks.AccountBlock), new(tradeblocks.AccountBlock), *new(AccountBlockValidator), errSignReceive
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
