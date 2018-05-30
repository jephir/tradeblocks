package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/jephir/tradeblocks/app"
)

func TestClient(t *testing.T) {
	store := app.NewBlockStore()

	dir, err := ioutil.TempDir("", "tradeblocks")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(dir)

	cmd := newClient(store, dir, 512)
	if err := cmd.init(); err != nil {
		t.Error(err)
		return
	}

	if _, err := cmd.register("test"); err != nil {
		t.Error(err)
		return
	}

	if err := cmd.save(); err != nil {
		t.Error(err)
		return
	}
}
