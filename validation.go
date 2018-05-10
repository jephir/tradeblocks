package tradeblocks

import (
	"os"
	"strconv"
)

func RegisterInputValidation() (goodInputs bool, addInfo string) {
	goodInputs = len(os.Args) == 3
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks register <name>"
	return
}

func LoginInputValidation() (goodInputs bool, addInfo string) {
	goodInputs = len(os.Args) == 3
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks login <name>"
	return
}

func IssueInputValidation() (goodInputs bool, addInfo string) {
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks issue <balance>"
	goodInputs = false
	if len(os.Args) == 3 {
		goodInputs = true
		if _, err := strconv.ParseFloat(os.Args[2], 64); err != nil {
			goodInputs = false
			addInfo = "CLI args invalid type.\n" +
				"Run this command with $ tradeblocks issue <balance: integer>"
		}
	}
	return
}

func SendInputValidation() (goodInputs bool, addInfo string) {
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks send <to_account> <token> <amount>"
	goodInputs = false
	if len(os.Args) == 5 {
		goodInputs = true
		if _, err := strconv.ParseFloat(os.Args[4], 64); err != nil {
			goodInputs = false
			addInfo = "CLI args invalid type.\n" +
				"Run this command with $ tradeblocks send" +
				"<to_account: string> <token: string> <amount: int> \n"
		}
	}
	return
}

func OpenInputValidation() (goodInputs bool, addInfo string) {
	goodInputs = len(os.Args) == 3
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks open <send_tx>"
	return
}

func ReceiveInputValidation() (goodInputs bool, addInfo string) {
	goodInputs = len(os.Args) == 3
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks login <send_tx>"
	return
}
