package web

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
)

const base = "http://localhost:8080"

var key, err = rsa.GenerateKey(rand.Reader, 512)

func TestWeb(t *testing.T) {
	// Setup keys
	key, address, err := GetAddress()
	if err != nil {
		t.Fatal(err)
	}
	// Setup test
	store := app.NewBlockStore()
	srv := NewServer(store)
	client := NewClient(base)

	// Create request
	b, err := tradeblocks.SignedAccountBlock(tradeblocks.NewIssueBlock(address, 100), key)
	if err != nil {
		t.Fatal(err)
	}
	expect := `{"Action":"issue","Account":"` + address + `","Token":"` + address + `","Previous":"","Representative":"","Balance":100,"Link":"","Signature":"` + b.Signature + `"}`

	req, err := client.NewPostAccountRequest(b)
	if err != nil {
		t.Fatal(err)
	}

	// Send request
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	res := w.Result()
	result, err := client.DecodeAccountResponse(res)
	if err != nil {
		t.Fatal(err)
	}

	// Check result
	s, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	got := string(s)
	if got != expect {
		t.Fatalf("Response was incorrect, got: %s, want: %s", got, expect)
	}
}

func TestBootstrap(t *testing.T) {
	key, address, err := GetAddress()
	key2, address2, err := GetAddress()
	if err != nil {
		t.Error(err)
	}

	// Create root server
	rs := app.NewBlockStore()
	b1, err := tradeblocks.SignedAccountBlock(tradeblocks.NewIssueBlock(address, 100), key)
	if err != nil {
		t.Fatal(err)
	}
	b2, err := tradeblocks.SignedAccountBlock(tradeblocks.NewIssueBlock(address2, 50), key2)
	if err != nil {
		t.Fatal(err)
	}

	rs.AddBlock(b1)
	rs.AddBlock(b2)
	srv := NewServer(rs)

	// Create connecting server
	client := NewClient(base)
	req, err := client.NewGetBlocksRequest()
	if err != nil {
		t.Fatal(err)
	}

	// Send request
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	res := w.Result()
	result, err := client.DecodeGetBlocksResponse(res)
	if err != nil {
		t.Fatal(err)
	}

	// Check result
	r1, ok := result[b1.Hash()]
	if !ok {
		t.Fatalf("missing block b1 '%s'", b1.Hash())
	}
	if err := r1.Equals(b1); err != nil {
		t.Fatal(err)
	}
	r2, ok := result[b2.Hash()]
	if !ok {
		t.Fatalf("missing block b2 '%s'", b2.Hash())
	}
	if err := r2.Equals(b2); err != nil {
		t.Fatal(err)
	}
}
