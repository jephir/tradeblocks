package node

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/jephir/tradeblocks/web"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	tb "github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
)

func TestBootstrapAndSync(t *testing.T) {
	key, address, err := GetAddress()
	key2, address2, err := GetAddress()
	key3, address3, err := GetAddress()
	if err != nil {
		t.Fatal(err)
	}

	// Create seed node
	seed, s := newNode(t, "")
	defer s.Close()

	// Create connecting client
	client := web.NewClient(s.URL)

	b1, err := tb.SignedAccountBlock(tb.NewIssueBlock(address, 100), key)
	if err != nil {
		t.Fatal(err)
	}

	b2, err := tb.SignedAccountBlock(tb.NewIssueBlock(address2, 50), key2)
	if err != nil {
		t.Fatal(err)
	}

	// Add blocks to seed node
	h1 := addAccountBlock(t, client, b1)
	h2 := addAccountBlock(t, client, b2)

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

	// Ensure that the new block is persisted
	dir := node2.storage.Dir()
	p := filepath.Join(dir, "accounts", h2)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		t.Fatalf("N2 didn't persist block %s at %s", h2, p)
	}

	// Add block to seed node
	b3, err := tb.SignedAccountBlock(tb.NewIssueBlock(address3, 15), key3)
	if err != nil {
		t.Fatal(err)
	}
	h3 := addAccountBlock(t, client, b3)

	// Check that connecting node 1 has new block
	if err := seed.Sync(); err != nil {
		t.Fatal(err)
	}
	if node1.store.GetAccountBlock(h3) == nil {
		t.Fatalf("N1 missing new block %s", h3)
	}

	// Ensure that the new block is persisted
	dir1 := node1.storage.Dir()
	p1 := filepath.Join(dir1, "accounts", h3)
	if _, err := os.Stat(p1); os.IsNotExist(err) {
		t.Fatalf("N1 didn't persist block %s at %s", h3, p1)
	}

	// Check that connecting node 2 has new block
	if err := node1.Sync(); err != nil {
		t.Fatal(err)
	}
	if node2.store.GetAccountBlock(h3) == nil {
		t.Fatalf("N2 missing new block %s", h3)
	}

	// Ensure that the new block is persisted
	dir2 := node2.storage.Dir()
	p2 := filepath.Join(dir2, "accounts", h3)
	if _, err := os.Stat(p2); os.IsNotExist(err) {
		t.Fatalf("N2 didn't persist block %s at %s", h3, p2)
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

var client = &http.Client{}

func addAccountBlock(t *testing.T, c *web.Client, b *tb.AccountBlock) string {
	req, err := c.NewPostAccountBlockRequest(b)
	if err != nil {
		t.Fatal(err)
	}
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var rb tb.AccountBlock
	if err := c.DecodeAccountBlockResponse(res, &rb); err != nil {
		t.Fatal(err)
	}
	return rb.Hash()
}

func addSwapBlock(t *testing.T, c *web.Client, b *tb.SwapBlock) string {
	req, err := c.NewPostSwapBlockRequest(b)
	if err != nil {
		t.Fatal(err)
	}
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var rb tb.SwapBlock
	if err := c.DecodeSwapBlockResponse(res, &rb); err != nil {
		t.Fatal(err)
	}
	return rb.Hash()
}

func addOrderBlock(t *testing.T, c *web.Client, b *tb.OrderBlock) string {
	req, err := c.NewPostOrderBlockRequest(b)
	if err != nil {
		t.Fatal(err)
	}
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var rb tb.OrderBlock
	if err := c.DecodeOrderBlockResponse(res, &rb); err != nil {
		t.Fatal(err)
	}
	return rb.Hash()
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
	ts := app.NewBlockTestTable(t)
	p1, a1 := app.CreateAccount(t)
	p2, a2 := app.CreateAccount(t)
	p1issue := ts.AddAccountBlock(p1, tb.NewIssueBlock(a1, 100))
	p1send := ts.AddAccountBlock(p1, tb.NewSendBlock(p1issue, a2, 50))
	p2open := ts.AddAccountBlock(p2, tb.NewOpenBlockFromSend(a2, p1send, 50))
	p2send := ts.AddAccountBlock(p2, tb.NewSendBlock(p2open, a1, 25))
	p1receive := ts.AddAccountBlock(p1, tb.NewReceiveBlockFromSend(p1send, p2send, 25))

	// Test order and swap chain
	p3, a3 := app.CreateAccount(t)
	p3issue := ts.AddAccountBlock(p3, tb.NewIssueBlock(a3, 100))
	p1ordersend := ts.AddAccountBlock(p1, tb.NewSendBlock(p1receive, tb.OrderAddress(a1, "test"), 5))
	p1orderopen := ts.AddOrderBlock(p1, tb.NewCreateOrderBlock(a1, p1ordersend, 5, "test", false, a3, 1, "", 0))
	p3orderswapsend := ts.AddAccountBlock(p3, tb.NewSendBlock(p3issue, tb.SwapAddress(a3, "test"), 10))
	p3orderswapoffer := ts.AddSwapBlock(p3, tb.NewOfferBlock(a3, p3orderswapsend, "test", a1, a1, 5, "", 0))
	ts.AddOrderBlock(p1, tb.NewAcceptOrderBlock(p1orderopen, tb.SwapAddress(a3, "test"), 5))
	p1orderswapcommit := ts.AddSwapBlock(p1, tb.NewCommitBlock(p3orderswapoffer, p1receive))
	ts.AddAccountBlock(p1, tb.NewOpenBlockFromSwap(a1, p1orderswapcommit, 10))
	ts.AddAccountBlock(p3, tb.NewOpenBlockFromSwap(a3, p1orderswapcommit, 5))

	_, s1 := newNode(t, "")
	defer s1.Close()

	n2, s2 := newNode(t, s1.URL)
	defer s2.Close()

	n3, s3 := newNode(t, s2.URL)
	defer s3.Close()

	for _, tt := range ts.GetAll() {
		t.Run(tt.T, func(t *testing.T) {
			h := addBlockToNode(t, s1.URL, tt)
			switch tt.T {
			case "account":
				if n2block := n2.store.GetAccountBlock(h); n2block == nil {
					t.Fatalf("node2 missing block %s", h)
				}
				if n3block := n3.store.GetAccountBlock(h); n3block == nil {
					t.Fatalf("node3 missing block %s", h)
				}
			case "swap":
				if n2block := n2.store.GetSwapBlock(h); n2block == nil {
					t.Fatalf("node2 missing block %s", h)
				}
				if n3block := n3.store.GetSwapBlock(h); n3block == nil {
					t.Fatalf("node3 missing block %s", h)
				}
			case "order":
				if n2block := n2.store.GetOrderBlock(h); n2block == nil {
					t.Fatalf("node2 missing block %s", h)
				}
				if n3block := n3.store.GetOrderBlock(h); n3block == nil {
					t.Fatalf("node3 missing block %s", h)
				}
			}
		})
	}
}

func addBlockToNode(t *testing.T, url string, b app.TypedBlock) string {
	client := web.NewClient(url)
	switch b.T {
	case "account":
		return addAccountBlock(t, client, b.AccountBlock)
	case "swap":
		return addSwapBlock(t, client, b.SwapBlock)
	case "order":
		return addOrderBlock(t, client, b.OrderBlock)
	}
	panic("node: unknown type")
}
