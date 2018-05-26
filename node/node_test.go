package node

import (
	"github.com/jephir/tradeblocks"
	"net/http/httptest"
	"testing"
)

func TestBootstrapAndSync(t *testing.T) {
	// Create seed node
	seed, s := newNode(t, "")
	defer s.Close()

	// Add blocks to seed node
	h1 := addBlock(t, seed, tradeblocks.NewIssueBlock("xtb:test1", 100))
	h2 := addBlock(t, seed, tradeblocks.NewIssueBlock("xtb:test2", 50))

	// Create connecting node 1
	node1, s1 := newNode(t, s.URL)
	defer s1.Close()

	// Check that connecting node 1 has all seed blocks
	if node1.store.AccountBlocks[h1] == nil {
		t.Fatalf("N1 missing existing block %s", h1)
	}
	if node1.store.AccountBlocks[h2] == nil {
		t.Fatalf("N1 missing existing block %s", h2)
	}

	// Create connecting node 2
	node2, s2 := newNode(t, s1.URL)
	defer s2.Close()

	// Check that connecting node 2 has all node 1 blocks
	if node2.store.AccountBlocks[h1] == nil {
		t.Fatalf("N2 missing existing block %s", h1)
	}
	if node2.store.AccountBlocks[h2] == nil {
		t.Fatalf("N2 missing existing block %s", h2)
	}

	// Add block to seed node
	h3 := addBlock(t, seed, tradeblocks.NewIssueBlock("xtb:test3", 15))

	// Check that connecting node 1 has new block
	if err := seed.Sync(); err != nil {
		t.Fatal(err)
	}
	if node1.store.AccountBlocks[h3] == nil {
		t.Fatalf("N1 missing new block %s", h3)
	}

	// Check that connecting node 2 has new block
	if err := node1.Sync(); err != nil {
		t.Fatal(err)
	}
	if node2.store.AccountBlocks[h3] == nil {
		t.Fatalf("N2 missing new block %s", h3)
	}
}

func newNode(t *testing.T, bootstrapURL string) (*Node, *httptest.Server) {
	n, err := NewNode()
	if err != nil {
		t.Fatal(err)
	}
	s := httptest.NewServer(n)
	//t.Log(s.URL)
	if bootstrapURL != "" {
		if err := n.Bootstrap(s.URL, bootstrapURL); err != nil {
			t.Fatal(err)
		}
	}
	return n, s
}

func addBlock(t *testing.T, n *Node, b *tradeblocks.AccountBlock) string {
	h, err := n.store.AddBlock(b)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestMissingParent(t *testing.T) {
	// Create seed node

	// Add blocks to seed node

	// Create connecting node 1 without bootstrap

	// Add valid block to node 1

	// Check that node 1 validates and loads required blocks from seed
}
