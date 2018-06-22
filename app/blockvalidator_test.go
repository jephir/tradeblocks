package app

import (
	"crypto/rsa"
	"strings"
	"testing"

	"github.com/jephir/tradeblocks"
)

// something to use for address that won't break the decrypter
const badAddress = "xtb:MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCz8_JKRKFLWxECvLqxyBt3L95IzqtFcxPMbXeCwGUzi0MNuG7Z_WxNfrKJZEZnepL_KzmFaa9gwXQ1GmPmFHShKjk7eqhglOyO6pPWWoCvxywY3hqVFvVE4hUchvzUTZNUSu1Pr-TOFzicYU4zXnzmh7r7cD4xC1N5CAIHtjAXKQIDAQAB"
const badSignature = "zypz3O++osG8rj3R9jirIvhZGwTtwwEZ16LTc3BdjUiJ6w5pBKd1JPDamXbSyIjCIC9wGTZx14IDVHLpL/wW8W8GI3d2ocsFfPolo81Wrgu1HN5ciklQ3Ph1MaxxO8/KND64k9cwG5EjH79M4vtiJMcl4EPM4BZ00yrcNxwSkik="

func openSetup(key *rsa.PrivateKey, address string, t *testing.T) (*tradeblocks.AccountBlock, *tradeblocks.AccountBlock, AccountBlockValidator, error) {
	s := NewBlockStore()
	i := tradeblocks.NewIssueBlock(address, 100.0)
	send := tradeblocks.NewSendBlock(i, address, 100.0)

	open := tradeblocks.NewOpenBlockFromSend(address, send, 100.0)

	validator := NewOpenValidator(s)

	err := i.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = send.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = open.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.AddAccountBlock(i); err != nil {
		t.Fatal("Block not found")
	}

	if err := s.AddAccountBlock(send); err != nil {
		t.Fatal("Block not found")
	}
	return open, send, validator, nil
}

func TestOpenBlockValidator(t *testing.T) {
	key, address := CreateAccount(t)
	// test for success
	open, send, validator, err := openSetup(key, address, t)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(open)
	if err != nil {
		t.Fatal(err)
	}

	// test for previous not null
	open, _, validator, errPrevNull := openSetup(key, address, t)
	if errPrevNull != nil {
		t.Fatal(errPrevNull)
	}
	open.Previous = open.Link

	err = open.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(open)
	expectedError := "previous field was not null"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test for send not existing
	open, _, validator, err = openSetup(key, address, t)
	if err != nil {
		t.Fatal(err)
	}
	open.Link = ""
	err = open.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(open)
	expectedError = "link field references invalid block"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test for send prev not existing
	t.Skip("TODO Block has to be modified before added to blockstore")
	open, send, validator, err = openSetup(key, address, t)
	if err != nil {
		t.Fatal(err)
	}
	send.Previous = ""
	err = send.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(open)
	expectedError = "send has no previous"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test for send not having necessary balance
	open, send, validator, err = openSetup(key, address, t)
	if err != nil {
		t.Fatal(err)
	}
	send.Balance = 51
	err = send.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(open)
	expectedError = "balance expected 100.000000; got 49.000000"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test for send not referencing this open
	open, send, validator, err = openSetup(key, address, t)
	if err != nil {
		t.Fatal(err)
	}
	send.Link = "WRONG_ACCOUNT"
	err = send.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(open)
	expectedError = "send link 'WRONG_ACCOUNT' does not reference account"
	if err == nil || !strings.Contains(err.Error(), expectedError) {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}
}

func openFromSwapSetup(key []*rsa.PrivateKey, address []string, t *testing.T) (*tradeblocks.AccountBlock, *tradeblocks.AccountBlock, *tradeblocks.AccountBlock, *tradeblocks.SwapBlock, *OpenBlockValidator, error) {
	s := NewBlockStore()

	i := tradeblocks.NewIssueBlock(address[0], 100.0)
	i2 := tradeblocks.NewIssueBlock(address[1], 100.0)

	send1 := tradeblocks.NewSendBlock(i, address[0]+":swap:test-ID", 50.0)
	send2 := tradeblocks.NewSendBlock(i2, address[0]+":swap:test-ID", 10.0)

	swap := tradeblocks.NewOfferBlock(address[0], send1, "test-ID", address[1], address[1], 10.0, "", 0.0)
	swap2 := tradeblocks.NewCommitBlock(swap, send2)

	open2 := tradeblocks.NewOpenBlockFromSwap(address[0], address[1], swap2, 10)
	open3 := tradeblocks.NewOpenBlockFromSwap(address[1], address[0], swap2, 50)

	if err := i.SignBlock(key[0]); err != nil {
		t.Fatal(err)
	}
	if err := i2.SignBlock(key[1]); err != nil {
		t.Fatal(err)
	}
	if err := send1.SignBlock(key[0]); err != nil {
		t.Fatal(err)
	}
	if err := send2.SignBlock(key[1]); err != nil {
		t.Fatal(err)
	}
	if err := swap.SignBlock(key[0]); err != nil {
		t.Fatal(err)
	}
	if err := swap2.SignBlock(key[1]); err != nil {
		t.Fatal(err)
	}
	if err := open2.SignBlock(key[0]); err != nil {
		t.Fatal(err)
	}
	if err := open3.SignBlock(key[1]); err != nil {
		t.Fatal(err)
	}

	if err := s.AddAccountBlock(i); err != nil {
		t.Fatal(err)
	}
	if err := s.AddAccountBlock(i2); err != nil {
		t.Fatal(err)
	}
	if err := s.AddAccountBlock(send1); err != nil {
		t.Fatal(err)
	}
	if err := s.AddAccountBlock(send2); err != nil {
		t.Fatal(err)
	}
	if err := s.AddSwapBlock(swap); err != nil {
		t.Fatal(err)
	}
	if err := s.AddSwapBlock(swap2); err != nil {
		t.Fatal(err)
	}
	if err := s.AddAccountBlock(open2); err != nil {
		t.Fatal(err)
	}
	if err := s.AddAccountBlock(open3); err != nil {
		t.Fatal(err)
	}

	validator := NewOpenValidator(s)

	return open2, open3, send1, swap2, validator, nil
}

