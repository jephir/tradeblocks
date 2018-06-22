package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/jephir/tradeblocks/node"
)

var addr = flag.String("listen", "localhost:8080", "listen address")
var bootstrap = flag.String("bootstrap", "", "bootstrap node URL")

func init() {
	flag.Parse()
}

func (cli *cli) handleNode() error {
	n, err := node.NewNode("data/blocks")
	if err != nil {
		return err
	}
	if *bootstrap != "" {
		if err := n.Bootstrap(*addr, *bootstrap); err != nil {
			return err
		}
	}
	fmt.Fprintln(cli.out, *addr)
	if err := http.ListenAndServe(*addr, n); err != nil {
		return err
	}
	return nil
}
