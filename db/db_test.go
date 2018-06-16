package db

import (
	"github.com/mattn/go-sqlite3"
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

	dataSourceName := f.Name() + "?_foreign_keys=true"
	db, err := NewDB(dataSourceName)
	if err != nil {
		if err, ok := err.(sqlite3.Error); ok {
			t.Fatal(err.ExtendedCode.Error())
		}
		t.Fatal(err)
	}

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}
