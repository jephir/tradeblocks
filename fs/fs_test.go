package fs

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
)

func TestFS(t *testing.T) {
	store1 := app.NewBlockStore()
	key, address := app.CreateAccount(t)
	key2, address2 := app.CreateAccount(t)

	b1 := tradeblocks.NewIssueBlock(address, 100)
	h1 := b1.Hash()

	if err := b1.SignBlock(key); err != nil {
		t.Fatal(err)
	}

	if err := store1.AddAccountBlock(b1); err != nil {
		t.Fatal(err)
	}

	b2 := tradeblocks.NewIssueBlock(address2, 100)
	h2 := b2.Hash()

	if err := b2.SignBlock(key2); err != nil {
		t.Fatal(err)
	}

	if err := store1.AddAccountBlock(b2); err != nil {
		t.Fatal(err)
	}

	dir, err := ioutil.TempDir("", "tradeblocks")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	bs1 := NewBlockStorage(store1, dir)
	if err := bs1.Save(); err != nil {
		t.Fatal(err)
	}

	store2 := app.NewBlockStore()
	bs2 := NewBlockStorage(store2, dir)
	if err := bs2.Load(); err != nil {
		t.Fatal(err)
	}

	b1check := store2.GetAccountBlock(h1)
	if b1check == nil {
		t.Fatalf("block [1] %s is missing", h1)
	}

	b2check := store2.GetAccountBlock(h2)
	if b2check == nil {
		t.Fatalf("block [2] %s is missing", h2)
	}
}
