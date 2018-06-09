package web

import (
	"encoding/json"
	"github.com/jephir/tradeblocks/tradeblockstest"
	"net/http/httptest"
	"testing"

	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
)

const base = "http://localhost:8080"

func TestWeb(t *testing.T) {
	// Setup keys
	p, a := tradeblockstest.CreateAccount(t)

	// Setup test
	store := app.NewBlockStore2()
	srv := NewServer(store)
	client := NewClient(base)

	// Create request
	b, err := tradeblocks.SignedAccountBlock(tradeblocks.NewIssueBlock(a, 100), p)
	if err != nil {
		t.Fatal(err)
	}
	expect := `{"Action":"issue","Account":"` + a + `","Token":"` + a + `","Previous":"","Representative":"","Balance":100,"Link":"","Signature":"` + b.Signature + `"}`

	req, err := client.NewPostAccountBlockRequest(b)
	if err != nil {
		t.Fatal(err)
	}

	// Send request
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	res := w.Result()
	var result tradeblocks.AccountBlock
	if err := client.DecodeAccountBlockResponse(res, &result); err != nil {
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
	p, a := tradeblockstest.CreateAccount(t)

	// Create root server
	rs := app.NewBlockStore2()
	b1, err := tradeblocks.SignedAccountBlock(tradeblocks.NewIssueBlock(a, 100), p)
	if err != nil {
		t.Fatal(err)
	}
	b2, err := tradeblocks.SignedAccountBlock(tradeblocks.NewIssueBlock(a, 50), p)
	if err != nil {
		t.Fatal(err)
	}

	if err := rs.AddAccountBlock(b1); err != nil {
		t.Fatal(err)
	}
	if err := rs.AddAccountBlock(b2); err != nil {
		t.Fatal(err)
	}
	srv := NewServer(rs)

	// Create connecting server
	client := NewClient(base)
	req, err := client.NewGetAccountBlocksRequest()
	if err != nil {
		t.Fatal(err)
	}

	// Send request
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	res := w.Result()
	result, err := client.DecodeGetAccountBlocksResponse(res)
	if err != nil {
		t.Fatal(err)
	}

	// Check result
	r1 := result[b1.Hash()].AccountBlock
	if err := r1.Equals(b1); err != nil {
		t.Fatal(err)
	}
	r2 := result[b2.Hash()].AccountBlock
	if err := r2.Equals(b2); err != nil {
		t.Fatal(err)
	}
}
