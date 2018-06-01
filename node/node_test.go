package node

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
)

func TestBootstrapAndSync(t *testing.T) {
	key, address, err := GetAddress()
	if err != nil {
		t.Error(err)
	}
	// Create seed node
	seed, s := newNode(t, "")
	defer s.Close()

	b1 := tradeblocks.NewIssueBlock(address, 100)
	b2 := tradeblocks.NewIssueBlock(address, 50)
	errSign := b1.SignBlock(key)
	if errSign != nil {
		t.Error(errSign)
	}
	errSign = b2.SignBlock(key)
	if errSign != nil {
		t.Error(errSign)
	}

	// Add blocks to seed node
	h1 := addBlock(t, seed, b1)
	h2 := addBlock(t, seed, b2)

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
	b3 := tradeblocks.NewIssueBlock(address, 15)
	errSign = b3.SignBlock(key)
	if errSign != nil {
		t.Error(errSign)
	}
	h3 := addBlock(t, seed, b3)

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
