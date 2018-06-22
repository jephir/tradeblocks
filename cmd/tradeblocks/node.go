package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/jephir/tradeblocks/node"
)

var addr = flag.String("listen", "localhost:8080", "listen address")
var bootstrap = flag.String("bootstrap", "", "bootstrap node URL")
var dir = flag.String("dir", ".", "database directory")

func init() {
	flag.Parse()
}

func (cli *cli) handleNode() error {
	n, err := node.NewNode(*dir)
	if err != nil {
		return err
	}
	if *addr != "" && *bootstrap != "" {
		if err := n.Bootstrap(*addr, *bootstrap); err != nil {
			return err
		}
	}
	if *addr != "" {
		fmt.Fprintln(cli.out, *addr)
		return http.ListenAndServe(*addr, n)
	}
	return nil
}
