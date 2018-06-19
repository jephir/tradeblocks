package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
	"github.com/jephir/tradeblocks/web"
)

var verifyLocalSigning = flag.Bool("verifylocalsigning", true, "verify signing of local blocks before sending to node")

func init() {
	flag.Parse()
}

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
	// create the Issue block
	account, err := c.getUserAccount()
	if err != nil {
		return nil, err
	}

	issue, err := c.signAccount(tradeblocks.NewIssueBlock(account, balance))
	if err != nil {
		return nil, err
	}

	if err := c.postAccountBlock(issue); err != nil {
		return nil, err
	}

	return issue, nil
}

func (c *client) send(to string, token string, amount float64) (*tradeblocks.AccountBlock, error) {
	// get the keys
	account, err := c.getUserAccount()
	if err != nil {
		return nil, err
	}

	previous, err := c.getHeadBlock(account, token)
	if err != nil {
		return nil, err
	}

	// create the send block
	send, err := c.signAccount(tradeblocks.NewSendBlock(previous, to, amount))
	if err != nil {
		return nil, fmt.Errorf("client: error creating send: %s", err.Error())
	}

	if err := c.postAccountBlock(send); err != nil {
		return nil, err
	}

	return send, nil
}

func (c *client) openFromSend(link string) (*tradeblocks.AccountBlock, error) {
	account, err := c.getUserAccount()
	if err != nil {
		return nil, err
	}

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
	open, err := c.signAccount(tradeblocks.NewOpenBlockFromSend(account, send, balance))
	if err != nil {
		return nil, err
	}

	if err := c.postAccountBlock(open); err != nil {
		return nil, err
	}

	return open, nil
}

func (c *client) openFromSwap(link string) (*tradeblocks.AccountBlock, error) {
	account, err := c.getUserAccount()
	if err != nil {
		return nil, err
	}

	// get the linked send
	swap, err := c.getSwapBlock(link)
	if err != nil {
		return nil, err
	}

	var send *tradeblocks.AccountBlock
	if account == swap.Account {
		send, err = c.getBlock(swap.Right)
		if err != nil {
			return nil, err
		}
	} else {
		send, err = c.getBlock(swap.Right)
		if err != nil {
			return nil, err
		}
	}

	sendParent, err := c.getBlock(send.Previous)
	if err != nil {
		return nil, err
	}
	balance := sendParent.Balance - send.Balance
	token := send.Token

	// create the Open
	open, err := c.signAccount(tradeblocks.NewOpenBlockFromSwap(account, token, swap, balance))
	if err != nil {
		return nil, err
	}

	if err := c.postAccountBlock(open); err != nil {
		return nil, err
	}

	return open, nil
}

func (c *client) receive(link string) (*tradeblocks.AccountBlock, error) {
	account, err := c.getUserAccount()
	if err != nil {
		return nil, err
	}

	// get the linked send
	send, err := c.getBlock(link)
	if err != nil {
		return nil, err
	}
	sendParent, err := c.getBlock(send.Previous)
	if err != nil {
		return nil, err
	}
	amount := sendParent.Balance - send.Balance

	// get the previous block on this chain
	previous, err := c.getHeadBlock(account, send.Token)
	if err != nil {
		return nil, err
	}

	// create the receive
	receive, err := c.signAccount(tradeblocks.NewReceiveBlockFromSend(previous, send, amount))
	if err != nil {
		return nil, err
	}

	if err := c.postAccountBlock(receive); err != nil {
		return nil, err
	}

	return receive, nil
}

func (c *client) offer(left, ID, counterparty, want string, quantity float64, executor string, fee float64) (*tradeblocks.SwapBlock, error) {
	account, err := c.getUserAccount()
	if err != nil {
		return nil, err
	}

	// get the linked send
	send, err := c.getBlock(left)
	if err != nil {
		return nil, err
	}

	// create the offer
	offer, err := c.signSwap(tradeblocks.NewOfferBlock(account, send, ID, counterparty, want, quantity, executor, fee))
	if err != nil {
		return nil, err
	}

	if err := c.postSwapBlock(offer); err != nil {
		return nil, err
	}

	return offer, nil
}

func (c *client) commit(offer string, send string) (*tradeblocks.SwapBlock, error) {
	// get the linked send
	right, err := c.getBlock(send)
	if err != nil {
		return nil, err
	}

	// get the original offer block
	offerBlock, err := c.getSwapBlock(offer)
	if err != nil {
		return nil, err
	}

	// create the commit
	commit, err := c.signSwap(tradeblocks.NewCommitBlock(offerBlock, right))
	if err := c.postSwapBlock(commit); err != nil {
		return nil, err
	}

	return commit, nil
}

func (c *client) refundLeft(offer string) (*tradeblocks.SwapBlock, error) {
	// get the original offer block
	offerBlock, err := c.getSwapBlock(offer)
	if err != nil {
		return nil, err
	}

	// get the original send
	left, err := c.getBlock(offerBlock.Left)
	if err != nil {
		return nil, err
	}

	// the address to refund to is the sends original address
	refundTo := left.Account

	// create the refund
	refundLeft, err := c.signSwap(tradeblocks.NewRefundLeftBlock(offerBlock, refundTo))
	if err != nil {
		return nil, err
	}

	if err := c.postSwapBlock(refundLeft); err != nil {
		return nil, err
	}

	return refundLeft, nil
}

