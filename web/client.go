package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jephir/tradeblocks"
)

// Client communicates with a TradeBlocks server
type Client struct {
	base string
}

// NewClient allocates and returns a new client with the specified node URL
func NewClient(url string) *Client {
	return &Client{base: url}
}

// NewPostAccountBlockRequest returns an http.Request to send the specified block
func (c *Client) NewPostAccountBlockRequest(b *tradeblocks.AccountBlock) (r *http.Request, err error) {
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(b)
	if err != nil {
		return
	}
	r, err = c.newRequest("POST", "/block", &buf)
	if err != nil {
		return
	}
	q := r.URL.Query()
	q.Add("type", "account")
	r.URL.RawQuery = q.Encode()
	r.Header.Set("Content-Type", "application/json")
	return
}

// NewPostSwapBlockRequest returns an http.Request to send the specified block
func (c *Client) NewPostSwapBlockRequest(b *tradeblocks.SwapBlock) (r *http.Request, err error) {
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(b)
	if err != nil {
		return
	}
	r, err = c.newRequest("POST", "/block", &buf)
	if err != nil {
		return
	}
	q := r.URL.Query()
	q.Add("type", "swap")
	r.URL.RawQuery = q.Encode()
	r.Header.Set("Content-Type", "application/json")
	return
}

// NewPostOrderBlockRequest returns an http.Request to send the specified block
func (c *Client) NewPostOrderBlockRequest(b *tradeblocks.OrderBlock) (r *http.Request, err error) {
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(b)
	if err != nil {
		return
	}
	r, err = c.newRequest("POST", "/block", &buf)
	if err != nil {
		return
	}
	q := r.URL.Query()
	q.Add("type", "order")
	r.URL.RawQuery = q.Encode()
	r.Header.Set("Content-Type", "application/json")
	return
}

// NewGetBlocksRequest returns an http.Request to get all blocks
func (c *Client) NewGetBlocksRequest() (r *http.Request, err error) {
	r, err = c.newRequest("GET", "/blocks", nil)
	return
}

// NewGetAccountBlocksRequest returns an http.Request to get all account blocks
func (c *Client) NewGetAccountBlocksRequest() (r *http.Request, err error) {
	r, err = c.newRequest("GET", "/blocks", nil)
	q := r.URL.Query()
	q.Add("type", "account")
	r.URL.RawQuery = q.Encode()
	return
}

// NewGetSwapBlocksRequest returns an http.Request to get all swap blocks
func (c *Client) NewGetSwapBlocksRequest() (r *http.Request, err error) {
	r, err = c.newRequest("GET", "/blocks", nil)
	q := r.URL.Query()
	q.Add("type", "swap")
	r.URL.RawQuery = q.Encode()
	return
}

// NewGetOrderBlocksRequest returns an http.Request to get all order blocks
func (c *Client) NewGetOrderBlocksRequest() (r *http.Request, err error) {
	r, err = c.newRequest("GET", "/blocks", nil)
	q := r.URL.Query()
	q.Add("type", "order")
	r.URL.RawQuery = q.Encode()
	return
}

// NewGetBlockRequest returns an http.Request to get a block by hash
func (c *Client) NewGetBlockRequest(hash string) (r *http.Request, err error) {
	hash = strings.TrimSpace(hash)
	r, err = c.newRequest("GET", "/block", nil)
	if err != nil {
		return
	}
	q := r.URL.Query()
	q.Add("hash", hash)
	r.URL.RawQuery = q.Encode()
	return
}

// NewGetAccountHeadRequest returns an http.Request to get the head of an account-token chain
func (c *Client) NewGetAccountHeadRequest(account, token string) (r *http.Request, err error) {
	account = strings.TrimSpace(account)
	token = strings.TrimSpace(token)
	r, err = c.newGetHeadRequest("account")
	if err != nil {
		return
	}
	q := r.URL.Query()
	q.Add("account", account)
	q.Add("token", token)
	r.URL.RawQuery = q.Encode()
	return
}

// NewGetSwapHeadRequest returns an http.Request to get the head of an account-id swap chain
func (c *Client) NewGetSwapHeadRequest(account, id string) (r *http.Request, err error) {
	account = strings.TrimSpace(account)
	id = strings.TrimSpace(id)
	r, err = c.newGetHeadRequest("swap")
	if err != nil {
		return
	}
	q := r.URL.Query()
	q.Add("account", account)
	q.Add("id", id)
	r.URL.RawQuery = q.Encode()
	return
}

// NewGetOrderHeadRequest returns an http.Request to get the head of an accound-id order chain
func (c *Client) NewGetOrderHeadRequest(account, id string) (r *http.Request, err error) {
	account = strings.TrimSpace(account)
	id = strings.TrimSpace(id)
	r, err = c.newGetHeadRequest("order")
	if err != nil {
		return
	}
	q := r.URL.Query()
	q.Add("account", account)
	q.Add("id", id)
	r.URL.RawQuery = q.Encode()
	return
}

