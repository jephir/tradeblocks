package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
	"github.com/jephir/tradeblocks/fs"
)

type client struct {
	dir     string
	keySize int

	store   *app.BlockStore
	storage *fs.BlockStorage
}

func newClient(store *app.BlockStore, dir string, keySize int) *client {
	c := &client{
		dir:     dir,
		keySize: keySize,
		store:   store,
	}
	c.storage = fs.NewBlockStorage(store, c.blocksDir())
	return c
}

func (c *client) init() error {
	if err := os.MkdirAll(c.blocksDir(), 0700); err != nil {
		return err
	}

	return c.storage.Load()
}

func (c *client) save() error {
	return c.storage.Save()
}

func (c *client) blocksDir() string {
	return filepath.Join(c.dir, "blocks")
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
	errSign := issue.SignBlock(privateKey)
	if errSign != nil {
		return nil, errSign
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
	errSign := send.SignBlock(privateKey)
	if errSign != nil {
		return nil, errSign
	}

	return send, nil
}

func (c *client) open(link string, balance float64) (*tradeblocks.AccountBlock, error) {
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

	// Check if we've already created a receive for the linked send, true if
	if c.alreadyLinked(link) {
		return nil, errors.New("open with the specified send already exists")
	}

	// get the linked send
	send, err := c.getBlock(link)
	if err != nil {
		return nil, err
	}

	// create the Open
	open, err := app.Open(publicKey, send, balance)
	if err != nil {
		return nil, err
	}

	// add the signature
	errSign := open.SignBlock(privateKey)
	if errSign != nil {
		return nil, errSign
	}

	return open, nil
}

func (c *client) receive(link string, amount float64) (*tradeblocks.AccountBlock, error) {
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

	// Check if we've already created a receive for the linked send, true if
	if c.alreadyLinked(link) {
		return nil, errors.New("receive with the specified send already exists")
	}

	// get the linked send
	send, err := c.getBlock(link)
	if err != nil {
		return nil, err
	}

	// get the previous block on this chain
	previous, err := c.getHeadBlock(publicKey, send.Token)
	if err != nil {
		return nil, err
	}

	// create the receive
	receive, err := app.Receive(publicKey, previous, send, amount)
	if err != nil {
		return nil, err
	}

	// add the signature
	errSign := receive.SignBlock(privateKey)
	if errSign != nil {
		return nil, errSign
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

// in that case, you can add a param for your validator factory that receives the BlockStore
// move it to the validator
func (c *client) alreadyLinked(hash string) bool {
	// Todo, get the blockstore from the client
	fmt.Printf("the blockstore is %v \n", c.store)
	block, err := c.store.GetBlock(hash)
	if err != nil || block == nil {
		return true
	}
	return false
}
