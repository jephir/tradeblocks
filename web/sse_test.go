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

	p, a := app.CreateAccount(t)

	// Setup test
	store := app.NewBlockStore2()
	issue, err := tradeblocks.SignedAccountBlock(tradeblocks.NewIssueBlock(a, 100), p)
	if err != nil {
		t.Fatal(err)
	}

	expect := `data: {"Block":{"Action":"issue","Account":"` + a + `","Token":"` + a + `","Previous":"","Representative":"","Balance":100,"Link":"","Signature":"` + issue.Signature + `"},"Hash":"` + issue.Hash() + `"}`

	if err := store.AddAccountBlock(issue); err != nil {
		t.Fatal(err)
	}
	srv := NewServer(store)

	// Send request
	req := httptest.NewRequest("GET", "/blocks", nil)
	q := req.URL.Query()
	q.Add("stream", "1")
	req.URL.RawQuery = q.Encode()

	w := httptest.NewRecorder()

	// Run test
	go func() {
		srv.ServeHTTP(w, req)
	}()
	for !w.Flushed {
	}
	for c := range srv.blockStream.clients {
		srv.blockStream.closeClient <- c
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
		t.Fatalf("Response was incorrect, got: \n%s, want: \n%s", got, expect)
	}
}