func (c *client) refundRight(refundLeft string) (*tradeblocks.SwapBlock, error) {
	// get the original offer block
	refundLeftBlock, err := c.getSwapBlock(refundLeft)
	if err != nil {
		return nil, err
	}

	// get the counterparty send
	right, err := c.getBlock(refundLeftBlock.Right)
	if err != nil {
		return nil, err
	}

	// the address to refund to is the sends original address
	refundTo := right.Account

	// create the refund
	refundRight, err := c.signSwap(tradeblocks.NewRefundRightBlock(refundLeftBlock, right, refundTo))
	if err != nil {
		return nil, err
	}

	if err := c.postSwapBlock(refundRight); err != nil {
		return nil, err
	}

	return refundRight, nil
}

func (c *client) createOrder(send string, ID string, partial bool, quote string, price float64, executor string, fee float64) (*tradeblocks.OrderBlock, error) {
	account, err := c.getUserAccount()
	if err != nil {
		return nil, err
	}

	// get the originating send
	sendBlock, err := c.getBlock(send)
	if err != nil {
		return nil, err
	}

	// get the previous of the send
	sendPrevBlock, err := c.getBlock(sendBlock.Previous)
	if err != nil {
		return nil, err
	}

	// balance of the order
	balance := sendPrevBlock.Balance - sendBlock.Balance

	// create the refund
	createOrderBlock, err := c.signOrder(tradeblocks.NewCreateOrderBlock(account, sendBlock, balance, ID, partial, quote, price, executor, fee))
	if err != nil {
		return nil, err
	}

	if err := c.postOrderBlock(createOrderBlock); err != nil {
		return nil, err
	}

	return createOrderBlock, nil
}

func (c *client) acceptOrder(swap string, link string) (*tradeblocks.OrderBlock, error) {
	account, err := c.getUserAccount()
	if err != nil {
		return nil, err
	}
	// get the previous order
	prevBlock, err := c.getHeadOrderBlock(account, swap)
	if err != nil {
		return nil, err
	}

	// get the swap by address
	swapBlock, err := c.getSwapBlock(swap)
	if err != nil {
		return nil, err
	}

	// balance of the order
	balance := prevBlock.Balance - swapBlock.Quantity

	// create the refund
	acceptOrderBlock, err := c.signOrder(tradeblocks.NewAcceptOrderBlock(prevBlock, link, balance))
	if err != nil {
		return nil, err
	}

	if err := c.postOrderBlock(acceptOrderBlock); err != nil {
		return nil, err
	}

	return acceptOrderBlock, nil
}

func (c *client) refundOrder(order string) (*tradeblocks.OrderBlock, error) {
	account, err := c.getUserAccount()
	if err != nil {
		return nil, err
	}

	// get the previous order
	prevBlock, err := c.getHeadOrderBlock(account, order)
	if err != nil {
		return nil, err
	}

	refundTo := prevBlock.Account

	// create the refund
	refundOrderBlock, err := c.signOrder(tradeblocks.NewRefundOrderBlock(prevBlock, refundTo))
	if err != nil {
		return nil, err
	}

	if err := c.postOrderBlock(refundOrderBlock); err != nil {
		return nil, err
	}

	return refundOrderBlock, nil
}

func (c *client) sell(quantity float64, base string, ppu float64, quote string) (tradeblocks.Block, error) {
	account, err := c.getUserAccount()
	if err != nil {
		return nil, err
	}

	// TODO match against buy order if found

	id := app.UniqueID()
	link := tradeblocks.OrderAddress(account, id)

	send, err := c.send(link, base, quantity)
	if err != nil {
		return nil, fmt.Errorf("client: error creating send for sell: %s", err.Error())
	}

	r, err := c.api.NewGetAddressRequest()
	if err != nil {
		return nil, err
	}

	res, err := c.http.Do(r)
	if err != nil {
		return nil, err
	}

	executor, err := c.api.DecodeGetAddressResponse(res)
	if err != nil {
		return nil, err
	}

	return c.createOrder(send.Hash(), id, false, quote, ppu, executor, 0)
}

