package tradeblocks

import (
	"testing"
)

func TestHash(t *testing.T) {
	expect := "PPD6EMELYLX4VDGQ5GILR3NUZCAEX3XL7ECC2HHEOZAU6Y2AK7LQ"
	b := NewIssueBlock("xtb:test", 100)
	h := b.Hash()
	if h != expect {
		t.Fatalf("Hash was incorrect, got: %s, want: %s", h, expect)
	}
}
