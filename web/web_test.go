package web

import (
	"encoding/json"
	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
	"net/http/httptest"
	"testing"
)

const base = "http://localhost:8080"

func TestWeb(t *testing.T) {
	expect := `{"Action":"issue","Account":"xtb:test","Token":"xtb:test","Previous":"","Representative":"","Balance":100,"Link":""}`

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
	expect := `{"7XUGVPQ2EEDUK4AHUUQQMVHHOJSTXHZUTRNCZHSVGQ3IDCBUZKSA":{"Action":"issue","Account":"xtb:test1","Token":"xtb:test1","Previous":"","Representative":"","Balance":100,"Link":""},"PGFQ6WR7ZRJDKCXW4N5DHKK4Q2YUTINGD7ETFDO2BKYV2S4V2OBA":{"Action":"issue","Account":"xtb:test2","Token":"xtb:test2","Previous":"","Representative":"","Balance":50,"Link":""}}`

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
