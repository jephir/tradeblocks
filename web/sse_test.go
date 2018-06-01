package web

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
)

func TestSSE(t *testing.T) {
	key, address, err := GetAddress()

	// Setup test
	store := app.NewBlockStore()
	issue := tradeblocks.NewIssueBlock(address, 100)
	errSign := issue.SignBlock(key)
	if errSign != nil {
		t.Error(errSign)
	}

	expect := `data: {"Action":"issue","Account":"` + address + `","Token":"` + address + `","Previous":"","Representative":"","Balance":100,"Link":"","Signature":"` + issue.Signature + `","Hash":"` + issue.Hash() + `"}`

	store.AddBlock(issue)
	srv := NewServer(store)

	// Send request
	req := httptest.NewRequest("GET", "/accounts", nil)
	w := httptest.NewRecorder()

	// Run test
	go func() {
		srv.ServeHTTP(w, req)
	}()
	for !w.Flushed {
	}
	for c := range srv.accountStream.clients {
		srv.accountStream.closeClient <- c
	}
	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code, got: %d, want: %d", res.StatusCode, http.StatusOK)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(string(b))
	if got != expect {
		t.Fatalf("Response was incorrect, got: %s, want: %s", got, expect)
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
	address, err := app.PublicKeyToAddress(pubKeyReader)
	if err != nil {
		return nil, "", err
	}
	return key, address, nil
}