func TestOpenFromSwapValidation(t *testing.T) {
	key, address := CreateAccount(t)
	key2, address2 := CreateAccount(t)
	key3, address3 := CreateAccount(t)

	keyList := []*rsa.PrivateKey{key, key2}
	addressList := []string{address, address2}
	// test for success
	open2, open3, send1, _, validator, err := openFromSwapSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	if err = validator.ValidateAccountBlock(open2); err != nil {
		t.Fatal(err)
	}

	if err = validator.ValidateAccountBlock(open3); err != nil {
		t.Fatal(err)
	}

	// test for previous not null
	open2, open3, send1, _, validator, err = openFromSwapSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}
	open2.Previous = send1.Hash()
	open3.Previous = send1.Hash()

	if err = open2.SignBlock(key); err != nil {
		t.Fatal(err)
	}
	if err = open3.SignBlock(key2); err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(open2)
	expectedError := "previous field was not null"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}
	err = validator.ValidateAccountBlock(open3)
	expectedError = "previous field was not null"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test account mismatch
	open2, open3, send1, _, validator, err = openFromSwapSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	open2.Account = address3
	open3.Account = address3

	if err = open2.SignBlock(key3); err != nil {
		t.Fatal(err)
	}
	if err = open3.SignBlock(key3); err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(open2)
	expectedError = "Account mismatch between receiver and sender"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}
	err = validator.ValidateAccountBlock(open3)
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test token mismatch
	open2, open3, send1, _, validator, err = openFromSwapSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	open2.Token = address3
	open3.Token = address3

	if err = open2.SignBlock(key); err != nil {
		t.Fatal(err)
	}
	if err = open3.SignBlock(key2); err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(open2)
	expectedError = "Can't receive different token types"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}
	err = validator.ValidateAccountBlock(open3)
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test bad balances
	open2, open3, send1, _, validator, err = openFromSwapSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	open2.Balance = 50
	open3.Balance = 10

	if err = open2.SignBlock(key); err != nil {
		t.Fatal(err)
	}
	if err = open3.SignBlock(key2); err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(open2)
	expectedError = "Mismatched balances receiving by offerer"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}
	err = validator.ValidateAccountBlock(open3)
	expectedError = "Mismatched balances receiving by committer"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}
}

func issueSetup(key *rsa.PrivateKey, address string, t *testing.T) (*tradeblocks.AccountBlock, AccountBlockValidator, error) {
	issue := tradeblocks.NewIssueBlock(address, 100)
	s := NewBlockStore()

	validator := NewIssueValidator(s)

	err := issue.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.AddAccountBlock(issue); err != nil {
		t.Fatal("block not founs")
	}

	return issue, validator, nil
}

func TestIssueBlockValidator(t *testing.T) {
	key, address := CreateAccount(t)
	// test for success
	issue, validator, err := issueSetup(key, address, t)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(issue)
	if err != nil {
		t.Fatal(err)
	}
}

func sendSetup(key *rsa.PrivateKey, address string, t *testing.T) (*tradeblocks.AccountBlock, AccountBlockValidator, error) {
	s := NewBlockStore()
	issue := tradeblocks.NewIssueBlock(address, 100.0)

	send := tradeblocks.NewSendBlock(issue, address, 100.0)

	validator := NewSendValidator(s)

	errSend := send.SignBlock(key)
	if errSend != nil {
		return nil, nil, errSend
	}

	err := issue.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.AddAccountBlock(issue); err != nil {
		t.Fatal("Block not found")
	}

	if err := s.AddAccountBlock(send); err != nil {
		t.Fatal("Block not found")
	}

	return send, validator, nil
}

func TestSendBlockValidator(t *testing.T) {
	key, address := CreateAccount(t)
	// test for success
	send, validator, err := sendSetup(key, address, t)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(send)
	if err != nil {
		t.Fatal(err)
	}

	// test for previous not null
	send, validator, errPrevNull := sendSetup(key, address, t)
	if errPrevNull != nil {
		t.Fatal(errPrevNull)
	}

	send.Previous = badAddress
	errSend := send.SignBlock(key)
	if errSend != nil {
		t.Fatal(errSend)
	}

	err = validator.ValidateAccountBlock(send)
	expectedError := "Previous block invalid"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}
}

func receiveSetup(key []*rsa.PrivateKey, address []string, t *testing.T) (*tradeblocks.AccountBlock, *tradeblocks.AccountBlock, AccountBlockValidator, error) {
	s := NewBlockStore()

	i := tradeblocks.NewIssueBlock(address[0], 100.0)
	send := tradeblocks.NewSendBlock(i, address[1], 20.0)
	send2 := tradeblocks.NewSendBlock(send, address[1], 20.0)

	open := tradeblocks.NewOpenBlockFromSend(address[1], send, 20.0)
	receive := tradeblocks.NewReceiveBlockFromSend(open, send2, 20)

	validator := NewReceiveValidator(s)

	if err := i.SignBlock(key[0]); err != nil {
		t.Fatal(err)
	}
	if err := open.SignBlock(key[1]); err != nil {
		t.Fatal(err)
	}
	if err := send.SignBlock(key[0]); err != nil {
		t.Fatal(err)
	}
	if err := send2.SignBlock(key[0]); err != nil {
		t.Fatal(err)
	}
	if err := receive.SignBlock(key[1]); err != nil {
		t.Fatal(err)
	}

	if err := s.AddAccountBlock(i); err != nil {
		t.Fatal(err)
	}
	if err := s.AddAccountBlock(send); err != nil {
		t.Fatal(err)
	}
	if err := s.AddAccountBlock(send2); err != nil {
		t.Fatal(err)
	}
	if err := s.AddAccountBlock(open); err != nil {
		t.Fatal(err)
	}
	if err := s.AddAccountBlock(receive); err != nil {
		t.Fatal(err)
	}

	return receive, send2, validator, nil
}

