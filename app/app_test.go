package app

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/jephir/tradeblocks"
)

var publicKey = strings.NewReader(`
-----BEGIN RSA PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDVleYQ+MOGhHVvkmzCkJrjI5CL
4NMHwNRl7SRnElFI2+nWjYMEwSOlp5pTcHBzjRhJOx1SbLtiKRKFg1Q9wUevNeWS
PMjB1l+LWmUTRqNTcAPQc0Vdeumjqs1P+eHERfk9MwqNsrPytvGwvNQJ05PkgLSk
Xu58kr5iXxMABIukbQIDAQAB
-----END RSA PUBLIC KEY-----
`)

var accountAddress = "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDVleYQ-MOGhHVvkmzCkJrjI5CL4NMHwNRl7SRnElFI2-nWjYMEwSOlp5pTcHBzjRhJOx1SbLtiKRKFg1Q9wUevNeWSPMjB1l-LWmUTRqNTcAPQc0Vdeumjqs1P-eHERfk9MwqNsrPytvGwvNQJ05PkgLSkXu58kr5iXxMABIukbQIDAQAB"

func TestRegister(t *testing.T) {
	if _, err := Register(ioutil.Discard, ioutil.Discard, "testuser", 1024); err != nil {
		t.Fatal(err)
	}
}

func TestIssue(t *testing.T) {
	expect := `{"Action":"issue","Account":"xtb:` + accountAddress + `","Token":"xtb:` + accountAddress + `","Previous":"","Representative":"xtb:` + accountAddress + `","Balance":100,"Link":"","Signature":""}`
	publicKey.Seek(0, io.SeekStart)
	issue, err := Issue(publicKey, 100)
	if err != nil {
		t.Fatal(err)
	}
	s, err := json.Marshal(issue)
	if err != nil {
		t.Fatal(err)
	}
	got := string(s)
	if got != expect {
		t.Fatalf("Issue was incorrect, got: %s,\nwant: %s", got, expect)
	}
}

func TestSend(t *testing.T) {
	previousText := "LTWNQGFK7UJSZE7HZKFIJKORRUJ3FTWSTGQGAAKVO3VB6NLHT7XA"
	expect := `{"Action":"send","Account":"xtb:` + accountAddress + `","Token":"xtb:` + accountAddress + `","Previous":"` + previousText + `","Representative":"xtb:` + accountAddress + `","Balance":50,"Link":"xtb:testreceiver","Signature":""}`
	publicKey.Seek(0, io.SeekStart)
	address, err := PublicKeyToAddress(publicKey)
	if err != nil {
		t.Fatal(err)
	}
	previous := tradeblocks.NewIssueBlock(address, 100)
	publicKey.Seek(0, io.SeekStart)
	send, err := Send(publicKey, previous, "xtb:testreceiver", 50.0)
	if err != nil {
		t.Fatal(err)
	}
	s, err := json.Marshal(send)
	if err != nil {
		t.Fatal(err)
	}
	got := string(s)
	if got != expect {
		t.Fatalf("Issue was incorrect, got: %s,\nwant: %s", got, expect)
	}
}

func TestOpenFromSend(t *testing.T) {
	linkText := "R4UHAP3NDCEAV7WON3L7NAGGTFYYSPVZADWCMSD5O4VD7IFE534Q"
	expect := `{"Action":"open","Account":"xtb:` + accountAddress + `","Token":"xtb:sender","Previous":"","Representative":"xtb:` + accountAddress + `","Balance":50,"Link":"` + linkText + `","Signature":""}`
	publicKey.Seek(0, io.SeekStart)
	address, err := PublicKeyToAddress(publicKey)
	if err != nil {
		t.Fatal(err)
	}
	issue := tradeblocks.NewIssueBlock("xtb:sender", 50.0)
	send := tradeblocks.NewSendBlock(issue, address, 50.0)
	publicKey.Seek(0, io.SeekStart)
	open, err := OpenFromSend(publicKey, send, 50.0)
	if err != nil {
		t.Fatal(err)
	}
	s, err := json.Marshal(open)
	if err != nil {
		t.Fatal(err)
	}
	got := string(s)
	if got != expect {
		t.Fatalf("Issue was incorrect, got: %s,\nwant: %s", got, expect)
	}
}

