package app

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"
)

func TestRegister(t *testing.T) {
	if err := Register(ioutil.Discard, ioutil.Discard, "testuser", 1024); err != nil {
		t.Error(err)
	}
}

func TestIssue(t *testing.T) {
	publicKey := `-----BEGIN RSA PUBLIC KEY-----
	MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDVleYQ+MOGhHVvkmzCkJrjI5CL
	4NMHwNRl7SRnElFI2+nWjYMEwSOlp5pTcHBzjRhJOx1SbLtiKRKFg1Q9wUevNeWS
	PMjB1l+LWmUTRqNTcAPQc0Vdeumjqs1P+eHERfk9MwqNsrPytvGwvNQJ05PkgLSk
	Xu58kr5iXxMABIukbQIDAQAB
	-----END RSA PUBLIC KEY-----
	`
	expect := `{"Action":"issue","Account":"xtb:GxcKrfJUyh10qZQd07mytbs0VP2CUlP6ixwl+/PhDyg=","Token":"xtb:GxcKrfJUyh10qZQd07mytbs0VP2CUlP6ixwl+/PhDyg=","Previous":"","Representative":"","Balance":100,"Link":"","Hash":"","PreviousBlock":null}`
	issue, err := Issue(strings.NewReader(publicKey), 100)
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