func TestReceiveBlockValidator(t *testing.T) {
	key, address := CreateAccount(t)
	key2, address2 := CreateAccount(t)
	key3, address3 := CreateAccount(t)

	keyList := []*rsa.PrivateKey{key, key2, key3}
	addressList := []string{address, address2, address3}

	// test for success
	receive, send2, validator, err := receiveSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(receive)
	if err != nil {
		t.Fatal(err)
	}

	// test for previous not null
	receive, _, validator, err = receiveSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	receive.Previous = badAddress
	err = receive.SignBlock(key2)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(receive)
	expectedError := "previous field was invalid"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test for link invalid
	receive, _, validator, err = receiveSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}
	receive.Link = ""
	err = receive.SignBlock(key2)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(receive)
	expectedError = "db: not found"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test for link previous invalid
	receive, send2, validator, err = receiveSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	t.Skip("TODO send2 needs to be modified before being inserted into block store")
	send2.Previous = badAddress
	err = send2.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(receive)
	expectedError = "link field's previous references invalid block"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test for send linked to this
	receive, send2, validator, err = receiveSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	send2.Previous = ""
	err = send2.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(receive)
	expectedError = "link field's previous references invalid block"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}
}

func receiveFromSwapSetup(key []*rsa.PrivateKey, address []string, t *testing.T) (*tradeblocks.AccountBlock, *tradeblocks.AccountBlock, *tradeblocks.SwapBlock, *ReceiveBlockValidator, error) {
	s := NewBlockStore()

	i := tradeblocks.NewIssueBlock(address[0], 100.0)
	send := tradeblocks.NewSendBlock(i, address[1], 20.0)
	send2 := tradeblocks.NewSendBlock(send, address[1], 20.0)

	open := tradeblocks.NewOpenBlockFromSend(address[1], send, 20.0)
	receive := tradeblocks.NewReceiveBlockFromSend(open, send2, 20)

	send3 := tradeblocks.NewSendBlock(send2, address[0]+":swap:test-ID", 50.0)
	send4 := tradeblocks.NewSendBlock(receive, address[0]+":swap:test-ID", 10.0)

	swap := tradeblocks.NewOfferBlock(address[0], send3, "test-ID", address[1], address[0], 10.0, "", 0.0)
	swap2 := tradeblocks.NewCommitBlock(swap, send4)

	receive2 := tradeblocks.NewReceiveBlockFromSwap(send3, swap2, 10)
	receive3 := tradeblocks.NewReceiveBlockFromSwap(send4, swap2, 50)

	if err := i.SignBlock(key[0]); err != nil {
		t.Fatal(err)
	}
	if err := open.SignBlock(key[1]); err != nil {
		t.Fatal(err)
	}
	if err := send.SignBlock(key[0]); err != nil {
		t.Fatal(err)
	}
	if err := send2.SignBlock(key[0]); err != nil {
		t.Fatal(err)
	}
	if err := receive.SignBlock(key[1]); err != nil {
		t.Fatal(err)
	}
	if err := send3.SignBlock(key[0]); err != nil {
		t.Fatal(err)
	}
	if err := send4.SignBlock(key[1]); err != nil {
		t.Fatal(err)
	}
	if err := swap.SignBlock(key[0]); err != nil {
		t.Fatal(err)
	}
	if err := swap2.SignBlock(key[1]); err != nil {
		t.Fatal(err)
	}
	if err := receive2.SignBlock(key[0]); err != nil {
		t.Fatal(err)
	}
	if err := receive3.SignBlock(key[1]); err != nil {
		t.Fatal(err)
	}

	if err := s.AddAccountBlock(i); err != nil {
		t.Fatal(err)
	}
	if err := s.AddAccountBlock(send); err != nil {
		t.Fatal(err)
	}
	if err := s.AddAccountBlock(send2); err != nil {
		t.Fatal(err)
	}
	if err := s.AddAccountBlock(open); err != nil {
		t.Fatal(err)
	}
	if err := s.AddAccountBlock(receive); err != nil {
		t.Fatal(err)
	}
	if err := s.AddAccountBlock(send3); err != nil {
		t.Fatal(err)
	}
	if err := s.AddAccountBlock(send4); err != nil {
		t.Fatal(err)
	}
	if err := s.AddSwapBlock(swap); err != nil {
		t.Fatal(err)
	}
	if err := s.AddSwapBlock(swap2); err != nil {
		t.Fatal(err)
	}
	if err := s.AddAccountBlock(receive2); err != nil {
		t.Fatal(err)
	}
	if err := s.AddAccountBlock(receive3); err != nil {
		t.Fatal(err)
	}

	validator := NewReceiveValidator(s)

	return receive2, receive3, swap2, validator, nil
}

func TestReceiveSwap(t *testing.T) {
	key, address := CreateAccount(t)
	key2, address2 := CreateAccount(t)
	key3, address3 := CreateAccount(t)

	keyList := []*rsa.PrivateKey{key, key2, key3}
	addressList := []string{address, address2, address3}

	// test for success
	receive2, receive3, _, validator, err := receiveFromSwapSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(receive2)
	if err != nil {
		t.Fatal(err)
	}

	// test for previous not null
	receive2, receive3, _, validator, err = receiveFromSwapSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	receive2.Previous = badAddress
	err = receive2.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(receive2)
	expectedError := "previous field was invalid"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// test for previous not null
	receive2, receive3, _, validator, err = receiveFromSwapSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	receive3.Previous = badAddress
	err = receive3.SignBlock(key2)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(receive3)
	expectedError = "previous field was invalid"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// receiver incorrect account
	receive2, receive3, _, validator, err = receiveFromSwapSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	receive3.Account = address3
	err = receive3.SignBlock(key3)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(receive3)
	expectedError = "Account mismatch between receiver and sender"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// wrong token
	receive2, receive3, _, validator, err = receiveFromSwapSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	receive3.Token = address3
	err = receive3.SignBlock(key2)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(receive3)
	expectedError = "Can't receive different token types"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// bad balances
	receive2, receive3, _, validator, err = receiveFromSwapSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	receive3.Balance = 200
	err = receive3.SignBlock(key2)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(receive3)
	expectedError = "Mismatched balances receiving by committer"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// bad balances
	receive2, receive3, _, validator, err = receiveFromSwapSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	receive2.Balance = 200
	err = receive2.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateAccountBlock(receive2)
	expectedError = "Mismatched balances receiving by offerer"
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
	blockStore := NewBlockStore()

	i := tradeblocks.NewIssueBlock(address[0], 100.0)
	send := tradeblocks.NewSendBlock(i, address[2]+":swap:test-ID", 50.0)

	i2 := tradeblocks.NewIssueBlock(address[1], 50.0)
	send2 := tradeblocks.NewSendBlock(i2, address[2]+":swap:test-ID", 10.0)

	swap := tradeblocks.NewOfferBlock(address[2], send, "test-ID", address[1], address[1], 10.0, "", 0.0)
	swap2 := tradeblocks.NewCommitBlock(swap, send2)

	if err := i.SignBlock(keys[0]); err != nil {
		t.Fatal(err)
	}

	if err := i2.SignBlock(keys[1]); err != nil {
		t.Fatal(err)
	}

	if err := send.SignBlock(keys[0]); err != nil {
		t.Fatal(err)
	}

	if err := send2.SignBlock(keys[1]); err != nil {
		t.Fatal(err)
	}

	if err := swap.SignBlock(keys[2]); err != nil {
		t.Fatal(err)
	}

	if err := swap2.SignBlock(keys[1]); err != nil {
		t.Fatal(err)
	}

	if err := blockStore.AddAccountBlock(i); err != nil {
		t.Fatal(err)
	}

	if err := blockStore.AddAccountBlock(send); err != nil {
		t.Fatal(err)
	}
	if err := blockStore.AddAccountBlock(i2); err != nil {
		t.Fatal(err)
	}

	if err := blockStore.AddAccountBlock(send2); err != nil {
		t.Fatal(err)
	}

	if err := blockStore.AddSwapBlock(swap); err != nil {
		t.Fatal(err)
	}

	if err := blockStore.AddSwapBlock(swap2); err != nil {
		t.Fatal(err)
	}

	validator := NewSwapValidator(blockStore)

	return swap, swap2, send, validator, nil
}

