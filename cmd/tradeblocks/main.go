package main

import (
	"os"
)

func main() {
	var command = os.Args[1]
	if command == "node" {

	} else {
		c := &cli{
			keySize:   4096,
			serverURL: "http://localhost:8080",
			dataDir:   "data",
			out:       os.Stdout,
		}
		if err := c.dispatch(os.Args); err != nil {
			panic(err)
		}
	}
}
