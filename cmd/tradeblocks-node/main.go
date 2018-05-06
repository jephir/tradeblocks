package main

import (
	"log"
	"net/http"

	"github.com/jephir/dexathon"
)

func main() {
	http.Handle("/account", dexathon.NewAccountServer())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
