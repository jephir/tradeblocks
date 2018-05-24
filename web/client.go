package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jephir/tradeblocks/app"
	"io"
	"net/http"
	"net/url"

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

// NewAccountBlockRequest returns an http.Request to send the specified block
func (c *Client) NewAccountBlockRequest(b *tradeblocks.AccountBlock) (r *http.Request, err error) {
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(b)
	if err != nil {
		return
	}
	r, err = c.newRequest("POST", "/account", &buf)
	r.Header.Set("Content-Type", "application/json")
	return
}

// DecodeResponse returns the result of a server request
func (c *Client) DecodeResponse(res *http.Response) (result tradeblocks.AccountBlock, err error) {
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
		return fmt.Errorf("unexpected status code %d", res.StatusCode)
	}
	return nil
}
