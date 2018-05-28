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

var accountAddress = "LS0tLS1CRUdJTiBSU0EgUFVCTElDIEtFWS0tLS0tCglNSUdmTUEwR0NTcUdTSWIzRFFFQkFRVUFBNEdOQURDQmlRS0JnUURWbGVZUStNT0doSFZ2a216Q2tKcmpJNUNMCgk0Tk1Id05SbDdTUm5FbEZJMituV2pZTUV3U09scDVwVGNIQnpqUmhKT3gxU2JMdGlLUktGZzFROXdVZXZOZVdTCglQTWpCMWwrTFdtVVRScU5UY0FQUWMwVmRldW1qcXMxUCtlSEVSZms5TXdxTnNyUHl0dkd3dk5RSjA1UGtnTFNrCglYdTU4a3I1aVh4TUFCSXVrYlFJREFRQUIKCS0tLS0tRU5EIFJTQSBQVUJMSUMgS0VZLS0tLS0KCQ"

func TestRegister(t *testing.T) {
	if _, err := Register(ioutil.Discard, ioutil.Discard, "testuser", 1024); err != nil {
		t.Fatal(err)
	}
}

func TestIssue(t *testing.T) {
	expect := `{"Action":"issue","Account":"xtb:` + accountAddress + `","Token":"xtb:` + accountAddress + `","Previous":"","Representative":"","Balance":100,"Link":"","Signature":null}`
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
	expect := `{"Action":"send","Account":"xtb:` + accountAddress + `","Token":"xtb:` + accountAddress + `","Previous":"OAPTOUS6G3HJ4YLRG7WNSZ6CH26R5HPD2PMRTJ7ABYNHBET5NSIA","Representative":"","Balance":50,"Link":"xtb:testreceiver","Signature":null}`
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
	expect := `{"Action":"open","Account":"xtb:` + accountAddress + `","Token":"xtb:sender","Previous":"","Representative":"","Balance":50,"Link":"KX4ZH6X3RALXWNJ4ULS2D64DUOQRWMSGUTRCP3B3M2GJUDS633SA","Signature":null}`
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
	expect := `{"Action":"receive","Account":"xtb:` + accountAddress + `","Token":"xtb:sender","Previous":"TJA6K3SXZURF76HGIULJZRDWAZ7PL62GAXCWEFVCNAXWKRL4K2JQ","Representative":"","Balance":75,"Link":"KX4ZH6X3RALXWNJ4ULS2D64DUOQRWMSGUTRCP3B3M2GJUDS633SA","Signature":null}`
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
		t.Fatalf("Issue was incorrect, got: %s,\nwant: %s", got, expect)
	}
}
