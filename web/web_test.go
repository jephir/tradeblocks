package web

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/jephir/tradeblocks"
)

const base = "http://localhost:8080"

func TestWeb(t *testing.T) {
	expect := `{"Action":"issue","Account":"xtb:test","Token":"xtb:test","Previous":"","Representative":"","Balance":100,"Link":"","Hash":"IUXJ2EGVQFRXKCDF4NJTUE7NTYPBPJPGYBBE4ZC6PIGBEFDFXW2Q","PreviousBlock":null}`
	srv := NewServer()
	client := NewClient(base)
	b := tradeblocks.NewIssueBlock("xtb:test", 100)
	req, err := client.NewAccountBlockRequest(b)
	if err != nil {
		t.Error(err)
	}
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	res := w.Result()
	result, err := client.DecodeResponse(res)
	if err != nil {
		t.Error(err)
	}
	s, err := json.Marshal(result)
	if err != nil {
		t.Error(err)
	}
	got := string(s)
	if got != expect {
		t.Errorf("Response was incorrect, got: %s, want: %s", got, expect)
	}
}