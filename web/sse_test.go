package web

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
)

func TestSSE(t *testing.T) {
	expect := `data: {"Action":"issue","Account":"xtb:test","Token":"xtb:test","Previous":"","Representative":"","Balance":100,"Link":"","Signature":null,"Hash":"GM6XD5BX4IYD5Z2RP5YMO457M7QLNRU4HDFCDUNTZ647PFA3YG5A"}`

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
