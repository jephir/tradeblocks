package web

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	err = json.NewEncoder(&buf).Encode(*b)
	if err != nil {
		return
	}
	u, err := url.Parse(c.base)
	if err != nil {
		return
	}
	u.Path = "/account"
	r, err = http.NewRequest("POST", u.String(), &buf)
	if err != nil {
		return
	}
	r.Header.Set("Content-Type", "application/json")
	return
}

// DecodeResponse returns the result of a server request
func (c *Client) DecodeResponse(res *http.Response) (result *tradeblocks.AccountBlock, err error) {
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpected status code %d", res.StatusCode)
		return
	}
	var block tradeblocks.AccountBlock
	err = json.NewDecoder(res.Body).Decode(&block)
	result = &block
	return
}
