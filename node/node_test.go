package node

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
)

func TestBootstrapAndSync(t *testing.T) {
	t.Skip("TODO Re-implement sync")

	key, address, err := GetAddress()
	key2, address2, err := GetAddress()
	key3, address3, err := GetAddress()
	if err != nil {
		t.Fatal(err)
	}
	// Create seed node
	seed, s := newNode(t, "")
	defer s.Close()

	b1, err := tradeblocks.SignedAccountBlock(tradeblocks.NewIssueBlock(address, 100), key)
	if err != nil {
		t.Fatal(err)
	}

	b2, err := tradeblocks.SignedAccountBlock(tradeblocks.NewIssueBlock(address2, 50), key2)
	if err != nil {
		t.Fatal(err)
	}

	// Add blocks to seed node
	h1 := addBlock(t, seed, b1)
	h2 := addBlock(t, seed, b2)

	// Create connecting node 1
	node1, s1 := newNode(t, s.URL)
	defer s1.Close()

	// Check that connecting node 1 has all seed blocks
	if node1.store.GetAccountBlock(h1) == nil {
		t.Fatalf("N1 missing existing block %s", h1)
	}
	if node1.store.GetAccountBlock(h2) == nil {
		t.Fatalf("N1 missing existing block %s", h2)
	}

	// Create connecting node 2
	node2, s2 := newNode(t, s1.URL)
	defer s2.Close()

	// Check that connecting node 2 has all node 1 blocks
	if node2.store.GetAccountBlock(h1) == nil {
		t.Fatalf("N2 missing existing block %s", h1)
	}
	if node2.store.GetAccountBlock(h2) == nil {
		t.Fatalf("N2 missing existing block %s", h2)
	}

	// Add block to seed node
	b3, err := tradeblocks.SignedAccountBlock(tradeblocks.NewIssueBlock(address3, 15), key3)
	if err != nil {
		t.Fatal(err)
	}
	h3 := addBlock(t, seed, b3)

	// Check that connecting node 1 has new block
	if err := seed.Sync(); err != nil {
		t.Fatal(err)
	}
	if node1.store.GetAccountBlock(h3) == nil {
		t.Fatalf("N1 missing new block %s", h3)
	}

	// Check that connecting node 2 has new block
	if err := node1.Sync(); err != nil {
		t.Fatal(err)
	}
	if node2.store.GetAccountBlock(h3) == nil {
		t.Fatalf("N2 missing new block %s", h3)
	}

	// Ensure that the new block is persisted
	dir := node2.storage.Dir()
	if _, err := os.Stat(filepath.Join(dir, h3)); os.IsNotExist(err) {
		t.Fatalf("N2 didn't persist block %s", h3)
	}
}

func newNode(t *testing.T, bootstrapURL string) (*Node, *httptest.Server) {
	dir, err := ioutil.TempDir("", "tradeblocks")
	if err != nil {
		t.Fatal(err)
	}

	n, err := NewNode(dir)
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
	err := n.store.AddAccountBlock(b)
	if err != nil {
		t.Fatal(err)
	}
	return b.Hash()
}

func TestMissingParent(t *testing.T) {
	// Create seed node

	// Add blocks to seed node

	// Create connecting node 1 without bootstrap

	// Add valid block to node 1

	// Check that node 1 validates and loads required blocks from seed
}

func GetAddress() (*rsa.PrivateKey, string, error) {
	var key, err = rsa.GenerateKey(rand.Reader, 512)
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, "", err
	}
	p := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	pubKeyReader := bytes.NewReader(p)
	address, err := app.PublicKeyToAddress(pubKeyReader)
	if err != nil {
		return nil, "", err
	}
	return key, address, nil
}

func TestSyncAllBlockTypes(t *testing.T) {
	t.Skip("TODO Re-implement sync")

	allBlocksTest := app.NewBlockTestTable(t)
	p1, a1 := app.CreateAccount(t)
	p2, a2 := app.CreateAccount(t)
	p1issue := allBlocksTest.AddAccountBlock(p1, tradeblocks.NewIssueBlock(a1, 100))
	p1send := allBlocksTest.AddAccountBlock(p1, tradeblocks.NewSendBlock(p1issue, a2, 50))
	p2open := allBlocksTest.AddAccountBlock(p2, tradeblocks.NewOpenBlockFromSend(a2, p1send, 50))
	p2send := allBlocksTest.AddAccountBlock(p2, tradeblocks.NewSendBlock(p2open, a1, 25))
	p1receive := allBlocksTest.AddAccountBlock(p1, tradeblocks.NewReceiveBlockFromSend(p1send, p2send, 25))

	// Test swap chain
	p3, a3 := app.CreateAccount(t)
	p3issue := allBlocksTest.AddAccountBlock(p3, tradeblocks.NewIssueBlock(a3, 100))
	p1swapsend := allBlocksTest.AddAccountBlock(p1, tradeblocks.NewSendBlock(p1receive, tradeblocks.SwapAddress(a1, "test"), 10))
	p1swapoffer := allBlocksTest.AddSwapBlock(p1, tradeblocks.NewOfferBlock(a1, p1swapsend, "test", a3, a3, 15, "", 0))
	p3swapsend := allBlocksTest.AddAccountBlock(p3, tradeblocks.NewSendBlock(p3issue, tradeblocks.SwapAddress(a1, "test"), 15))
	p3swapcommit := allBlocksTest.AddSwapBlock(p3, tradeblocks.NewCommitBlock(p1swapoffer, p3swapsend))
	/* p1swapreceive := */ allBlocksTest.AddAccountBlock(p1, tradeblocks.NewOpenBlockFromSwap(a1, p3swapcommit, 15))
	/* p3swapreceive := */ allBlocksTest.AddAccountBlock(p3, tradeblocks.NewOpenBlockFromSwap(a3, p3swapcommit, 10))

	n1, s1 := newNode(t, "")
	defer s1.Close()

	n2, s2 := newNode(t, s1.URL)
	defer s2.Close()

	n3, s3 := newNode(t, s2.URL)
	defer s3.Close()

	for _, tt := range allBlocksTest.GetAll() {
		t.Run(tt.T, func(t *testing.T) {
			h := addBlockToNode(t, n1, tt)
			if n2block := n2.store.GetAccountBlock(h); n2block == nil {
				t.Fatalf("node2 missing block %s", h)
			}
			if n3block := n3.store.GetAccountBlock(h); n3block == nil {
				t.Fatalf("node3 missing block %s", h)
			}
		})
	}

}

func addBlockToNode(t *testing.T, n *Node, b app.TypedBlock) string {
	switch b.T {
	case "account":
		if err := n.store.AddAccountBlock(b.AccountBlock); err != nil {
			t.Fatal(err)
		}
		return b.AccountBlock.Hash()
	case "swap":
		if err := n.store.AddSwapBlock(b.SwapBlock); err != nil {
			t.Fatal(err)
		}
		return b.SwapBlock.Hash()
	case "order":
		if err := n.store.AddOrderBlock(b.OrderBlock); err != nil {
			t.Fatal(err)
		}
		return b.OrderBlock.Hash()
	}
	panic("node: unknown type")
}
