package app

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"testing"

	tb "github.com/jephir/tradeblocks"
)

func TestBlockStore(t *testing.T) {
	key, address, err := GetAddress()
	if err != nil {
		t.Fatal(err)
	}
	s := NewBlockStore()
	b, err := tb.SignedAccountBlock(tb.NewIssueBlock(address, 100), key)
	if err != nil {
		t.Fatal(err)
	}
	expect := `{"Action":"issue","Account":"` + address + `","Token":"` + address + `","Previous":"","Representative":"","Balance":100,"Link":"","Signature":"` + b.Signature + `"}`

	if err := s.AddAccountBlock(b); err != nil {
		t.Fatal(err)
	}
	res := s.GetAccountBlock(b.Hash())
	ss, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	got := string(ss)
	if got != expect {
		t.Fatalf("Issue was incorrect, got: %s, want: %s", got, expect)
	}
}

func TestDoubleSpend(t *testing.T) {
	t.Skip()
	key, address, err := GetAddress()
	if err != nil {
		t.Fatal(err)
	}
	s := NewBlockStore()

	b, err := tb.SignedAccountBlock(tb.NewIssueBlock(address, 100), key)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.AddAccountBlock(b); err != nil {
		t.Fatal(err)
	}

	b1, err := tb.SignedAccountBlock(tb.NewSendBlock(b, address, 10), key)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.AddAccountBlock(b1); err != nil {
		t.Fatal(err)
	}
	b2, err := tb.SignedAccountBlock(tb.NewSendBlock(b, address, 10), key)
	if err != nil {
		t.Fatal(err)
	}

	err = s.AddAccountBlock(b2)
	if err == nil {
		t.Fatal("Expected error, got none")
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