func (c *Client) newGetHeadRequest(t string) (r *http.Request, err error) {
	r, err = c.newRequest("GET", "/head", nil)
	if err != nil {
		return
	}
	q := r.URL.Query()
	q.Add("type", t)
	r.URL.RawQuery = q.Encode()
	return
}

// NewGetBuyOrdersRequest returns matching buy orders for the specified parameters
func (c *Client) NewGetBuyOrdersRequest(base string, ppu float64, quote string) (r *http.Request, err error) {
	r, err = c.newRequest("GET", "/orders", nil)
	q := r.URL.Query()
	q.Add("side", "buy")
	q.Add("base", base)
	q.Add("ppu", strconv.FormatFloat(ppu, 'E', -1, 64))
	q.Add("quote", quote)
	r.URL.RawQuery = q.Encode()
	return
}

// NewGetSellOrdersRequest returns matching sell orders for the specified parameters
func (c *Client) NewGetSellOrdersRequest(base string, ppu float64, quote string) (r *http.Request, err error) {
	r, err = c.newRequest("GET", "/orders", nil)
	q := r.URL.Query()
	q.Add("side", "sell")
	q.Add("base", base)
	q.Add("ppu", strconv.FormatFloat(ppu, 'E', -1, 64))
	q.Add("quote", quote)
	r.URL.RawQuery = q.Encode()
	return
}

// NewGetAddressRequest returns the address (public key) of the node
func (c *Client) NewGetAddressRequest() (r *http.Request, err error) {
	r, err = c.newRequest("GET", "/address", nil)
	return
}

// DecodeAccountBlockResponse returns the result of an account block request
func (c *Client) DecodeAccountBlockResponse(res *http.Response, result *tradeblocks.AccountBlock) error {
	if err := c.checkResponse(res); err != nil {
		return err
	}
	return json.NewDecoder(res.Body).Decode(&result)
}

// DecodeSwapBlockResponse returns the result of a swap block request
func (c *Client) DecodeSwapBlockResponse(res *http.Response, result *tradeblocks.SwapBlock) error {
	if err := c.checkResponse(res); err != nil {
		return err
	}
	return json.NewDecoder(res.Body).Decode(&result)
}

// DecodeOrderBlockResponse returns the result of an order block request
func (c *Client) DecodeOrderBlockResponse(res *http.Response, result *tradeblocks.OrderBlock) error {
	if err := c.checkResponse(res); err != nil {
		return err
	}
	return json.NewDecoder(res.Body).Decode(&result)
}

// DecodeGetBlocksResponse returns the result of a blocks request
func (c *Client) DecodeGetBlocksResponse(res *http.Response) (result []tradeblocks.Block, err error) {
	err = c.checkResponse(res)
	if err != nil {
		return
	}
	err = json.NewDecoder(res.Body).Decode(&result)
	return
}

// DecodeGetAccountBlocksResponse returns the result of a blocks request
func (c *Client) DecodeGetAccountBlocksResponse(res *http.Response) (result map[string]tradeblocks.NetworkAccountBlock, err error) {
	err = c.checkResponse(res)
	if err != nil {
		return
	}
	err = json.NewDecoder(res.Body).Decode(&result)
	return
}

// DecodeGetSwapBlocksResponse returns the result of a blocks request
func (c *Client) DecodeGetSwapBlocksResponse(res *http.Response) (result map[string]tradeblocks.NetworkSwapBlock, err error) {
	err = c.checkResponse(res)
	if err != nil {
		return
	}
	err = json.NewDecoder(res.Body).Decode(&result)
	return
}

// DecodeGetOrderBlocksResponse returns the result of a blocks request
func (c *Client) DecodeGetOrderBlocksResponse(res *http.Response) (result map[string]tradeblocks.NetworkOrderBlock, err error) {
	err = c.checkResponse(res)
	if err != nil {
		return
	}
	err = json.NewDecoder(res.Body).Decode(&result)
	return
}

// DecodeGetOrdersArrayResponse returns the result of a blocks request
func (c *Client) DecodeGetOrdersArrayResponse(res *http.Response, result []*tradeblocks.OrderBlock) error {
	if err := c.checkResponse(res); err != nil {
		return err
	}
	return json.NewDecoder(res.Body).Decode(&result)
}

// DecodeGetAddressResponse returns the result of a get address request
func (c *Client) DecodeGetAddressResponse(res *http.Response) (string, error) {
	if err := c.checkResponse(res); err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *Client) newRequest(method, path string, body io.Reader) (r *http.Request, err error) {
	u, err := url.Parse(c.base)
	if err != nil {
		return
	}
	u.Path = path
	r, err = http.NewRequest(method, u.String(), body)
	return
}

func (c *Client) checkResponse(res *http.Response) error {
	if res.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("client: unexpected status %d on %s: %s", res.StatusCode, res.Request.URL.String(), string(b))
	}
	return nil
}
