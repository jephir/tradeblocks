package app

import (
	"encoding/json"
	"github.com/jephir/tradeblocks"
	"testing"
)

func TestBlockStore(t *testing.T) {
	expect := `{"Action":"issue","Account":"xtb:test","Token":"xtb:test","Previous":"","Representative":"","Balance":100,"Link":"","Hash":"IUXJ2EGVQFRXKCDF4NJTUE7NTYPBPJPGYBBE4ZC6PIGBEFDFXW2Q","PreviousBlock":null}`
	s := NewBlockStore()
	b := tradeblocks.NewIssueBlock("xtb:test", 100)
	if err := s.AddBlock(b); err != nil {
		t.Error(err)
	}
	res := s.GetBlock(b.Hash)
	ss, err := json.Marshal(res)
	if err != nil {
		t.Error(err)
	}
	got := string(ss)
	if got != expect {
		t.Errorf("Issue was incorrect, got: %s, want: %s", got, expect)
	}
}
