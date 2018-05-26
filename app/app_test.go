package app

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/jephir/tradeblocks"
)

var publicKey = strings.NewReader(`-----BEGIN RSA PUBLIC KEY-----
	MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDVleYQ+MOGhHVvkmzCkJrjI5CL
	4NMHwNRl7SRnElFI2+nWjYMEwSOlp5pTcHBzjRhJOx1SbLtiKRKFg1Q9wUevNeWS
	PMjB1l+LWmUTRqNTcAPQc0Vdeumjqs1P+eHERfk9MwqNsrPytvGwvNQJ05PkgLSk
	Xu58kr5iXxMABIukbQIDAQAB
	-----END RSA PUBLIC KEY-----
	`)

func TestRegister(t *testing.T) {
	if _, err := Register(ioutil.Discard, ioutil.Discard, "testuser", 1024); err != nil {
		t.Fatal(err)
	}
}

func TestIssue(t *testing.T) {
	expect := `{"Action":"issue","Account":"xtb:GxcKrfJUyh10qZQd07mytbs0VP2CUlP6ixwl-_PhDyg","Token":"xtb:GxcKrfJUyh10qZQd07mytbs0VP2CUlP6ixwl-_PhDyg","Previous":"","Representative":"","Balance":100,"Link":""}`
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
		t.Fatalf("Issue was incorrect, got: %s, want: %s", got, expect)
	}
}

func TestSend(t *testing.T) {
	expect := `{"Action":"send","Account":"xtb:GxcKrfJUyh10qZQd07mytbs0VP2CUlP6ixwl-_PhDyg","Token":"xtb:GxcKrfJUyh10qZQd07mytbs0VP2CUlP6ixwl-_PhDyg","Previous":"R2W4NU4TXEPL76D7VTLXC5OMMGQWMBIFYCIYSJ3OK5T4CDDQJLDQ","Representative":"","Balance":50,"Link":"xtb:testreceiver"}`
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
		t.Fatalf("Issue was incorrect, got: %s, want: %s", got, expect)
	}
}

func TestOpen(t *testing.T) {
	expect := `{"Action":"open","Account":"xtb:GxcKrfJUyh10qZQd07mytbs0VP2CUlP6ixwl-_PhDyg","Token":"xtb:sender","Previous":"","Representative":"","Balance":50,"Link":"QVZSXSRJFSK23EZTG5NZF3AH3XXVEA5QDH65HNVSXDLIEBZ64Z3Q"}`
	publicKey.Seek(0, io.SeekStart)
	address, err := PublicKeyToAddress(publicKey)
	if err != nil {
		t.Fatal(err)
	}
	issue := tradeblocks.NewIssueBlock("xtb:sender", 50.0)
	send := tradeblocks.NewSendBlock(issue, address, 50.0)
	publicKey.Seek(0, io.SeekStart)
	open, err := Open(publicKey, send, 50.0)
	if err != nil {
		t.Fatal(err)
	}
	s, err := json.Marshal(open)
	if err != nil {
		t.Fatal(err)
	}
	got := string(s)
	if got != expect {
		t.Fatalf("Issue was incorrect, got: %s, want: %s", got, expect)
	}
}

func TestReceive(t *testing.T) {
	expect := `{"Action":"receive","Account":"xtb:GxcKrfJUyh10qZQd07mytbs0VP2CUlP6ixwl-_PhDyg","Token":"xtb:sender","Previous":"WZKWLP6XX5HOC7PXAA5XTP5V3R6DRAD3GKREUMJ2O35G6HHV5PZQ","Representative":"","Balance":75,"Link":"QVZSXSRJFSK23EZTG5NZF3AH3XXVEA5QDH65HNVSXDLIEBZ64Z3Q"}`
	publicKey.Seek(0, io.SeekStart)
	address, err := PublicKeyToAddress(publicKey)
	if err != nil {
		t.Fatal(err)
	}
	issue := tradeblocks.NewIssueBlock("xtb:sender", 50.0)
	send := tradeblocks.NewSendBlock(issue, address, 50.0)
	send2 := tradeblocks.NewSendBlock(send, address, 25.0)
	previous := tradeblocks.NewOpenBlock(address, send2, 25.0)
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
		t.Fatalf("Issue was incorrect, got: %s, want: %s", got, expect)
	}
}
