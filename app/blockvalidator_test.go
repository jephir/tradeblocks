package app

import (
	"io"
	"testing"

	"github.com/jephir/tradeblocks"
)

func openSetup() (*tradeblocks.AccountBlock, *tradeblocks.AccountBlock, AccountBlockValidator, error) {
	publicKey.Seek(0, io.SeekStart)
	address, errKey := PublicKeyToAddress(publicKey)
	if errKey != nil {
		return new(tradeblocks.AccountBlock), new(tradeblocks.AccountBlock), *new(AccountBlockValidator), errKey
	}

	s := NewBlockStore()
	i := tradeblocks.NewIssueBlock("xtb:test", 100.0)
	s.AddBlock(i)
	send := tradeblocks.NewSendBlock(i, address, 100.0)
	s.AddBlock(send)

	publicKey.Seek(0, io.SeekStart)
	open, errOpen := Open(publicKey, send, 100.0)
	if errOpen != nil {
		return new(tradeblocks.AccountBlock), new(tradeblocks.AccountBlock), *new(AccountBlockValidator), errKey
	}

	validator := NewOpenValidator(s)

	return open, send, validator, nil
}

func TestOpenBlockValidator(t *testing.T) {

	// test for success
	open, _, validator, errSetup := openSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	err := validator.ValidateAccountBlock(open)
	if err != nil {
		t.Error(err)
	}

	// test for previous not null
	open, _, validator, errPrevNull := openSetup()
	if errPrevNull != nil {
		t.Error(errPrevNull)
	}
	open.Previous = open.Link

	err = validator.ValidateAccountBlock(open)
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
	publicKey.Seek(0, io.SeekStart)
	issue, errIssue := Issue(publicKey, 100)
	s := NewBlockStore()
	s.AddBlock(issue)

	if errIssue != nil {
		return new(tradeblocks.AccountBlock), *new(AccountBlockValidator), errIssue
	}

	validator := NewIssueValidator(s)

	return issue, validator, nil
}

func TestIssueBlockValidator(t *testing.T) {

	// test for success
	issue, validator, errSetup := issueSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	err := validator.ValidateAccountBlock(issue)
	if err != nil {
		t.Error(err)
	}
}

func sendSetup() (*tradeblocks.AccountBlock, AccountBlockValidator, error) {
	s := NewBlockStore()
	address, errKey := PublicKeyToAddress(publicKey)
	if errKey != nil {
		return new(tradeblocks.AccountBlock), *new(AccountBlockValidator), errKey
	}

	i := tradeblocks.NewIssueBlock("xtb:test", 100.0)
	s.AddBlock(i)
	send := tradeblocks.NewSendBlock(i, address, 100.0)
	s.AddBlock(send)

	validator := NewSendValidator(s)

	return send, validator, nil
}

func TestSendBlockValidator(t *testing.T) {

	// test for success
	send, validator, errSetup := sendSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	err := validator.ValidateAccountBlock(send)
	if err != nil {
		t.Error(err)
	}

	// test for previous not null
	send, validator, errPrevNull := sendSetup()
	if errPrevNull != nil {
		t.Error(errPrevNull)
	}
	send.Previous = ""

	err = validator.ValidateAccountBlock(send)
	expectedError := "previous field was invalid"
	if err == nil || err.Error() != expectedError {
		t.Errorf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}
}

func receiveSetup() (*tradeblocks.AccountBlock, *tradeblocks.AccountBlock, AccountBlockValidator, error) {
	s := NewBlockStore()

	i := tradeblocks.NewIssueBlock("xtb:initiator", 100.0)
	s.AddBlock(i)
	send := tradeblocks.NewSendBlock(i, "xtb:target", 50.0)
	s.AddBlock(send)

	i2 := tradeblocks.NewIssueBlock("xtb:target", 100.0)
	s.AddBlock(i2)

	publicKey.Seek(0, io.SeekStart)
	receive := tradeblocks.NewReceiveBlock(i2, send, 50)

	validator := NewReceiveValidator(s)

	return receive, send, validator, nil
}

func TestReceiveBlockValidator(t *testing.T) {
	// test for success
	receive, _, validator, errSetup := receiveSetup()
	if errSetup != nil {
		t.Error(errSetup)
	}

	err := validator.ValidateAccountBlock(receive)
	if err != nil {
		t.Error(err)
	}

	// test for previous not null
	receive, _, validator, errPrevNull := receiveSetup()
	if errPrevNull != nil {
		t.Error(errPrevNull)
	}
	receive.Previous = ""

	err = validator.ValidateAccountBlock(receive)
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
