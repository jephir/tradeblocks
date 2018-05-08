package app

import (
	"encoding/json"
	"github.com/jephir/tradeblocks"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

var publicKey = strings.NewReader(`-----BEGIN RSA PUBLIC KEY-----
	MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDVleYQ+MOGhHVvkmzCkJrjI5CL
	4NMHwNRl7SRnElFI2+nWjYMEwSOlp5pTcHBzjRhJOx1SbLtiKRKFg1Q9wUevNeWS
	PMjB1l+LWmUTRqNTcAPQc0Vdeumjqs1P+eHERfk9MwqNsrPytvGwvNQJ05PkgLSk
	Xu58kr5iXxMABIukbQIDAQAB
	-----END RSA PUBLIC KEY-----
	`)

func TestRegister(t *testing.T) {
	if err := Register(ioutil.Discard, ioutil.Discard, "testuser", 1024); err != nil {
		t.Error(err)
	}
}

func TestIssue(t *testing.T) {
	expect := `{"Action":"issue","Account":"xtb:GxcKrfJUyh10qZQd07mytbs0VP2CUlP6ixwl+/PhDyg=","Token":"xtb:GxcKrfJUyh10qZQd07mytbs0VP2CUlP6ixwl+/PhDyg=","Previous":"","Representative":"","Balance":100,"Link":"","Hash":"","PreviousBlock":null}`
	publicKey.Seek(0, io.SeekStart)
	issue, err := Issue(publicKey, 100)
	if err != nil {
		t.Error(err)
	}
	s, err := json.Marshal(issue)
	if err != nil {
		t.Error(err)
	}
	got := string(s)
	if got != expect {
		t.Errorf("Issue was incorrect, got: %s, want: %s", got, expect)
	}
}

func TestSend(t *testing.T) {
	expect := `{"Action":"send","Account":"xtb:GxcKrfJUyh10qZQd07mytbs0VP2CUlP6ixwl+/PhDyg=","Token":"xtb:GxcKrfJUyh10qZQd07mytbs0VP2CUlP6ixwl+/PhDyg=","Previous":"","Representative":"","Balance":50,"Link":"xtb:testreceiver","Hash":"","PreviousBlock":null}`
	publicKey.Seek(0, io.SeekStart)
	address, err := PublicKeyToAddress(publicKey)
	if err != nil {
		t.Error(err)
	}
	previous := tradeblocks.NewIssueBlock(address, 100)
	publicKey.Seek(0, io.SeekStart)
	send, err := Send(publicKey, previous, "xtb:testreceiver", 50.0)
	if err != nil {
		t.Error(err)
	}
	s, err := json.Marshal(send)
	if err != nil {
		t.Error(err)
	}
	got := string(s)
	if got != expect {
		t.Errorf("Issue was incorrect, got: %s, want: %s", got, expect)
	}
}
