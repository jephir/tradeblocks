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

const publicKey = `-----BEGIN RSA PUBLIC KEY-----
	MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA2A3P6QwIxXPLNJxDXbM2
	Cw3nZ6Y2D3niv0WqWdrByjuq6rXiN4voFwSa3GHod62KqMlc+B6n9bnQKPm7YeiX
	Tr9p5i02zipK0W7vKrSvzouyRQ5HRIYKBcBtuZ++aGU72QQXUDNlXP0PcaEVbsnx
	PZatStWitkZbai+JyZOx17eNLpqXkDndgS3TXVceDm4LUSEomWDufXrdo/w0yhMa
	KDIyJNU3qltXcaLeczL8nokZSrjEQ0//zhMejUsfgonPYBalC4QE0Muv/8sJu7Ex
	XIAW/wFLgfbyWzHU3NxBPhspBONI8rEU7PC+Zw7x8KvgqqQenR+BatGzXJtHGhUe
	2LBpy9A0MjUC1Tw4EauDLTW/lk9x4audmmCWIxSJ5nAnv6x5KlAKNuw/ZbOnpxUa
	aymwB7wMiPdyveASzil9Iaq9ezlyY8WQFgWVXpkWxoJ7C9A+0TI9CG0xaAfHGzZ5
	2mrYncvR15u+NtA3jrmgns8uh2Y83wVutgJozcmN0W2TI6s2DWJr+8ZUMGJJ92T/
	Opf2gX/Y8DruBj0FG5H0yurhTJTEgb1UlGrgp+rMQH04vLPbFSM3IRtVrWEX0uip
	M7NDNlUdzbOFuMwLgWML4hQ44hA6wCQzY5RqMhEn5uoZ7MlCoc0y4syosNiHNp9z
	yAOzsR91kdD1tJs4R9daJf8CAwEAAQ==
	-----END RSA PUBLIC KEY-----
	`

const privateKey = `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`

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

func TestIssue(t *testing.T) {
	t.Skip()

	dir, err := ioutil.TempDir("", "tradeblocks")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if err := ioutil.WriteFile(filepath.Join(dir, "test.pem"), []byte(privateKey), 0640); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, "test.pub"), []byte(publicKey), 0640); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, "user"), []byte("test"), 0640); err != nil {
		t.Fatal(err)
	}

	c := &cli{
		keySize:   4096,
		serverURL: "http://localhost:8080",
		dataDir:   dir,
		out:       ioutil.Discard,
	}
	if err := c.dispatch([]string{"tradeblocks", "issue", "100"}); err != nil {
		t.Fatal(err)
	}
}

func TestDemo(t *testing.T) {
	t.Skip()
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
		x.t.Fatal(err)
	}
	return output.String()
}
