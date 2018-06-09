package main

import (
	"fmt"
	"io"
	"strconv"

	"github.com/jephir/tradeblocks"
)

type cli struct {
	keySize   int
	serverURL string
	dataDir   string
	out       io.Writer
}

func (cli *cli) dispatch(args []string) error {
	var command = args[1]
	var block *tradeblocks.AccountBlock
	var swapBlock *tradeblocks.SwapBlock
	var orderBlock *tradeblocks.OrderBlock
	var err error

	cmd := newClient(cli.dataDir, cli.serverURL, cli.keySize)
	switch command {
	case "register":
		goodInputs, addInfo := registerInputValidation(args)
		if goodInputs {
			address, err := cmd.register(args[2])
			if err != nil {
				return err
			}
			fmt.Fprintln(cli.out, address)
		} else {
			cmd.badInputs("register", addInfo)
		}
	case "login":
		goodInputs, addInfo := loginInputValidation(args)
		if goodInputs {
			address, err := cmd.login(args[2])
			if err != nil {
				return err
			}
			fmt.Fprintln(cli.out, address)
		} else {
			cmd.badInputs("login", addInfo)
		}
	case "issue":
		goodInputs, addInfo := issueInputValidation(args)
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
		goodInputs, addInfo := sendInputValidation(args)
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
		goodInputs, addInfo := openInputValidation(args)
		if goodInputs {
			block, err = cmd.open(args[2])
			if err != nil {
				return err
			}
		} else {
			cmd.badInputs("open", addInfo)
		}
	case "receive":
		goodInputs, addInfo := receiveInputValidation(args)
		if goodInputs {
			block, err = cmd.receive(args[2])
			if err != nil {
				return err
			}
		} else {
			cmd.badInputs("receive", addInfo)
		}
	case "offer":
		goodInputs, addInfo := offerInputValidation(args)
		if goodInputs {
			quantity, _ := strconv.ParseFloat(args[6], 64)
			if len(args) == 7 {
				swapBlock, err = cmd.offer(args[2], args[3], args[4], args[5], quantity, "", 0.0)
			} else if len(args) == 9 {
				fee, _ := strconv.ParseFloat(args[9], 64)
				swapBlock, err = cmd.offer(args[2], args[3], args[4], args[5], quantity, args[7], fee)
			}
			if err != nil {
				return err
			}
		} else {
			cmd.badInputs("offer", addInfo)
		}
	case "commit":
		goodInputs, addInfo := commitInputValidation(args)
		if goodInputs {
			swapBlock, err = cmd.commit(args[2], args[3])
			if err != nil {
				return err
			}
		} else {
			cmd.badInputs("commit", addInfo)
		}
	case "refund-left":
		goodInputs, addInfo := refundLeftInputValidation(args)
		if goodInputs {
			swapBlock, err = cmd.refundLeft(args[2])
			if err != nil {
				return err
			}
		} else {
			cmd.badInputs("refundLeft", addInfo)
		}
	case "refund-right":
		goodInputs, addInfo := refundRightInputValidation(args)
		if goodInputs {
			swapBlock, err = cmd.refundLeft(args[2])
			if err != nil {
				return err
			}
		} else {
			cmd.badInputs("refundRight", addInfo)
		}
	case "create-order":
		goodInputs, addInfo := createOrderInputValidation(args)
		if goodInputs {
			partial, _ := strconv.ParseBool(args[4])
			quantity, _ := strconv.ParseFloat(args[6], 64)
			if len(args) == 7 {
				orderBlock, err = cmd.createOrder(args[2], args[3], partial, args[5], quantity, "", 0.0)
			} else if len(args) == 9 {
				fee, _ := strconv.ParseFloat(args[8], 64)
				orderBlock, err = cmd.createOrder(args[2], args[3], partial, args[5], quantity, args[7], fee)
			}
			if err != nil {
				return err
			}
		} else {
			cmd.badInputs("order", addInfo)
		}
	case "accept-order":
		goodInputs, addInfo := acceptOrderInputValidation(args)
		if goodInputs {
			orderBlock, err = cmd.acceptOrder(args[2], args[3])
			if err != nil {
				return err
			}
		} else {
			cmd.badInputs("order", addInfo)
		}
	case "refund-order":
		goodInputs, addInfo := refundOrderInputValidation(args)
		if goodInputs {
			orderBlock, err = cmd.refundOrder(args[2])
			if err != nil {
				return err
			}
		} else {
			cmd.badInputs("order", addInfo)
		}
	case "trade":
		// TODO Implement trading
		fmt.Fprintln(cli.out, "TWJOTBNV7AKQQNND2G6HZRZM4AD2ZNBQOZPF7UTRS6DBBKJ5ZILA")
	}

	if block != nil {
		fmt.Fprintln(cli.out, block.Hash())
	}

	if swapBlock != nil {
		fmt.Fprintln(cli.out, swapBlock.Hash())
	}

	if orderBlock != nil {
		fmt.Fprintln(cli.out, orderBlock.Hash())
	}

	return nil
}
