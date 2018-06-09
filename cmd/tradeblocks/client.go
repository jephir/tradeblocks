package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/jephir/tradeblocks/web"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
)

type client struct {
	dir     string
	keySize int
	api     *web.Client
	http    *http.Client
}

func newClient(dir, host string, keySize int) *client {
	return &client{
		dir:     dir,
		keySize: keySize,
		api:     web.NewClient(host),
		http:    &http.Client{},
	}
}

func (c *client) badInputs(funcName string, additionalInfo string) error {
	fmt.Printf("Error in function %s \n", funcName)
	fmt.Printf("Invalid inputs: %v \n", strings.Join(os.Args[2:], ", "))
	fmt.Printf("Additional information: \n%s", additionalInfo)
	os.Exit(1)
	return nil
}

func (c *client) register(name string) (address string, err error) {
	privateKeyPath := filepath.Join(c.dir, name+".pem")
	privateKeyFile, err := os.Create(privateKeyPath)
	if err != nil {
		return
	}
	defer privateKeyFile.Close()
	publicKeyPath := filepath.Join(c.dir, name+".pub")
	publicKeyFile, err := os.Create(publicKeyPath)
	if err != nil {
		return
	}
	defer publicKeyFile.Close()
	address, err = app.Register(privateKeyFile, publicKeyFile, name, c.keySize)
	if err != nil {
		return
	}
	if err := publicKeyFile.Close(); err != nil {
		return "", err
	}
	if err := privateKeyFile.Close(); err != nil {
		return "", err
	}
	return
}

func (c *client) login(name string) (address string, err error) {
	if err := ioutil.WriteFile(filepath.Join(c.dir, "user"), []byte(name), 0644); err != nil {
		return "", err
	}
	f, err := c.openPublicKey()
	if err != nil {
		return
	}
	defer f.Close()
	address, err = app.PublicKeyToAddress(f)
	return
}

func (c *client) issue(balance float64) (*tradeblocks.AccountBlock, error) {
	// get the keys
	publicKey, err := c.openPublicKey()
	if err != nil {
		return nil, err
	}
	privateKey, err := c.openPrivateKey()
	if err != nil {
		return nil, err
	}
	defer publicKey.Close()
	defer privateKey.Close()

	// create the Issue block
	issue, err := app.Issue(publicKey, balance)
	if err != nil {
		return nil, err
	}

	// add the signature
	if err := sign(privateKey, issue); err != nil {
		return nil, err
	}

	if err := c.postAccountBlock(issue); err != nil {
		return nil, err
	}

	return issue, nil
}

func (c *client) send(to string, token string, amount float64) (*tradeblocks.AccountBlock, error) {
	// get the keys
	publicKey, err := c.openPublicKey()
	if err != nil {
		return nil, err
	}
	privateKey, err := c.openPrivateKey()
	if err != nil {
		return nil, err
	}
	defer publicKey.Close()
	defer privateKey.Close()

	previous, err := c.getHeadBlock(publicKey, token)
	if err != nil {
		return nil, err
	}

	// create the send block
	send, err := app.Send(publicKey, previous, to, amount)
	if err != nil {
		return nil, err
	}

	// add the signature
	if err := sign(privateKey, send); err != nil {
		return nil, err
	}

	if err := c.postAccountBlock(send); err != nil {
		return nil, err
	}

	return send, nil
}

func (c *client) open(link string) (*tradeblocks.AccountBlock, error) {
	// get the keys
	publicKey, err := c.openPublicKey()
	if err != nil {
		return nil, err
	}
	privateKey, err := c.openPrivateKey()
	if err != nil {
		return nil, err
	}
	defer publicKey.Close()
	defer privateKey.Close()

	// get the linked send
	send, err := c.getBlock(link)
	if err != nil {
		return nil, err
	}
	sendParent, err := c.getBlock(send.Previous)
	if err != nil {
		return nil, err
	}
	balance := sendParent.Balance - send.Balance

	// create the Open
	open, err := app.Open(publicKey, send, balance)
	if err != nil {
		return nil, err
	}

	// add the signature
	if err := sign(privateKey, open); err != nil {
		return nil, err
	}

	if err := c.postAccountBlock(open); err != nil {
		return nil, err
	}

	return open, nil
}

