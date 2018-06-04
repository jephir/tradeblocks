package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jephir/tradeblocks/app"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
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

// NewPostAccountRequest returns an http.Request to send the specified block
func (c *Client) NewPostAccountRequest(b *tradeblocks.AccountBlock) (r *http.Request, err error) {
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(b)
	if err != nil {
		return
	}
	r, err = c.newRequest("POST", "/account", &buf)
	if err != nil {
		return
	}
	r.Header.Set("Content-Type", "application/json")
	return
}

// DecodeAccountResponse returns the result of a PostAccount request
func (c *Client) DecodeAccountResponse(res *http.Response) (result tradeblocks.AccountBlock, err error) {
	err = c.checkResponse(res)
	if err != nil {
		return
	}
	err = json.NewDecoder(res.Body).Decode(&result)
	return
}

// NewGetBlocksRequest returns an http.Request to get all blocks
func (c *Client) NewGetBlocksRequest() (r *http.Request, err error) {
	r, err = c.newRequest("GET", "/blocks", nil)
	return
}

// DecodeGetBlocksResponse returns the result of a GetBlocks request
func (c *Client) DecodeGetBlocksResponse(res *http.Response) (result app.AccountBlocksMap, err error) {
	err = c.checkResponse(res)
	if err != nil {
		return
	}
	err = json.NewDecoder(res.Body).Decode(&result)
	return
}

// NewGetAccountBlockRequest returns an http.Request to get an account block by hash
func (c *Client) NewGetAccountBlockRequest(hash string) (r *http.Request, err error) {
	hash = strings.TrimSpace(hash)
	r, err = c.newRequest("GET", "/account", nil)
	if err != nil {
		return
	}
	q := r.URL.Query()
	q.Add("hash", hash)
	r.URL.RawQuery = q.Encode()
	return
}

// DecodeGetAccountBlockResponse returns the result of a GetAccountHead request
func (c *Client) DecodeGetAccountBlockResponse(res *http.Response, result *tradeblocks.AccountBlock) error {
	if err := c.checkResponse(res); err != nil {
		return err
	}
	return json.NewDecoder(res.Body).Decode(result)
}

// NewGetAccountHeadRequest returns an http.Request to get the head of an account-token chain
func (c *Client) NewGetAccountHeadRequest(account, token string) (r *http.Request, err error) {
	account = strings.TrimSpace(account)
	token = strings.TrimSpace(token)
	r, err = c.newRequest("GET", "/head", nil)
	if err != nil {
		return
	}
	q := r.URL.Query()
	q.Add("account", account)
	q.Add("token", token)
	r.URL.RawQuery = q.Encode()
	return
}

// DecodeGetAccountHeadResponse returns the result of a GetAccountHead request
func (c *Client) DecodeGetAccountHeadResponse(res *http.Response, result *tradeblocks.AccountBlock) error {
	if err := c.checkResponse(res); err != nil {
		return err
	}
	return json.NewDecoder(res.Body).Decode(result)
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
		return fmt.Errorf("client: unexpected status code %d: %s", res.StatusCode, string(b))
	}
	return nil
}
