package tradeblocks

// AccountBlock represents a block in the account blockchain
type AccountBlock struct {
	Action         string
	Account        string
	Token          string
	Previous       string
	Representative string
	Balance        float64
	Link           string

	// Calculated on local machine
	Hash          string
	PreviousBlock *AccountBlock
}

// NewIssueBlock initializes a new crypto token
func NewIssueBlock(account string, balance float64) *AccountBlock {
	return &AccountBlock{
		Action:         "issue",
		Account:        account,
		Token:          account,
		Previous:       "",
		Representative: "",
		Balance:        balance,
		Link:           "",
	}
}

// NewOpenBlock initializes the start of an account blockchain
func NewOpenBlock(account string, send *AccountBlock) *AccountBlock {
	return &AccountBlock{
		Action:         "open",
		Account:        account,
		Token:          send.Token,
		Previous:       "",
		Representative: "",
		Balance:        send.PreviousBlock.Balance - send.Balance,
		Link:           send.Hash,
	}
}

// NewSendBlock initializes a send to the specified address
func NewSendBlock(account string, previous *AccountBlock, to string, amount float64) *AccountBlock {
	return &AccountBlock{
		Action:         "send",
		Account:        account,
		Token:          previous.Token,
		Previous:       previous.Hash,
		Representative: previous.Representative,
		Balance:        previous.Balance - amount,
		Link:           to,
	}
}

// NewReceiveBlock initializes a receive of tokens
func NewReceiveBlock(account string, previous *AccountBlock, send *AccountBlock) *AccountBlock {
	balance := 0.0 // send.PreviousBlock.Balance - send.Balance
	return &AccountBlock{
		Action:         "receive",
		Account:        account,
		Token:          previous.Token,
		Previous:       previous.Hash,
		Representative: previous.Representative,
		Balance:        balance,
		Link:           send.Hash,
	}
}

// SwapBlock represents a block in the swap blockchain
type SwapBlock struct {
	Action       string
	Account      string
	ID           string
	Previous     string
	Left         string
	Right        string
	Counterparty string
	Want         string
	Quantity     float64
}

// OrderBlock represents a block in the order blockchain
type OrderBlock struct {
	Action   string
	Account  string
	Token    string
	ID       string
	Previous string
	Balance  float64
	Quote    string
	Price    float64
	Link     string
	Partial  bool
}