// only for offer, each action has different test
func TestSwapOfferValidation(t *testing.T) {
	key, address := CreateAccount(t)
	key2, address2 := CreateAccount(t)
	key3, address3 := CreateAccount(t)

	keyList := []*rsa.PrivateKey{key, key2, key3}
	addressList := []string{address, address2, address3}

	swap, swap2, _, validator, err := swapOfferSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	// base success
	err = validator.ValidateSwapBlock(swap)
	if err != nil {
		t.Fatal(err)
	}

	// signature verifying tests
	// Note: can't do executors on offers
	swap, _, _, validator, err = swapOfferSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	// random signature
	swap.Signature = "garbage"
	err = validator.ValidateSwapBlock(swap)
	expectedError := "illegal base64 data at input byte 4"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// different key signature
	swap, _, _, validator, err = swapOfferSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	err = swap.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}
	err = validator.ValidateSwapBlock(swap)
	expectedError = "crypto/rsa: verification error"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// is there a previous? Hint: there shouldn't be
	swap, swap2, _, validator, err = swapOfferSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	swap.Previous = swap2.Hash()
	err = swap.SignBlock(key3)
	if err != nil {
		t.Fatal(err)
	}
	err = validator.ValidateSwapBlock(swap)
	expectedError = "prev and right must be null together"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// make sure there's an account block from the left field
	swap, _, _, validator, err = swapOfferSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	// no block
	swap.Left = "not a block"
	err = swap.SignBlock(key3)
	if err != nil {
		t.Fatal(err)
	}
	err = validator.ValidateSwapBlock(swap)
	expectedError = "link field references invalid block"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	//not a send
	tempI := tradeblocks.NewIssueBlock("address", 100.0)
	swap.Left = tempI.Hash()
	err = swap.SignBlock(key3)
	if err != nil {
		t.Fatal(err)
	}
	err = validator.ValidateSwapBlock(swap)
	expectedError = "link field references invalid block"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}
}

// only for commit, each action has different test
func TestSwapCommitValidation(t *testing.T) {
	key, address := CreateAccount(t)
	key2, address2 := CreateAccount(t)
	key3, address3 := CreateAccount(t)
	key4, address4 := CreateAccount(t)

	keyList := []*rsa.PrivateKey{key, key2, key3}
	addressList := []string{address, address2, address3}

	swap, swap2, send, validator, err := swapOfferSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	// base success
	err = validator.ValidateSwapBlock(swap2)
	if err != nil {
		t.Fatal(err)
	}

	// signature verifying tests
	// Note: can't do executors on offers
	_, swap2, send, validator, err = swapOfferSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	// random signature
	swap2.Signature = "garbage"
	err = validator.ValidateSwapBlock(swap2)
	expectedError := "illegal base64 data at input byte 4"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	swap2.Previous = send.Hash()
	err = swap2.SignBlock(key2)
	if err != nil {
		t.Fatal(err)
	}
	err = validator.ValidateSwapBlock(swap2)
	expectedError = "previous must be not null"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// bad account decoding
	_, swap2, send, validator, err = swapOfferSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	swap2.Counterparty = "help"
	err = swap2.SignBlock(key2)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateSwapBlock(swap2)
	if err == nil || err != ErrInvalidAddress {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, ErrInvalidAddress)
	}

	// bad executor
	swap2.Executor = "help"
	err = swap2.SignBlock(key2)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateSwapBlock(swap2)
	if err == nil || err != ErrInvalidAddress {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, ErrInvalidAddress)
	}

	// good executor
	t.Skip("TODO swap must be modified before inserting into blockstore")
	swap, swap2, send, validator, err = swapOfferSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	swap.Executor = address4
	err = swap.SignBlock(key3)
	if err != nil {
		t.Fatal(err)
	}

	swap2.Executor = address4
	err = swap2.SignBlock(key4)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateSwapBlock(swap2)
	if err != nil {
		t.Fatalf("Expected success on signing with executor, got %v", err)
	}

	// swaps not same fields
	_, swap2, _, validator, err = swapOfferSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}
	// long list of fields that could trigger this:
	// Account, Token, ID, Left, RefundLeft, RefundRight,
	// Counterparty, Want, Quantity, Executor, Fee
	// Note on account: used as address to get publicKey, changing will cause
	// a pem parsing error
	swap2.ID = "help"
	err = swap2.SignBlock(key2)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateSwapBlock(swap2)
	expectedError = "Counterparty swap has incorrect fields: must match originating swap"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// original send doesn't exist (linked to bad swap)
	swap, swap2, _, validator, err = swapOfferSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	swap.Left = badAddress
	err = swap.SignBlock(key3)
	if err != nil {
		t.Fatal(err)
	}

	swap2.Left = badAddress
	err = swap2.SignBlock(key2)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateSwapBlock(swap2)
	expectedError = "originating send not found"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// counter send doesn't exist
	swap, swap2, _, validator, err = swapOfferSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	swap2.Right = badAddress
	err = swap2.SignBlock(key2)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateSwapBlock(swap2)
	expectedError = "counter send not found"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// counter send doesn't exist
	swap, swap2, _, validator, err = swapOfferSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	send2, err := validator.blockStore.GetAccountBlock(swap2.Right)
	if err != nil {
		t.Fatal(err)
	}
	if send2 == nil {
		t.Fatal("Block not found")
	}

	send2.Previous = badAddress
	err = send2.SignBlock(key2)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateSwapBlock(swap2)
	expectedError = "counter send prev not found"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// counter send doesn't exist
	swap, swap2, _, validator, err = swapOfferSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	send2, err = validator.blockStore.GetAccountBlock(swap2.Right)
	if err != nil {
		t.Fatal(err)
	}
	if send2 == nil {
		t.Fatal("Block not found")
	}
	send2.Balance = 0
	err = send2.SignBlock(key2)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateSwapBlock(swap2)
	expectedError = "amount/token requested not sent"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}
}