func TestOpenFromSwap(t *testing.T) {
	_, address := CreateAccount(t)
	key2, address2 := CreateAccount(t)

	i := tradeblocks.NewIssueBlock(address, 100.0)
	send := tradeblocks.NewSendBlock(i, address2+":swap:test-ID", 50.0)

	i2 := tradeblocks.NewIssueBlock(address, 50.0)
	send2 := tradeblocks.NewSendBlock(i2, address2+":swap:test-ID", 10.0)

	swap := tradeblocks.NewOfferBlock(address2, send, "test-ID", address, address, 10.0, "", 0.0)
	swap2 := tradeblocks.NewCommitBlock(swap, send2)

	b, err := x509.MarshalPKIXPublicKey(&key2.PublicKey)
	if err != nil {
		return
	}
	publicKey := bytes.NewReader(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: b,
	}))

	open, err := OpenFromSwap(publicKey, address, swap2, 50.0)

	if err != nil {
		t.Fatal(err)
	}
	s, err := json.Marshal(open)
	if err != nil {
		t.Fatal(err)
	}
	got := string(s)
	expect := `{"Action":"open","Account":"` + address2 + `","Token":"` + address + `","Previous":"","Representative":"` + address2 + `","Balance":50,"Link":"` + swap2.Hash() + `","Signature":"` + open.Signature + `"}`

	if got != expect {
		t.Fatalf("Issue was incorrect, got: %s,\nwant: %s", got, expect)
	}
}

func TestReceive(t *testing.T) {
	previousText := "A56LDHHQFZJP4XYZXMBIEEULNQX72DFT6OF5FATZYBTEZU6EEXYA"
	linkText := "R4UHAP3NDCEAV7WON3L7NAGGTFYYSPVZADWCMSD5O4VD7IFE534Q"
	expect := `{"Action":"receive","Account":"xtb:` + accountAddress + `","Token":"xtb:sender","Previous":"` + previousText + `","Representative":"xtb:` + accountAddress + `","Balance":75,"Link":"` + linkText + `","Signature":""}`
	publicKey.Seek(0, io.SeekStart)
	address, err := PublicKeyToAddress(publicKey)
	if err != nil {
		t.Fatal(err)
	}
	issue := tradeblocks.NewIssueBlock("xtb:sender", 50.0)
	send := tradeblocks.NewSendBlock(issue, address, 50.0)
	send2 := tradeblocks.NewSendBlock(send, address, 25.0)
	previous := tradeblocks.NewOpenBlockFromSend(address, send2, 25.0)
	publicKey.Seek(0, io.SeekStart)
	receive, err := Receive(publicKey, previous, send, 50.0)
	if err != nil {
		t.Fatal(err)
	}
	s, err := json.Marshal(receive)
	if err != nil {
		t.Fatal(err)
	}
	got := string(s)
	if got != expect {
		t.Fatalf("Issue was incorrect, got: %s,\nwant: %s", got, expect)
	}
}

func TestAddress(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}

	addr, err := PrivateKeyToAddress(priv)
	if err != nil {
		t.Fatal(err)
	}

	b, err := AddressToPublicKey(addr)
	if err != nil {
		t.Fatal(err)
	}

	block, _ := pem.Decode(b)
	if block == nil {
		t.Fatal("no block")
	}

	var pub *rsa.PublicKey
	p, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	pub, ok := p.(*rsa.PublicKey)
	if !ok {
		t.Fatal("key not an RSA key")
	}

	expect := priv.PublicKey

	if pub.E != expect.E {
		t.Fatalf("E doesn't match; expected %d, got %d", expect.E, pub.E)
	}

	if pub.N.Cmp(expect.N) != 0 {
		t.Fatalf("N doesn't match; expected %d, got %d (diff %d)", expect.N, pub.N, pub.N.Cmp(expect.N))
	}
}
