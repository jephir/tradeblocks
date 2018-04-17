package main

// AccountBlock represents a block in the account blockchain
type AccountBlock struct {
	Action         string
	Account        string
	Token          string
	Previous       string
	Representative string
	Balance        float64
	Link           string
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