func swapRefundLeftSetup(key []*rsa.PrivateKey, address []string, t *testing.T) (*tradeblocks.SwapBlock, *tradeblocks.SwapBlock, *tradeblocks.AccountBlock, *SwapBlockValidator, error) {
	blockStore := NewBlockStore()

	i := tradeblocks.NewIssueBlock(address[0], 100.0)
	send := tradeblocks.NewSendBlock(i, address[2]+":swap:test-ID", 50.0)

	i2 := tradeblocks.NewIssueBlock(address[1], 50.0)
	swap := tradeblocks.NewOfferBlock(address[2], send, "test-ID", address[1], address[1], 10.0, "", 0.0)

	refundLeft := tradeblocks.NewRefundLeftBlock(swap, address[0])

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

	err = swap.SignBlock(key[2])
	if err != nil {
		t.Fatal(err)
	}

	err = refundLeft.SignBlock(key[2])
	if err != nil {
		t.Fatal(err)
	}

	if err := blockStore.AddAccountBlock(i); err != nil {
		t.Fatal(err)
	}
	if err := blockStore.AddAccountBlock(send); err != nil {
		t.Fatal(err)
	}
	if err := blockStore.AddAccountBlock(i2); err != nil {
		t.Fatal(err)
	}
	if err := blockStore.AddSwapBlock(swap); err != nil {
		t.Fatal(err)
	}
	if err := blockStore.AddSwapBlock(refundLeft); err != nil {
		t.Fatal(err)
	}

	validator := NewSwapValidator(blockStore)

	return swap, refundLeft, send, validator, nil
}

// only for refund-left, each action has different test
func TestSwapRefundLeftValidation(t *testing.T) {
	key, address := CreateAccount(t)
	key2, address2 := CreateAccount(t)
	key3, address3 := CreateAccount(t)

	keyList := []*rsa.PrivateKey{key, key2, key3}
	addressList := []string{address, address2, address3}

	swap, refundLeft, _, validator, err := swapRefundLeftSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	// base success
	err = validator.ValidateSwapBlock(refundLeft)
	if err != nil {
		t.Fatal(err)
	}

	// signature verifying tests
	// Note: can't do executors on offers
	_, refundLeft, send, validator, err := swapRefundLeftSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	// random signature
	refundLeft.Signature = "garbage"
	err = validator.ValidateSwapBlock(refundLeft)
	expectedError := "illegal base64 data at input byte 4"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// invalid previous
	refundLeft.Previous = send.Hash()
	err = refundLeft.SignBlock(key3)
	if err != nil {
		t.Fatal(err)
	}
	err = validator.ValidateSwapBlock(refundLeft)
	expectedError = "previous must be not null"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// long list of fields that could trigger this:
	// Account, Token, ID, Left, Right, RefundRight,
	// Counterparty, Want, Quantity, Executor, Fee
	// Note on account: used as address to get publicKey, changing will cause
	// a pem parsing error
	_, refundLeft, send, validator, err = swapRefundLeftSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	refundLeft.Fee = 1.0
	err = refundLeft.SignBlock(key3)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateSwapBlock(refundLeft)
	expectedError = "Counterparty swap has incorrect fields: must match originating swap"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// originating send invalid
	t.Skip("TODO swap must be modified before inserting into blockstore")
	swap, refundLeft, send, validator, err = swapRefundLeftSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	swap.Left = badAddress
	err = swap.SignBlock(key3)
	if err != nil {
		t.Fatal(err)
	}

	refundLeft.Left = badAddress
	err = refundLeft.SignBlock(key3)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateSwapBlock(refundLeft)
	expectedError = "Originating send is invalid or not found"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// refundleft is not the initiators account
	_, refundLeft, send, validator, err = swapRefundLeftSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	refundLeft.RefundLeft = badAddress
	err = refundLeft.SignBlock(key3)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateSwapBlock(refundLeft)
	expectedError = "Refund must be to initiator's account"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}
}

func swapRefundRightSetup(key []*rsa.PrivateKey, address []string, t *testing.T) (*tradeblocks.SwapBlock, *tradeblocks.SwapBlock, *tradeblocks.SwapBlock, *tradeblocks.AccountBlock, *SwapBlockValidator, error) {
	blockStore := NewBlockStore()

	i := tradeblocks.NewIssueBlock(address[0], 100.0)
	send := tradeblocks.NewSendBlock(i, address[2]+":swap:test-ID", 50.0)
	i2 := tradeblocks.NewIssueBlock(address[1], 50)
	send2 := tradeblocks.NewSendBlock(i2, address[2]+":swap:test-ID", 10.0)
	swap := tradeblocks.NewOfferBlock(address[2], send, "test-ID", address[1], address[1], 10.0, "", 0.0)
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

	err = refundRight.SignBlock(key[1])
	if err != nil {
		t.Fatal(err)
	}

	if err := blockStore.AddAccountBlock(i); err != nil {
		t.Fatal(err)
	}
	if err := blockStore.AddAccountBlock(send); err != nil {
		t.Fatal(err)
	}
	if err := blockStore.AddAccountBlock(i2); err != nil {
		t.Fatal(err)
	}
	if err := blockStore.AddAccountBlock(send2); err != nil {
		t.Fatal(err)
	}

	if err := blockStore.AddSwapBlock(swap); err != nil {
		t.Fatal(err)
	}
	if err := blockStore.AddSwapBlock(refundLeft); err != nil {
		t.Fatal(err)
	}
	if err := blockStore.AddSwapBlock(refundRight); err != nil {
		t.Fatal(err)
	}

	validator := NewSwapValidator(blockStore)

	return swap, refundLeft, refundRight, send2, validator, nil
}

