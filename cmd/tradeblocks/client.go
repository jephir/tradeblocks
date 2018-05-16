package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/jephir/tradeblocks"
)

type client struct {
	base string
}

func (c *client) SendAccountBlock(b *tradeblocks.AccountBlock) (result *tradeblocks.AccountBlock, err error) {
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
	res, err := http.Post(u.String(), "application/json", &buf)
	if err != nil {
		return
	}
	err = json.NewDecoder(res.Body).Decode(result)
	if err != nil {
		return
	}
	return
}
