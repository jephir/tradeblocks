package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestCLI(t *testing.T) {
	dir, err := ioutil.TempDir("", "tradeblocks")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(dir)

	c := &cli{
		keySize:   4096,
		serverURL: "http://localhost:8080",
		dataDir:   dir,
	}
	if err := c.dispatch([]string{"register", "test"}); err != nil {
		t.Fatal(err)
	}
}
