package main

import (
	"testing"
)

func TestValidation(t *testing.T) {
	ok, _ := registerInputValidation([]string{"tradeblocks", "register", "test"})
	if !ok {
		t.Fatalf("register failed; expected ok")
	}
}
