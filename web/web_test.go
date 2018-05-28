package web

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
)

const base = "http://localhost:8080"

func TestWeb(t *testing.T) {
	expect := `{"Action":"issue","Account":"xtb:test","Token":"xtb:test","Previous":"","Representative":"","Balance":100,"Link":"","Signature":null}`

	// Setup test
	store := app.NewBlockStore()
	srv := NewServer(store)
	client := NewClient(base)

	// Create request
	b := tradeblocks.NewIssueBlock("xtb:test", 100)
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
	expect := `{"NUOPXEVRVJQ5OCCK5HHEKKHWMJWRTNUI5YR2KND4PATWXY3UEPBA":{"Action":"issue","Account":"xtb:test1","Token":"xtb:test1","Previous":"","Representative":"","Balance":100,"Link":"","Signature":null},"Q5PF7QU2YNHUT4VVKXXWO2GNRYKVJE43BDAJHU4YD6UHKJL2I6IA":{"Action":"issue","Account":"xtb:test2","Token":"xtb:test2","Previous":"","Representative":"","Balance":50,"Link":"","Signature":null}}`

	// Create root server
	rs := app.NewBlockStore()
	b1 := tradeblocks.NewIssueBlock("xtb:test1", 100)
	b2 := tradeblocks.NewIssueBlock("xtb:test2", 50)
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
	s, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	got := string(s)
	if got != expect {
		t.Fatalf("Response was incorrect, got: %s, want: %s", got, expect)
	}
}