func (c *client) buy(quantity float64, base string, ppu float64, quote string) ([]tradeblocks.Block, error) {
	account, err := c.getUserAccount()
	if err != nil {
		return nil, err
	}

	r, err := c.api.NewGetBuyOrdersRequest(base, ppu, quote)
	if err != nil {
		return nil, err
	}

	res, err := c.http.Do(r)
	if err != nil {
		return nil, err
	}

	orders, err := c.api.DecodeGetOrdersArrayResponse(res)
	if err != nil {
		return nil, err
	}

	var swaps []tradeblocks.Block
	for quantity != 0 {
		// Get order
		if len(orders) == 0 {
			return nil, fmt.Errorf("client: not enough orders to fill buy")
		}
		var order *tradeblocks.OrderBlock
		order, orders = orders[0], orders[1:]

		receiveQuantity := math.Min(quantity, order.Balance)
		sendQuantity := receiveQuantity * ppu

		send, err := c.send(tradeblocks.SwapAddress(account, order.ID), quote, sendQuantity)
		if err != nil {
			return nil, err
		}

		swap, err := c.offer(send.Hash(), order.ID, order.Account, base, receiveQuantity, order.Executor, order.Fee)
		if err != nil {
			return nil, err
		}
		swaps = append(swaps, swap)
		quantity -= receiveQuantity
	}

	return swaps, nil
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

func (c *client) getHeadBlock(address, token string) (*tradeblocks.AccountBlock, error) {
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

func (c *client) getHeadSwapBlock(address, id string) (*tradeblocks.SwapBlock, error) {
	r, err := c.api.NewGetSwapHeadRequest(address, id)
	if err != nil {
		return nil, err
	}
	res, err := c.http.Do(r)
	if err != nil {
		return nil, err
	}
	var result tradeblocks.SwapBlock
	if err := c.api.DecodeSwapBlockResponse(res, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) getHeadOrderBlock(address, id string) (*tradeblocks.OrderBlock, error) {
	r, err := c.api.NewGetOrderHeadRequest(address, id)
	if err != nil {
		return nil, err
	}
	res, err := c.http.Do(r)
	if err != nil {
		return nil, err
	}
	var result tradeblocks.OrderBlock
	if err := c.api.DecodeOrderBlockResponse(res, &result); err != nil {
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

func (c *client) getSwapBlock(hash string) (*tradeblocks.SwapBlock, error) {
	r, err := c.api.NewGetBlockRequest(hash)
	if err != nil {
		return nil, err
	}
	res, err := c.http.Do(r)
	if err != nil {
		return nil, err
	}
	var result tradeblocks.SwapBlock
	if err := c.api.DecodeSwapBlockResponse(res, &result); err != nil {
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

func (c *client) postSwapBlock(b *tradeblocks.SwapBlock) error {
	req, err := c.api.NewPostSwapBlockRequest(b)
	if err != nil {
		return err
	}
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	var rb tradeblocks.SwapBlock
	if err := c.api.DecodeSwapBlockResponse(res, &rb); err != nil {
		return err
	}
	return nil
}

func (c *client) postOrderBlock(b *tradeblocks.OrderBlock) error {
	req, err := c.api.NewPostOrderBlockRequest(b)
	if err != nil {
		return err
	}
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	var rb tradeblocks.OrderBlock
	if err := c.api.DecodeOrderBlockResponse(res, &rb); err != nil {
		return err
	}
	return nil
}

func (c *client) getUserAccount() (string, error) {
	privateKey, err := c.openPrivateKey()
	if err != nil {
		return "", err
	}
	defer privateKey.Close()
	priv, err := parsePrivateKey(privateKey)
	if err != nil {
		return "", err
	}
	addr, err := app.PrivateKeyToAddress(priv)
	if err != nil {
		return "", err
	}
	return addr, nil
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

func (c *client) signAccount(b *tradeblocks.AccountBlock) (*tradeblocks.AccountBlock, error) {
	priv, err := c.getPrivateKey()
	if err != nil {
		return nil, err
	}
	b.Normalize()
	if err := b.SignBlock(priv); err != nil {
		return nil, err
	}
	if *verifyLocalSigning {
		pub, err := app.AddressToRSAKey(b.Account)
		if err != nil {
			return nil, err
		}
		if err := b.VerifyBlock(pub); err != nil {
			return nil, fmt.Errorf("client: verification error: %s", err.Error())
		}
	}
	return b, nil
}

func (c *client) signSwap(b *tradeblocks.SwapBlock) (*tradeblocks.SwapBlock, error) {
	priv, err := c.getPrivateKey()
	if err != nil {
		return nil, err
	}
	b.Normalize()
	if err := b.SignBlock(priv); err != nil {
		return nil, err
	}
	if *verifyLocalSigning {
		pub, err := app.AddressToRSAKey(b.Account)
		if err != nil {
			return nil, err
		}
		if err := b.VerifyBlock(pub); err != nil {
			return nil, fmt.Errorf("client: verification error: %s", err.Error())
		}
	}
	return b, nil
}

func (c *client) signOrder(b *tradeblocks.OrderBlock) (*tradeblocks.OrderBlock, error) {
	priv, err := c.getPrivateKey()
	if err != nil {
		return nil, err
	}
	b.Normalize()
	if err := b.SignBlock(priv); err != nil {
		return nil, err
	}
	if *verifyLocalSigning {
		pub, err := app.AddressToRSAKey(b.Account)
		if err != nil {
			return nil, err
		}
		if err := b.VerifyBlock(pub); err != nil {
			return nil, fmt.Errorf("client: verification error: %s", err.Error())
		}
	}
	return b, nil
}

func (c *client) getPrivateKey() (*rsa.PrivateKey, error) {
	privateKey, err := c.openPrivateKey()
	if err != nil {
		return nil, err
	}
	defer privateKey.Close()
	return parsePrivateKey(privateKey)
}