func (c *client) receive(link string) (*tradeblocks.AccountBlock, error) {
	// get the keys
	publicKey, err := c.openPublicKey()
	if err != nil {
		return nil, err
	}
	privateKey, err := c.openPrivateKey()
	if err != nil {
		return nil, err
	}
	defer publicKey.Close()
	defer privateKey.Close()

	// get the linked send
	send, err := c.getBlock(link)
	if err != nil {
		return nil, err
	}
	sendParent, err := c.getBlock(send.Previous)
	if err != nil {
		return nil, err
	}
	balance := sendParent.Balance - send.Balance

	// get the previous block on this chain
	previous, err := c.getHeadBlock(publicKey, send.Token)
	if err != nil {
		return nil, err
	}

	// create the receive
	receive, err := app.Receive(publicKey, previous, send, balance)
	if err != nil {
		return nil, err
	}

	// add the signature
	if err := sign(privateKey, receive); err != nil {
		return nil, err
	}

	if err := c.postAccountBlock(receive); err != nil {
		return nil, err
	}

	return receive, nil
}

func (c *client) openPublicKey() (*os.File, error) {
	userPath := filepath.Join(c.dir, "user")
	user, err := ioutil.ReadFile(userPath)
	if err != nil {
		return nil, err
	}
	p := filepath.Join(c.dir, string(user)+".pub")
	return os.Open(p)
}

func (c *client) openPrivateKey() (*os.File, error) {
	userPath := filepath.Join(c.dir, "user")
	user, err := ioutil.ReadFile(userPath)
	if err != nil {
		return nil, err
	}
	p := filepath.Join(c.dir, string(user)+".pem")
	return os.Open(p)
}

func openPublicKey() (*os.File, error) {
	user, err := ioutil.ReadFile("user")
	if err != nil {
		return nil, err
	}
	return os.Open(string(user) + ".pub")
}

func openPrivateKey() (*os.File, error) {
	user, err := ioutil.ReadFile("user")
	if err != nil {
		return nil, err
	}
	return os.Open(string(user) + ".pem")
}

func (c *client) getHeadBlock(publicKey io.Reader, token string) (*tradeblocks.AccountBlock, error) {
	address, err := app.PublicKeyToAddress(publicKey)
	if err != nil {
		return nil, err
	}
	r, err := c.api.NewGetAccountHeadRequest(address, token)
	if err != nil {
		return nil, err
	}
	res, err := c.http.Do(r)
	if err != nil {
		return nil, err
	}
	var result tradeblocks.AccountBlock
	if err := c.api.DecodeAccountBlockResponse(res, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) getBlock(hash string) (*tradeblocks.AccountBlock, error) {
	r, err := c.api.NewGetBlockRequest(hash)
	if err != nil {
		return nil, err
	}
	res, err := c.http.Do(r)
	if err != nil {
		return nil, err
	}
	var result tradeblocks.AccountBlock
	if err := c.api.DecodeAccountBlockResponse(res, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) postAccountBlock(b *tradeblocks.AccountBlock) error {
	req, err := c.api.NewPostAccountBlockRequest(b)
	if err != nil {
		return err
	}

	res, err := c.http.Do(req)
	if err != nil {
		return err
	}

	var rb tradeblocks.AccountBlock
	if err := c.api.DecodeAccountBlockResponse(res, &rb); err != nil {
		return err
	}

	return nil
}

// in that case, you can add a param for your validator factory that receives the BlockStore
// move it to the validator
func (c *client) alreadyLinked(hash string) bool {
	// TODO the check needs to be done on the node by iterating over all the existing
	// blocks. if the block specifies the same
	// link as the specified hash, then it's already linked

	// block, err := c.getBlock(hash)
	// if err != nil || block == nil {
	// 	return true
	// }
	return false
}

func parsePrivateKey(r io.Reader) (*rsa.PrivateKey, error) {
	keyBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	p, _ := pem.Decode(keyBytes)
	if p == nil {
		return nil, errors.New("client: no PEM data found")
	}

	return x509.ParsePKCS1PrivateKey(p.Bytes)
}

func sign(privateKey io.Reader, b *tradeblocks.AccountBlock) error {
	b.Normalize()
	priv, err := parsePrivateKey(privateKey)
	if err != nil {
		return err
	}
	return b.SignBlock(priv)
}
