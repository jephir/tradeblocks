package main

import (
	"bytes"
	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/web"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jephir/tradeblocks/node"
)

func TestCLI(t *testing.T) {
	dir, err := ioutil.TempDir("", "tradeblocks")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Setup server
	blockdir := filepath.Join(dir, "blocks")
	if err := os.Mkdir(blockdir, 0755); err != nil {
		t.Fatal(err)
	}
	n, err := node.NewNode(blockdir)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(n)
	defer ts.Close()

	// Run test
	c := &cli{
		keySize:   1024,
		serverURL: ts.URL,
		dataDir:   dir,
		out:       ioutil.Discard,
	}
	if err := c.dispatch([]string{"tradeblocks", "register", "test"}); err != nil {
		t.Fatal(err)
	}
	if err := c.dispatch([]string{"tradeblocks", "login", "test"}); err != nil {
		t.Fatal(err)
	}
	if err := c.dispatch([]string{"tradeblocks", "issue", "100"}); err != nil {
		t.Fatal(err)
	}
}

func TestDemo(t *testing.T) {
	_, s := newNode(t, "")
	defer s.Close()

	dir, err := ioutil.TempDir("", "tradeblocks")
	if err != nil {
		t.Fatal(err)
		return
	}
	defer os.RemoveAll(dir)

	x := newExecutor(t, s.URL, dir)

	xtbAlice := x.exec("tradeblocks", "register", "alice")
	xtbAppleCoin := x.exec("tradeblocks", "register", "apple-coin")
	x.exec("tradeblocks", "login", "apple-coin")
	x.exec("tradeblocks", "issue", "1000")

	xtbSend := x.exec("tradeblocks", "send", xtbAlice, xtbAppleCoin, "50")
	x.exec("tradeblocks", "login", "alice")
	x.exec("tradeblocks", "open", xtbSend)
}

func TestLimitOrders(t *testing.T) {
	n, s := newNode(t, "")
	defer s.Close()

	dir, err := ioutil.TempDir("", "tradeblocks")
	if err != nil {
		t.Fatal(err)
		return
	}
	defer os.RemoveAll(dir)

	x := newExecutor(t, s.URL, dir)

	t1 := x.exec("tradeblocks", "register", "t1")
	t2 := x.exec("tradeblocks", "register", "t2")
	x.exec("tradeblocks", "login", "t1")
	x.exec("tradeblocks", "issue", "1000")
	x.exec("tradeblocks", "login", "t2")
	x.exec("tradeblocks", "issue", "1000")

	// Sell 100 units of t2 coin for t1 coin at 2 price per unit (200 t1)
	x.exec("tradeblocks", "sell", "100", t2, "2", t1)

	// Buy 100 units of t2 coin for t1 coin at 2 price per unit (200 t1)
	x.exec("tradeblocks", "login", "t1")
	offerHash := x.exec("tradeblocks", "buy", "100", t2, "2", t1)

	// Check resulting swap
	client := web.NewClient(s.URL)

	offerReq, err := client.NewGetBlockRequest(offerHash)
	if err != nil {
		t.Fatal(err)
	}
	offerW := httptest.NewRecorder()
	n.ServeHTTP(offerW, offerReq)
	offerRes := offerW.Result()
	var offer tradeblocks.SwapBlock
	if err := client.DecodeSwapBlockResponse(offerRes, &offer); err != nil {
		t.Fatal(err)
	}

	commitReq, err := client.NewGetSwapHeadRequest(t1, offer.ID)
	if err != nil {
		t.Fatal(err)
	}
	commitW := httptest.NewRecorder()
	n.ServeHTTP(commitW, commitReq)
	commitRes := commitW.Result()
	var commit tradeblocks.SwapBlock
	if err := client.DecodeSwapBlockResponse(commitRes, &commit); err != nil {
		t.Fatal(err)
	}

	if commit.Action != "commit" {
		t.Fatalf("expected action 'commit', got '%s'", commit.Action)
	}
	if commit.Account != strings.TrimSpace(t1) {
		t.Fatalf("expected account '%s', got '%s'", strings.TrimSpace(t1), commit.Account)
	}
	if commit.Counterparty != strings.TrimSpace(t2) {
		t.Fatalf("expected counterparty '%s', got '%s'", strings.TrimSpace(t2), commit.Counterparty)
	}
	if commit.Quantity != 100.0 {
		t.Fatalf("expected quantity '%f', got '%f", 100.0, commit.Quantity)
	}
	if commit.Token != strings.TrimSpace(t1) {
		t.Fatalf("expected token '%s', got '%s'", strings.TrimSpace(t1), commit.Token)
	}
	if commit.Right == "" {
		t.Fatal("expected right to not be empty")
	}

	// Check that right is valid
	sendReq, err := client.NewGetOrderHeadRequest(t2, offer.ID)
	if err != nil {
		t.Fatal(err)
	}
	sendW := httptest.NewRecorder()
	n.ServeHTTP(sendW, sendReq)
	sendRes := sendW.Result()
	var send tradeblocks.OrderBlock
	if err := client.DecodeOrderBlockResponse(sendRes, &send); err != nil {
		t.Fatal(err)
	}

	link := tradeblocks.SwapAddress(strings.TrimSpace(t1), offer.ID)
	if send.Link != link {
		t.Fatalf("expected link '%s', got '%s'", link, send.Link)
	}
}

func newNode(t *testing.T, bootstrapURL string) (*node.Node, *httptest.Server) {
	dir, err := ioutil.TempDir("", "tradeblocks")
	if err != nil {
		t.Fatal(err)
	}

	n, err := node.NewNode(dir)
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

type executor struct {
	t *testing.T
	c *cli
}

func newExecutor(t *testing.T, serverURL, dataDir string) *executor {

	c := &cli{
		keySize:   1024,
		serverURL: serverURL,
		dataDir:   dataDir,
	}
	return &executor{
		t: t,
		c: c,
	}
}

func (x *executor) exec(cmd ...string) string {
	output := &bytes.Buffer{}
	x.c.out = output
	if err := x.c.dispatch(cmd); err != nil {
		x.t.Fatalf("%s: %s", strings.Join(cmd, " "), err.Error())
	}
	return strings.TrimSpace(output.String())
}
