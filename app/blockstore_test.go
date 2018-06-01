package app

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"testing"

	"github.com/jephir/tradeblocks"
)

func TestBlockStore(t *testing.T) {
	key, address, err := GetAddress()
	if err != nil {
		t.Error(err)
	}
	s := NewBlockStore()
	b := tradeblocks.NewIssueBlock(address, 100)
	errSign := b.SignBlock(key)
	if errSign != nil {
		t.Error(errSign)
	}
	expect := `{"Action":"issue","Account":"` + address + `","Token":"` + address + `","Previous":"","Representative":"","Balance":100,"Link":"","Signature":"` + b.Signature + `"}`

	if _, err := s.AddBlock(b); err != nil {
		t.Error(err)
	}
	res, _ := s.GetBlock(b.Hash())
	ss, err := json.Marshal(res)
	if err != nil {
		t.Error(err)
	}
	got := string(ss)
	if got != expect {
		t.Errorf("Issue was incorrect, got: %s, want: %s", got, expect)
	}
}

func TestDoubleSpend(t *testing.T) {
	key, address, err := GetAddress()
	if err != nil {
		t.Error(err)
	}
	s := NewBlockStore()
	b := tradeblocks.NewIssueBlock(address, 100)
	errSign := b.SignBlock(key)
	if errSign != nil {
		t.Error(errSign)
	}

	if _, err := s.AddBlock(b); err != nil {
		t.Error(err)
	}
	b1 := tradeblocks.NewSendBlock(b, address, 10)
	errSign = b1.SignBlock(key)
	if errSign != nil {
		t.Error(errSign)
	}

	if _, err := s.AddBlock(b1); err != nil {
		t.Error(err)
	}
	b2 := tradeblocks.NewSendBlock(b, address, 10)
	errSign = b2.SignBlock(key)
	if errSign != nil {
		t.Error(errSign)
	}

	_, err = s.AddBlock(b2)
	if _, ok := err.(*blockConflictError); !ok {
		t.Errorf("expected block conflict error, got '%s'", err.Error())
	}
}

func GetAddress() (*rsa.PrivateKey, string, error) {
	var key, err = rsa.GenerateKey(rand.Reader, 512)
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, "", err
	}
	p := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	pubKeyReader := bytes.NewReader(p)
	address, err := PublicKeyToAddress(pubKeyReader)
	if err != nil {
		return nil, "", err
	}
	return key, address, nil
}