// only for refund-left, each action has different test
func TestSwapRefundRightValidation(t *testing.T) {
	key, address := CreateAccount(t)
	key2, address2 := CreateAccount(t)
	key3, address3 := CreateAccount(t)
	key4, address4 := CreateAccount(t)

	keyList := []*rsa.PrivateKey{key, key2, key3}
	addressList := []string{address, address2, address3}

	swap, refundLeft, refundRight, send2, validator, err := swapRefundRightSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	// base success
	err = validator.ValidateSwapBlock(refundRight)
	if err != nil {
		t.Fatal(err)
	}

	// signature verifying tests
	// Note: can't do executors on offers
	_, _, refundRight, _, validator, err = swapRefundRightSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	// random signature
	refundRight.Signature = "garbage"
	err = validator.ValidateSwapBlock(refundRight)
	expectedError := "illegal base64 data at input byte 4"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// bad executor
	_, refundLeft, refundRight, _, validator, err = swapRefundRightSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	refundLeft.Executor = "help"
	err = refundLeft.SignBlock(key3)
	if err != nil {
		t.Fatal(err)
	}

	refundRight.Executor = "help"
	err = refundRight.SignBlock(key3)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateSwapBlock(refundRight)
	if err == nil || err != ErrInvalidAddress {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// good executor
	t.Skip("TODO blocks must be modified before inserting into block store")
	_, refundLeft, refundRight, _, validator, err = swapRefundRightSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	refundLeft.Executor = address4
	err = refundLeft.SignBlock(key3)
	if err != nil {
		t.Fatal(err)
	}

	refundRight.Executor = address4
	err = refundRight.SignBlock(key4)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateSwapBlock(refundRight)
	if err != nil {
		t.Fatalf("Expected success on signing with executor, got %v", err)
	}

	// invalid previous
	_, refundLeft, refundRight, send2, validator, err = swapRefundRightSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	refundRight.Previous = send2.Hash()
	err = refundRight.SignBlock(key2)
	if err != nil {
		t.Fatal(err)
	}
	err = validator.ValidateSwapBlock(refundRight)
	expectedError = "previous must be not null"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// previous must be refund-left
	_, refundLeft, refundRight, _, validator, err = swapRefundRightSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	refundRight.Previous = swap.Hash()
	err = refundRight.SignBlock(key2)
	if err != nil {
		t.Fatal(err)
	}
	err = validator.ValidateSwapBlock(refundRight)
	expectedError = "Previous must be a refund-left"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// swap alignment
	// long list of fields that could trigger this:
	// Account, Token, ID, Left, Right, RefundLeft,
	// Counterparty, Want, Quantity, Executor, Fee
	// Note on account: used as address to get publicKey, changing will cause
	// a pem parsing error
	_, refundLeft, refundRight, send2, validator, err = swapRefundRightSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	refundRight.Fee = 1.0
	err = refundRight.SignBlock(key2)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateSwapBlock(refundRight)
	expectedError = "Refund Right must match Refund Left fields"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// invalid counterparty send
	_, refundLeft, refundRight, send2, validator, err = swapRefundRightSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	refundRight.Right = badAddress
	err = refundRight.SignBlock(key2)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateSwapBlock(refundRight)
	expectedError = "counterparty send not found/invalid"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// wrong address to refund
	// changing block.Account will break the decrypter, make the send Account wrong
	_, refundLeft, refundRight, send2, validator, err = swapRefundRightSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	refundRight.RefundRight = badAddress
	err = refundRight.SignBlock(key2)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateSwapBlock(refundRight)
	expectedError = "Account for refund must be same as original send account"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}
}

func createOrderSetup(key []*rsa.PrivateKey, address []string, t *testing.T) (*tradeblocks.AccountBlock, *tradeblocks.OrderBlock, *OrderBlockValidator, error) {
	blockStore := NewBlockStore()

	i := tradeblocks.NewIssueBlock(address[0], 100.0)
	send := tradeblocks.NewSendBlock(i, address[0]+":order:ID0", 50.0)
	order := tradeblocks.NewCreateOrderBlock(address[0], send, 50, "ID0", false, "quote0", 10.0, "", 0.0)

	err := i.SignBlock(key[0])
	if err != nil {
		t.Fatal(err)
	}

	err = send.SignBlock(key[0])
	if err != nil {
		t.Fatal(err)
	}

	err = order.SignBlock(key[0])
	if err != nil {
		t.Fatal(err)
	}

	if err := blockStore.AddAccountBlock(i); err != nil {
		t.Fatal(err)
	}
	if err := blockStore.AddAccountBlock(send); err != nil {
		t.Fatal(err)
	}
	if err := blockStore.AddOrderBlock(order); err != nil {
		t.Fatal(err)
	}

	validator := NewOrderValidator(blockStore)

	return send, order, validator, nil
}

func TestCreateOrderValidation(t *testing.T) {
	key, address := CreateAccount(t)
	key2, address2 := CreateAccount(t)
	key3, address3 := CreateAccount(t)

	keyList := []*rsa.PrivateKey{key, key2, key3}
	addressList := []string{address, address2, address3}

	send, order, validator, err := createOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	// base success
	err = validator.ValidateOrderBlock(order)
	if err != nil {
		t.Fatal(err)
	}

	// bad signature format
	_, order, validator, err = createOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	order.Signature = "garbage"
	err = validator.ValidateOrderBlock(order)
	expectedError := "illegal base64 data at input byte 4"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// signed by wrong key
	_, order, validator, err = createOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	order.Signature = badSignature
	err = validator.ValidateOrderBlock(order)
	expectedError = "crypto/rsa: verification error"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// send is invalid
	_, order, validator, err = createOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	order.Link = badAddress
	err = order.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(order)
	expectedError = "Order linked send not found"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	t.Skip("TODO Send must be modified before inserting into block store")
	// send is not to the order
	send, order, validator, err = createOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	send.Link = badAddress
	err = send.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(order)
	expectedError = "Linked send block does not send to this order"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// send previous not found
	send, order, validator, err = createOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	send.Previous = badAddress
	err = send.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(order)
	expectedError = "Linked send block does not have a valid previous"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// balances don't line up
	send, order, validator, err = createOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	send.Balance = 75
	err = send.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(order)
	expectedError = "Balance sent and Balance created do not match up"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// balances don't line up v2
	send, order, validator, err = createOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	order.Balance = 75
	err = order.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(order)
	expectedError = "Balance sent and Balance created do not match up"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// try to executor sign
	send, order, validator, err = createOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	order.Executor = address2
	err = order.SignBlock(key2)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(order)
	expectedError = "crypto/rsa: verification error"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}
}

