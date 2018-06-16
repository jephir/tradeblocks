package db

import (
	"github.com/jephir/tradeblocks"
	"io/ioutil"
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	f, err := ioutil.TempFile("", "tradeblocks")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	dataSourceName := f.Name()
	db, err := NewDB(dataSourceName)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}
func TestInsertAccountBlock(t *testing.T) {
	f, err := ioutil.TempFile("", "tradeblocks")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	dataSourceName := f.Name()
	db, err := NewDB(dataSourceName)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	b := tradeblocks.NewIssueBlock("xtb:test", 500)
	if err := db.InsertAccountBlock(b); err != nil {
		t.Fatal(err)
	}

	check, err := db.GetAccountBlock(b.Hash())
	if err != nil {
		t.Fatal(err)
	}
	if err := b.Equals(check); err != nil {
		t.Fatal(err)
	}
}
