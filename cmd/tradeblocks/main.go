package main

import (
	"fmt"
	"github.com/jephir/tradeblocks/app"
	"net/http/httptest"
	"os"
	"strconv"

	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/web"
)

const keySize = 4096
const serverURL = "http://localhost:8080"
const dataDir = "data"

func main() {
	var command = os.Args[1]
	var block *tradeblocks.AccountBlock
	var err error

	store := app.NewBlockStore()
	srv := web.NewServer(store)
	c := web.NewClient(serverURL)
	cmd := newClient(store, dataDir, keySize)

	if err := cmd.init(); err != nil {
		panic(err)
	}

	switch command {
	case "register":
		goodInputs, addInfo := registerInputValidation()
		if goodInputs {
			address, err := cmd.register(os.Args[2])
			if err != nil {
				panic(err)
			}
			fmt.Println(address)
		} else {
			cmd.badInputs("register", addInfo)
		}
	case "login":
		goodInputs, addInfo := loginInputValidation()
		if goodInputs {
			address, err := cmd.login(os.Args[2])
			if err != nil {
				panic(err)
			}
			fmt.Println(address)
		} else {
			cmd.badInputs("login", addInfo)
		}
	case "issue":
		goodInputs, addInfo := issueInputValidation()
		if goodInputs {
			balance, _ := strconv.ParseFloat(os.Args[2], 64)
			block, err = cmd.issue(balance)
			if err != nil {
				panic(err)
			}
		} else {
			cmd.badInputs("issue", addInfo)
		}
	case "send":
		goodInputs, addInfo := sendInputValidation()
		if goodInputs {
			amount, _ := strconv.ParseFloat(os.Args[4], 64)
			block, err = cmd.send(os.Args[2], os.Args[3], amount)
			if err != nil {
				panic(err)
			}
		} else {
			cmd.badInputs("send", addInfo)
		}
	case "open":
		goodInputs, addInfo := openInputValidation()
		if goodInputs {
			block, err = cmd.open(os.Args[2])
			if err != nil {
				panic(err)
			}
		} else {
			cmd.badInputs("open", addInfo)
		}
	case "receive":
		goodInputs, addInfo := receiveInputValidation()
		if goodInputs {
			block, err = cmd.receive(os.Args[2])
			if err != nil {
				panic(err)
			}
		} else {
			cmd.badInputs("receive", addInfo)
		}
	case "trade":
		// TODO Implement trading
		fmt.Println("TWJOTBNV7AKQQNND2G6HZRZM4AD2ZNBQOZPF7UTRS6DBBKJ5ZILA")
	}

	if block != nil {
		req, err := c.NewPostAccountRequest(block)
		if err != nil {
			panic(err)
		}
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		res := w.Result()
		result, err := c.DecodeAccountResponse(res)
		if err != nil {
			panic(err)
		}
		fmt.Println(result.Hash)
	}

	if err := cmd.save(); err != nil {
		panic(err)
	}
}