func acceptOrderSetup(key []*rsa.PrivateKey, address []string, t *testing.T) (*tradeblocks.AccountBlock, *tradeblocks.OrderBlock, *tradeblocks.OrderBlock, *tradeblocks.SwapBlock, *OrderBlockValidator, error) {
	blockStore := NewBlockStore()

	i := tradeblocks.NewIssueBlock(address[0], 100.0)
	send := tradeblocks.NewSendBlock(i, address[0]+":order:test-ID", 50.0)
	order := tradeblocks.NewCreateOrderBlock(address[0], send, 50, "test-ID", false, address[1], 10.0, "", 0.0)

	i2 := tradeblocks.NewIssueBlock(address[1], 500.0)
	send2 := tradeblocks.NewSendBlock(i2, address[2]+":swap:test-ID", 500)
	swap := tradeblocks.NewOfferBlock(address[2], send2, "test-ID", address[0], address[0], 50, "", 0.0)
	order2 := tradeblocks.NewAcceptOrderBlock(order, address[2]+":swap:test-ID", 0)

	err := i.SignBlock(key[0])
	if err != nil {
		t.Fatal(err)
	}

	err = send.SignBlock(key[0])
	if err != nil {
		t.Fatal(err)
	}

	err = order.SignBlock(key[0])
	if err != nil {
		t.Fatal(err)
	}

	err = i2.SignBlock(key[1])
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

	err = order2.SignBlock(key[0])
	if err != nil {
		t.Fatal(err)
	}

	if err := blockStore.AddAccountBlock(i); err != nil {
		t.Fatal(err)
	}
	if err := blockStore.AddAccountBlock(i2); err != nil {
		t.Fatal(err)
	}
	if err := blockStore.AddAccountBlock(send); err != nil {
		t.Fatal(err)
	}
	if err := blockStore.AddAccountBlock(send2); err != nil {
		t.Fatal(err)
	}
	if err := blockStore.AddSwapBlock(swap); err != nil {
		t.Fatal(err)
	}
	if err := blockStore.AddOrderBlock(order); err != nil {
		t.Fatal(err)
	}
	if err := blockStore.AddOrderBlock(order2); err != nil {
		t.Fatal(err)
	}
	validator := NewOrderValidator(blockStore)

	return send2, order, order2, swap, validator, nil
}

func TestAcceptOrderValidation(t *testing.T) {
	key, address := CreateAccount(t)
	key2, address2 := CreateAccount(t)
	key3, address3 := CreateAccount(t)
	key4, address4 := CreateAccount(t)

	keyList := []*rsa.PrivateKey{key, key2, key3, key4}
	addressList := []string{address, address2, address3, address4}

	_, createOrder, acceptOrder, _, validator, err := acceptOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	// base success
	err = validator.ValidateOrderBlock(acceptOrder)
	if err != nil {
		t.Fatal(err)
	}

	// bad signature format
	_, _, acceptOrder, _, validator, err = acceptOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	acceptOrder.Signature = "garbage"
	err = validator.ValidateOrderBlock(acceptOrder)
	expectedError := "illegal base64 data at input byte 4"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// signed by wrong key
	_, _, acceptOrder, _, validator, err = acceptOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	acceptOrder.Signature = badSignature
	err = validator.ValidateOrderBlock(acceptOrder)
	expectedError = "crypto/rsa: verification error"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// signed by executor
	t.Skip("TODO blocks must be modified before inserting into block store")
	_, createOrder, acceptOrder, _, validator, err = acceptOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	createOrder.Executor = address4
	err = createOrder.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}
	acceptOrder.Executor = address4
	err = acceptOrder.SignBlock(key4)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(acceptOrder)
	if err != nil {
		t.Fatal(err)
	}

	// signed by executor gone wrong
	_, createOrder, acceptOrder, _, validator, err = acceptOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	createOrder.Executor = address4
	err = createOrder.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}
	acceptOrder.Executor = address4
	err = acceptOrder.SignBlock(key3)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(acceptOrder)
	expectedError = "crypto/rsa: verification error"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// previous invalid
	_, _, acceptOrder, _, validator, err = acceptOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	acceptOrder.Previous = badAddress
	err = acceptOrder.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(acceptOrder)
	expectedError = "Previous block invalid"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// fields don't match
	_, _, acceptOrder, _, validator, err = acceptOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	acceptOrder.Fee = -1
	err = acceptOrder.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(acceptOrder)
	expectedError = "Fields did not line up with order creation"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// fields don't match v2
	_, _, acceptOrder, _, validator, err = acceptOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	acceptOrder.Partial = true
	err = acceptOrder.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(acceptOrder)
	expectedError = "Fields did not line up with order creation"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// fields don't match v2
	_, _, acceptOrder, _, validator, err = acceptOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	acceptOrder.ID = badAddress
	err = acceptOrder.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(acceptOrder)
	expectedError = "Fields did not line up with order creation"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// linked swap invalid
	_, _, acceptOrder, _, validator, err = acceptOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	acceptOrder.Link = badAddress + ":swap:test-bad"
	err = acceptOrder.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(acceptOrder)
	expectedError = "Swap link validation failed"
	if err == nil || !strings.Contains(err.Error(), expectedError) {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// linked swap doesn't point to offer
	_, createOrder, acceptOrder, _, validator, err = acceptOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	createOrder.Account = address4
	err = createOrder.SignBlock(key4)
	if err != nil {
		t.Fatal(err)
	}

	acceptOrder.Account = address4
	err = acceptOrder.SignBlock(key4)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(acceptOrder)
	expectedError = "The swap must have counterparty point to this order"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// linked swap bad ID
	_, createOrder, acceptOrder, _, validator, err = acceptOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	createOrder.ID = "bad ID"
	if err = createOrder.SignBlock(key); err != nil {
		t.Fatal(err)
	}

	acceptOrder.ID = "bad ID"
	if err = acceptOrder.SignBlock(key); err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(acceptOrder)
	expectedError = "The swap must have same ID as the order"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// mismatched token type v1
	_, createOrder, acceptOrder, _, validator, err = acceptOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	createOrder.Token = "bad Token"
	if err = createOrder.SignBlock(key); err != nil {
		t.Fatal(err)
	}

	acceptOrder.Token = "bad Token"
	if err = acceptOrder.SignBlock(key); err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(acceptOrder)
	expectedError = "Swap and Order token mismatch"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// balances don't line up
	_, _, acceptOrder, _, validator, err = acceptOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	acceptOrder.Balance = -1
	err = acceptOrder.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(acceptOrder)
	expectedError = "Invalid block balance, must be greater than zero"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// Non accepting partial given partial
	_, _, acceptOrder, _, validator, err = acceptOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	acceptOrder.Balance = 5
	err = acceptOrder.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(acceptOrder)
	expectedError = "Balance must be paid in full for blocks with Partial = false"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// order not given full amount
	_, createOrder, acceptOrder, _, validator, err = acceptOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	createOrder.Partial = true
	err = createOrder.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	acceptOrder.Partial = true
	acceptOrder.Balance = 5
	err = acceptOrder.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(acceptOrder)
	expectedError = "Balance sent to order is invalid"
	if err == nil || !strings.Contains(err.Error(), expectedError) {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}
}

