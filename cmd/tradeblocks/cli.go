package main

import (
	"fmt"
	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
	"github.com/jephir/tradeblocks/web"
	"net/http/httptest"
	"strconv"
)

type cli struct {
	keySize   int
	serverURL string
	dataDir   string
}

func (cli *cli) dispatch(args []string) error {
	var command = args[1]
	var block *tradeblocks.AccountBlock
	var err error

	store := app.NewBlockStore()
	srv := web.NewServer(store)
	c := web.NewClient(cli.serverURL)
	cmd := newClient(store, cli.dataDir, cli.keySize)

	if err := cmd.init(); err != nil {
		return err
	}

	switch command {
	case "register":
		goodInputs, addInfo := registerInputValidation()
		if goodInputs {
			address, err := cmd.register(args[2])
			if err != nil {
				return err
			}
			fmt.Println(address)
		} else {
			cmd.badInputs("register", addInfo)
		}
	case "login":
		goodInputs, addInfo := loginInputValidation()
		if goodInputs {
			address, err := cmd.login(args[2])
			if err != nil {
				return err
			}
			fmt.Println(address)
		} else {
			cmd.badInputs("login", addInfo)
		}
	case "issue":
		goodInputs, addInfo := issueInputValidation()
		if goodInputs {
			balance, _ := strconv.ParseFloat(args[2], 64)
			block, err = cmd.issue(balance)
			if err != nil {
				return err
			}
		} else {
			cmd.badInputs("issue", addInfo)
		}
	case "send":
		goodInputs, addInfo := sendInputValidation()
		if goodInputs {
			amount, _ := strconv.ParseFloat(args[4], 64)
			block, err = cmd.send(args[2], args[3], amount)
			if err != nil {
				return err
			}
		} else {
			cmd.badInputs("send", addInfo)
		}
	case "open":
		goodInputs, addInfo := openInputValidation()
		if goodInputs {
			block, err = cmd.open(args[2])
			if err != nil {
				return err
			}
		} else {
			cmd.badInputs("open", addInfo)
		}
	case "receive":
		goodInputs, addInfo := receiveInputValidation()
		if goodInputs {
			block, err = cmd.receive(args[2])
			if err != nil {
				return err
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
			return err
		}
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		res := w.Result()
		result, err := c.DecodeAccountResponse(res)
		if err != nil {
			return err
		}
		fmt.Println(result.Hash)
	}

	return cmd.save()
}
