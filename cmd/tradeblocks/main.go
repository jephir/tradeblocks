package main

import (
	"flag"
	"os"
)

var n = flag.String("node", "localhost:8080", "node address to connect to")

func main() {
	c := &cli{
		keySize:   4096,
		serverURL: "http://" + *n,
		dataDir:   ".",
		out:       os.Stdout,
	}
	if err := c.dispatch(os.Args); err != nil {
		panic(err)
	}
}
