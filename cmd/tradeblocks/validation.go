package main

import (
	"strconv"
)

func registerInputValidation(args []string) (goodInputs bool, addInfo string) {
	goodInputs = len(args) == 3
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks register <name: string>"
	return
}

func loginInputValidation(args []string) (goodInputs bool, addInfo string) {
	goodInputs = len(args) == 3
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks login <name: string>"
	return
}

func issueInputValidation(args []string) (goodInputs bool, addInfo string) {
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks issue <balance: float64>"
	goodInputs = false
	if len(args) == 3 {
		goodInputs = true
		if _, err := strconv.ParseFloat(args[2], 64); err != nil {
			goodInputs = false
			addInfo = "CLI args invalid type.\n" +
				"Run this command with $ tradeblocks issue <balance: float>"
		}
	}
	return
}

func sendInputValidation(args []string) (goodInputs bool, addInfo string) {
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks send <to_account> <token> <amount>"
	goodInputs = false
	if len(args) == 5 {
		goodInputs = true
		if _, err := strconv.ParseFloat(args[4], 64); err != nil {
			goodInputs = false
			addInfo = "CLI args invalid type.\n" +
				"Run this command with $ tradeblocks send" +
				"<to_account: string> <token: string> <amount: float> \n"
		}
	}
	return
}

func openFromSendInputValidation(args []string) (goodInputs bool, addInfo string) {
	goodInputs = len(args) == 3
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks open-from-send <send_tx: string>"
	return
}

func openFromSwapInputValidation(args []string) (goodInputs bool, addInfo string) {
	goodInputs = len(args) == 3
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks open-from-swap <swap_tx: string>"
	return
}

func receiveInputValidation(args []string) (goodInputs bool, addInfo string) {
	goodInputs = len(args) == 3
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks receive <send_tx: string>"
	return
}

func offerInputValidation(args []string) (goodInputs bool, addInfo string) {
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks offer. Executor and Fee must be present or absent together\n" +
		"<left: string> <ID: string> <counterparty: string> <want: string> <quantity: float> \n" +
		"<executor: string, OPTIONAL> <fee: float, OPTIONAL> \n"
	goodInputs = false
	if len(args) == 7 {
		goodInputs = true
		if _, err := strconv.ParseFloat(args[6], 64); err != nil {
			goodInputs = false
			addInfo = "CLI args invalid type for quantity, not including optional executor/fee.\n" +
				"Run this command with $ tradeblocks offer. Executor and Fee must be present or absent together\n" +
				"<left: string> <ID: string> <counterparty: string> <want: string> <quantity: float> \n" +
				"<executor: string, OPTIONAL> <fee: float, OPTIONAL> \n"
		}
	}
	if len(args) == 9 {
		goodInputs = true
		if _, err := strconv.ParseFloat(args[6], 64); err != nil {
			goodInputs = false
			addInfo = "CLI args invalid type for quantity, including optional executor/fee.\n" +
				"Run this command with $ tradeblocks offer. Executor and Fee must be present or absent together\n" +
				"<left: string> <ID: string> <counterparty: string> <want: string> <quantity: float> \n" +
				"<executor: string, OPTIONAL> <fee: float, OPTIONAL> \n"
		}
		if _, err := strconv.ParseFloat(args[8], 64); err != nil {
			goodInputs = false
			addInfo = "CLI args invalid type for fee, including optional executor/fee.\n" +
				"Run this command with $ tradeblocks offer. Executor and Fee must be present or absent together\n" +
				"<left: string> <ID: string> <counterparty: string> <want: string> <quantity: float> \n" +
				"<executor: string, OPTIONAL> <fee: float, OPTIONAL> \n"
		}
	}
	return
}

func commitInputValidation(args []string) (goodInputs bool, addInfo string) {
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks commit <offer: string> <right: string>\n"
	goodInputs = len(args) == 4
	return
}

func refundLeftInputValidation(args []string) (goodInputs bool, addInfo string) {
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks refund-left <offer: string>\n"
	goodInputs = len(args) == 3
	return
}

func refundRightInputValidation(args []string) (goodInputs bool, addInfo string) {
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks refund-right <refund-left: string>\n"
	goodInputs = len(args) == 3
	return
}

func createOrderInputValidation(args []string) (goodInputs bool, addInfo string) {
	addInfo = "CLI args invalid length.\n" +
		"Executor and Fee must be present or absent together. Run this command with $ tradeblocks create-order\n" +
		"<send: string> <ID: string> <partial: boolean> <quote: string> <price: float> \n" +
		"<executor: string, OPTIONAL> <fee: float, OPTIONAL>  \n"
	goodInputs = true
	if len(args) == 7 {
		if _, err := strconv.ParseBool(args[4]); err != nil {
			goodInputs = false
			addInfo += "Invalid partial, must be a boolean\n"
		}
		if _, err := strconv.ParseFloat(args[6], 64); err != nil {
			goodInputs = false
			addInfo += "Invalid price, must be a float64\n"
		}
	} else if len(args) == 9 {
		if _, err := strconv.ParseBool(args[4]); err != nil {
			goodInputs = false
			addInfo += "Invalid partial, must be a boolean\n"
		}
		if _, err := strconv.ParseFloat(args[6], 64); err != nil {
			goodInputs = false
			addInfo += "Invalid quantity, must be a float64\n"
		}
		if _, err := strconv.ParseFloat(args[8], 64); err != nil {
			goodInputs = false
			addInfo += "Invalid quantity, must be a float64\n"
		}
	} else {
		goodInputs = false
	}
	return
}

func acceptOrderInputValidation(args []string) (goodInputs bool, addInfo string) {
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks accept-order <order: string> <link: string>\n"
	goodInputs = len(args) == 4
	return
}

func refundOrderInputValidation(args []string) (goodInputs bool, addInfo string) {
	addInfo = "CLI args invalid length.\n" +
		"Run this command with $ tradeblocks refund-order <order: string>\n"
	goodInputs = len(args) == 3
	return
}
