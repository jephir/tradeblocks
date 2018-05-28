package app

import (
	"encoding/json"
	"testing"

	"github.com/jephir/tradeblocks"
)

func TestBlockStore(t *testing.T) {
	expect := `{"Action":"issue","Account":"xtb:test","Token":"xtb:test","Previous":"","Representative":"","Balance":100,"Link":"","Signature":""}`
	s := NewBlockStore()
	b := tradeblocks.NewIssueBlock("xtb:test", 100)
	if _, err := s.AddBlock(b); err != nil {
		t.Error(err)
	}
	res, _ := s.GetBlock(b.Hash())
	ss, err := json.Marshal(res)
	if err != nil {
		t.Error(err)
	}
	got := string(ss)
	if got != expect {
		t.Errorf("Issue was incorrect, got: %s, want: %s", got, expect)
	}
}
