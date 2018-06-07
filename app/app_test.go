package app

import (
	"encoding/json"
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

var accountAddress = "Ci0tLS0tQkVHSU4gUlNBIFBVQkxJQyBLRVktLS0tLQpNSUdmTUEwR0NTcUdTSWIzRFFFQkFRVUFBNEdOQURDQmlRS0JnUURWbGVZUStNT0doSFZ2a216Q2tKcmpJNUNMCjROTUh3TlJsN1NSbkVsRkkyK25XallNRXdTT2xwNXBUY0hCempSaEpPeDFTYkx0aUtSS0ZnMVE5d1Vldk5lV1MKUE1qQjFsK0xXbVVUUnFOVGNBUFFjMFZkZXVtanFzMVArZUhFUmZrOU13cU5zclB5dHZHd3ZOUUowNVBrZ0xTawpYdTU4a3I1aVh4TUFCSXVrYlFJREFRQUIKLS0tLS1FTkQgUlNBIFBVQkxJQyBLRVktLS0tLQo"

func TestRegister(t *testing.T) {
	if _, err := Register(ioutil.Discard, ioutil.Discard, "testuser", 1024); err != nil {
		t.Fatal(err)
	}
}

func TestIssue(t *testing.T) {
	expect := `{"Action":"issue","Account":"xtb:` + accountAddress + `","Token":"xtb:` + accountAddress + `","Previous":"","Representative":"","Balance":100,"Link":"","Signature":""}`
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
	previousText := "4RGSBQSWQKRQXQP2FSLIOGHS6SJ6JNWW4SLKKGUSPRF4CTX7S24Q"
	expect := `{"Action":"send","Account":"xtb:` + accountAddress + `","Token":"xtb:` + accountAddress + `","Previous":"` + previousText + `","Representative":"","Balance":50,"Link":"xtb:testreceiver","Signature":""}`
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

func TestOpen(t *testing.T) {
	linkText := "SSKW5XVFSPHSHKLWWBDZ2XZ5TUTUINEOXRHGMXNJJZB6ATZCDUAQ"
	expect := `{"Action":"open","Account":"xtb:` + accountAddress + `","Token":"xtb:sender","Previous":"","Representative":"","Balance":50,"Link":"` + linkText + `","Signature":""}`
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
		t.Fatalf("Issue was incorrect, got: %s,\nwant: %s", got, expect)
	}
}

func TestReceive(t *testing.T) {
	previousText := "XNCHOTIUY6O5GUOYVLS2TSD5E2U4R75QWD5AWD6SP2CFYH3G7CNA"
	linkText := "SSKW5XVFSPHSHKLWWBDZ2XZ5TUTUINEOXRHGMXNJJZB6ATZCDUAQ"
	expect := `{"Action":"receive","Account":"xtb:` + accountAddress + `","Token":"xtb:sender","Previous":"` + previousText + `","Representative":"","Balance":75,"Link":"` + linkText + `","Signature":""}`
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
