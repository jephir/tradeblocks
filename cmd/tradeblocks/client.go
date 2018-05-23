package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
)

type client struct {
	dir     string
	keySize int

	store *app.BlockStore
}

func newClient(store *app.BlockStore, dir string, keySize int) *client {
	return &client{
		store:   store,
		dir:     dir,
		keySize: keySize,
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
	if err := ioutil.WriteFile("user", []byte(name), 0644); err != nil {
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
	key, err := c.openPublicKey()
	if err != nil {
		return nil, err
	}
	defer key.Close()
	return app.Issue(key, balance)
}

func (c *client) send(to string, token string, amount float64) (*tradeblocks.AccountBlock, error) {
	key, err := c.openPublicKey()
	if err != nil {
		return nil, err
	}
	defer key.Close()
	previous, err := c.getHeadBlock(key, token)
	if err != nil {
		return nil, err
	}
	return app.Send(key, previous, to, amount)
}

func (c *client) open(link string) (*tradeblocks.AccountBlock, error) {
	key, err := c.openPublicKey()
	if err != nil {
		return nil, err
	}
	defer key.Close()
	send, err := c.getBlock(link)
	if err != nil {
		return nil, err
	}
	return app.Open(key, send)
}

func (c *client) receive(link string) (*tradeblocks.AccountBlock, error) {
	key, err := c.openPublicKey()
	if err != nil {
		return nil, err
	}
	defer key.Close()
	send, err := c.getBlock(link)
	if err != nil {
		return nil, err
	}
	previous, err := c.getHeadBlock(key, send.Token)
	if err != nil {
		return nil, err
	}
	return app.Receive(key, previous, send)
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

func (c *client) getHeadBlock(publicKey io.Reader, token string) (*tradeblocks.AccountBlock, error) {
	address, err := app.PublicKeyToAddress(publicKey)
	if err != nil {
		return nil, err
	}
	// TODO Get the block from the server
	return &tradeblocks.AccountBlock{
		Action:         "open",
		Account:        address,
		Token:          token,
		Previous:       "",
		Representative: "",
		Balance:        100,
		Link:           "",
	}, nil
}

func (c *client) getBlock(hash string) (*tradeblocks.AccountBlock, error) {
	// TODO Get the block from the server
	return &tradeblocks.AccountBlock{
		Action:         "open",
		Account:        "xtb:test",
		Token:          "xtb:testtoken",
		Previous:       "",
		Representative: "",
		Balance:        100,
		Link:           "",
	}, nil
}
