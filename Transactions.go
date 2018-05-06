package dexathon

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
)

// BaseTransaction is the format for basic transactions options:
// ['open', 'issue', 'send', 'receive', 'change']
type BaseTransaction struct {
	action         string
	account        string
	token          string
	previous       string
	representative string
	balance        float64
	link           string
}

// SellTransaction to create a sell order
type SellTransaction struct {
	action   string
	account  string
	token    string
	id       string
	previous string
	balance  float64
	quote    string
	price    float64
	link     string
	partial  bool
}

// SwapTransaction to initiate a swap of two tokens
type SwapTransaction struct {
	action      string
	account     string
	id          string
	previous    string
	left        string
	right       string
	counterplay string
	want        string
	quantity    float64
}

// NewChange creates a new openTransaction with the given sendTx
func NewChange(representative string) (newTransaction BaseTransaction) {
	var account = LoadAccount()
	var balance = LoadBalance()
	var link = "" // purposeful empty (nil) value: see docs
	var previous = LoadPrevious()
	var token = LoadToken()

	newTransaction = BaseTransaction{
		action:         "change",
		account:        account,
		token:          token,
		previous:       previous,
		representative: representative,
		balance:        balance,
		link:           link,
	}
	fmt.Printf("this is the open transaction: \n")
	spew.Dump(newTransaction)
	return newTransaction
}

// NewIssue creates a new openTransaction with the given sendTx
func NewIssue(balance float64) (newTransaction BaseTransaction) {
	var account = LoadAccount()
	var link = ""     // purposeful empty (nil) value: see docs
	var previous = "" // purposeful empty (nil) value: see docs
	var representative = LoadRepresentative()
	var token = LoadToken()

	newTransaction = BaseTransaction{
		action:         "issue",
		account:        account,
		token:          token,
		previous:       previous,
		representative: representative,
		balance:        balance,
		link:           link,
	}
	fmt.Printf("this is the open transaction: \n")
	spew.Dump(newTransaction)
	return newTransaction
}

// NewOpen creates a new openTransaction with the given sendTx
func NewOpen(link string) (newTransaction BaseTransaction) {
	var account = LoadAccount()
	var balance = LoadBalance() // load from the tx?
	var previous = ""           // purposeful empty (nil) value: see docs
	var representative = LoadRepresentative()
	var token = LoadToken()

	newTransaction = BaseTransaction{
		action:         "open",
		account:        account,
		token:          token,
		previous:       previous,
		representative: representative,
		balance:        balance,
		link:           link,
	}
	fmt.Printf("this is the open transaction: \n")
	spew.Dump(newTransaction)
	return newTransaction
}

// NewReceive creates a new openTransaction with the given sendTx
func NewReceive(link string) (newTransaction BaseTransaction) {
	var account = LoadAccount()
	var balance = LoadBalance()
	var previous = LoadPrevious()
	var representative = LoadRepresentative()
	var token = LoadToken()

	newTransaction = BaseTransaction{
		action:         "receive",
		account:        account,
		token:          token,
		previous:       previous,
		representative: representative,
		balance:        balance,
		link:           link,
	}
	fmt.Printf("this is the open transaction: \n")
	spew.Dump(newTransaction)
	return newTransaction
}

// NewSend creates a new openTransaction with the given sendTx
func NewSend(toAccount, token string, amount float64) (newTransaction BaseTransaction) {
	var account = LoadAccount()
	var curBalance = LoadBalance()
	var previous = LoadPrevious()
	var representative = LoadRepresentative()

	var newBalance = curBalance - amount

	newTransaction = BaseTransaction{
		action:         "send",
		account:        account,
		token:          token,
		previous:       previous,
		representative: representative,
		balance:        newBalance,
		link:           toAccount,
	}
	fmt.Printf("this is the open transaction: \n")
	spew.Dump(newTransaction)
	return newTransaction
}
