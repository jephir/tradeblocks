package web

import (
	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
	"io/ioutil"
	"os"
	"testing"
)

func TestFS(t *testing.T) {
	store1 := app.NewBlockStore()

	b1 := tradeblocks.NewIssueBlock("xtb:test1", 100)
	h1, err := app.AccountBlockHash(b1)
	if err != nil {
		t.Error(err)
	}
	store1.AccountBlocks["test1"] = b1

	b2 := tradeblocks.NewIssueBlock("xtb:test2", 100)
	h2, err := app.AccountBlockHash(b2)
	if err != nil {
		t.Error(err)
	}
	store1.AccountBlocks["test2"] = b2

	dir, err := ioutil.TempDir("", "web")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(dir)

	bs1 := newBlockStorage(store1, dir)
	if err := bs1.save(); err != nil {
		t.Error(err)
		return
	}

	store2 := app.NewBlockStore()
	bs2 := newBlockStorage(store2, dir)
	if err := bs2.load(); err != nil {
		t.Error(err)
		return
	}

	b1check, err := store2.GetBlock(h1)
	if err != nil {
		t.Error(err)
		return
	}
	if b1check == nil {
		t.Errorf("block %s is missing", h1)
		return
	}

	b2check, err := store2.GetBlock(h2)
	if err != nil {
		t.Error(err)
		return
	}
	if b2check == nil {
		t.Errorf("block %s is missing", h2)
		return
	}
}
