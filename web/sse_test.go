package web

import (
	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSSE(t *testing.T) {
	expect := `data: {"Action":"issue","Account":"xtb:test","Token":"xtb:test","Previous":"","Representative":"","Balance":100,"Link":"","Hash":"PPD6EMELYLX4VDGQ5GILR3NUZCAEX3XL7ECC2HHEOZAU6Y2AK7LQ"}`

	// Setup test
	store := app.NewBlockStore()
	store.AddBlock(tradeblocks.NewIssueBlock("xtb:test", 100))
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
