package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/jephir/tradeblocks/node"
)

const addr = "localhost:8080"

func main() {
	var command = os.Args[1]
	if command == "node" {
		n, err := node.NewNode("data/blocks")
		if err != nil {
			panic(err)
		}
		fmt.Printf("tradeblocks: listening on '%s'\n", addr)
		if err := http.ListenAndServe(addr, n); err != nil {
			panic(err)
		}
	} else {
		c := &cli{
			keySize:   4096,
			serverURL: "http://" + addr,
			dataDir:   "data",
			out:       os.Stdout,
		}
		if err := c.dispatch(os.Args); err != nil {
			panic(err)
		}
	}
}
