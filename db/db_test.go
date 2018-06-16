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

func TestInsertSwapBlock(t *testing.T) {
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

	issue := tradeblocks.NewIssueBlock("xtb:issuer", 100)
	send := tradeblocks.NewSendBlock(issue, "xtb:test", 100)
	b := tradeblocks.NewOfferBlock("xtb:test", send, "test", "xtb:counterparty", "xtb:want", 100, "", 0)
	if err := db.InsertSwapBlock(b); err != nil {
		t.Fatal(err)
	}

	check, err := db.GetSwapBlock(b.Hash())
	if err != nil {
		t.Fatal(err)
	}
	if b.Hash() != check.Hash() {
		t.Fatalf("block not found")
	}
}

func TestInsertOrderBlock(t *testing.T) {
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

	issue := tradeblocks.NewIssueBlock("xtb:issuer", 100)
	send := tradeblocks.NewSendBlock(issue, "xtb:test", 100)
	b := tradeblocks.NewCreateOrderBlock("xtb:test", send, 100, "test", false, "xtb:quote", 12.5, "", 0)
	if err := db.InsertOrderBlock(b); err != nil {
		t.Fatal(err)
	}

	check, err := db.GetOrderBlock(b.Hash())
	if err != nil {
		t.Fatal(err)
	}
	if b.Hash() != check.Hash() {
		t.Fatalf("block not found")
	}
}

func TestInsertConfirmBlock(t *testing.T) {
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

	b := tradeblocks.NewConfirmBlock(nil, "xtb:test", "xtb:addr", "123abc")
	if err := db.InsertConfirmBlock(b); err != nil {
		t.Fatal(err)
	}

	check, err := db.GetConfirmBlock(b.Hash())
	if err != nil {
		t.Fatal(err)
	}
	if b.Hash() != check.Hash() {
		t.Fatalf("block not found")
	}
}
