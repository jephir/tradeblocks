package main

import (
	"os"
	"strconv"
)

func registerInputValidation() (goodInputs bool, addInfo string) {
	goodInputs = len(os.Args) == 3
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks register <name>"
	return
}

func loginInputValidation() (goodInputs bool, addInfo string) {
	goodInputs = len(os.Args) == 3
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks login <name>"
	return
}

func issueInputValidation() (goodInputs bool, addInfo string) {
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks issue <balance>"
	goodInputs = false
	if len(os.Args) == 3 {
		goodInputs = true
		if _, err := strconv.ParseFloat(os.Args[2], 64); err != nil {
			goodInputs = false
			addInfo = "CLI args invalid type.\n" +
				"Run this command with $ tradeblocks issue <balance: float>"
		}
	}
	return
}

func sendInputValidation() (goodInputs bool, addInfo string) {
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks send <to_account> <token> <amount>"
	goodInputs = false
	if len(os.Args) == 5 {
		goodInputs = true
		if _, err := strconv.ParseFloat(os.Args[4], 64); err != nil {
			goodInputs = false
			addInfo = "CLI args invalid type.\n" +
				"Run this command with $ tradeblocks send" +
				"<to_account: string> <token: string> <amount: float> \n"
		}
	}
	return
}

func openInputValidation() (goodInputs bool, addInfo string) {
	goodInputs = len(os.Args) == 3
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks open <send_tx>"
	return
}

func receiveInputValidation() (goodInputs bool, addInfo string) {
	goodInputs = len(os.Args) == 3
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks receive <send_tx>"
	return
}
