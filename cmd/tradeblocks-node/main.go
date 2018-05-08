package main

import (
	"log"
	"net/http"

	"github.com/jephir/tradeblocks"
)

func main() {
	http.Handle("/account", tradeblocks.NewAccountServer())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