func refundOrderSetup(key []*rsa.PrivateKey, address []string, t *testing.T) (*tradeblocks.OrderBlock, *tradeblocks.OrderBlock, *OrderBlockValidator, error) {
	blockStore := NewBlockStore()

	i := tradeblocks.NewIssueBlock(address[0], 100.0)
	send := tradeblocks.NewSendBlock(i, address[0]+":order:test-ID", 50.0)
	order := tradeblocks.NewCreateOrderBlock(address[0], send, 50, "test-ID", false, address[1], 10.0, "", 0.0)
	refund := tradeblocks.NewRefundOrderBlock(order, address[0])

	err := i.SignBlock(key[0])
	if err != nil {
		t.Fatal(err)
	}

	err = send.SignBlock(key[0])
	if err != nil {
		t.Fatal(err)
	}

	err = order.SignBlock(key[0])
	if err != nil {
		t.Fatal(err)
	}

	err = refund.SignBlock(key[0])
	if err != nil {
		t.Fatal(err)
	}

	if err := blockStore.AddAccountBlock(i); err != nil {
		t.Fatal(err)
	}

	if err := blockStore.AddAccountBlock(send); err != nil {
		t.Fatal(err)
	}

	if err := blockStore.AddOrderBlock(order); err != nil {
		t.Fatal(err)
	}

	if err := blockStore.AddOrderBlock(refund); err != nil {
		t.Fatal(err)
	}

	validator := NewOrderValidator(blockStore)

	return order, refund, validator, nil
}

func TestRefundOrderValidation(t *testing.T) {
	key, address := CreateAccount(t)
	key2, address2 := CreateAccount(t)
	key3, address3 := CreateAccount(t)
	key4, address4 := CreateAccount(t)

	keyList := []*rsa.PrivateKey{key, key2, key3, key4}
	addressList := []string{address, address2, address3, address4}

	_, refund, validator, err := refundOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	// base success
	err = validator.ValidateOrderBlock(refund)
	if err != nil {
		t.Fatal(err)
	}

	// bad signature format
	_, refund, validator, err = refundOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	refund.Signature = "garbage"
	err = validator.ValidateOrderBlock(refund)
	expectedError := "illegal base64 data at input byte 4"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// signed by wrong key
	_, refund, validator, err = refundOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	refund.Signature = badSignature
	err = validator.ValidateOrderBlock(refund)
	expectedError = "crypto/rsa: verification error"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// previous invalid
	_, refund, validator, err = refundOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	refund.Previous = badAddress
	err = refund.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(refund)
	expectedError = "Previous block invalid"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// fields mismatch v1
	_, refund, validator, err = refundOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	refund.Account = badAddress
	err = refund.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(refund)
	if err == nil || err != ErrInvalidAddress {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// fields mismatch v2
	_, refund, validator, err = refundOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	refund.Token = badAddress
	err = refund.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(refund)
	expectedError = "Fields did not line up with head order block"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// fields mismatch v3
	_, refund, validator, err = refundOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	refund.Balance = 13
	err = refund.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(refund)
	expectedError = "Fields did not line up with head order block"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// fields mismatch v4
	_, refund, validator, err = refundOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	refund.ID = "bad-ID"
	err = refund.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(refund)
	expectedError = "Fields did not line up with head order block"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// fields mismatch v5
	_, refund, validator, err = refundOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	refund.Quote = badAddress
	err = refund.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(refund)
	expectedError = "Fields did not line up with head order block"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// fields mismatch v6
	_, refund, validator, err = refundOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	refund.Price = 0.0
	err = refund.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(refund)
	expectedError = "Fields did not line up with head order block"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// fields mismatch v7
	_, refund, validator, err = refundOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	refund.Partial = true
	err = refund.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(refund)
	expectedError = "Fields did not line up with head order block"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// fields mismatch v8
	_, refund, validator, err = refundOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	refund.Executor = badAddress
	err = refund.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(refund)
	expectedError = "Fields did not line up with head order block"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}

	// refund to wrong account
	_, refund, validator, err = refundOrderSetup(keyList, addressList, t)
	if err != nil {
		t.Fatal(err)
	}

	refund.Link = badAddress
	err = refund.SignBlock(key)
	if err != nil {
		t.Fatal(err)
	}

	err = validator.ValidateOrderBlock(refund)
	expectedError = "Must refund to the original sender"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("error \"%v\" did not match \"%s\" ", err, expectedError)
	}
}
