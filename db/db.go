package db

import (
	"database/sql"
	"fmt"
	"github.com/jephir/tradeblocks"
	_ "github.com/mattn/go-sqlite3" // sqlite driver
)

type scanner interface {
	Scan(dest ...interface{}) error
}

// DB represents a database
type DB struct {
	db *sql.DB
}

// NewDB connects to the specified data source
func NewDB(dataSourceName string) (*DB, error) {
	dataSourceName += "?_foreign_keys=true"
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}
	d := &DB{
		db: db,
	}
	if err := d.init(); err != nil {
		return nil, err
	}
	return d, nil
}

func (m *DB) init() (err error) {
	s := make(map[string]string)
	s["createAccountsTable"] = `CREATE TABLE IF NOT EXISTS accounts(
		action TEXT NOT NULL CHECK (action IN ('open', 'issue', 'send', 'receive')),
		account TEXT NOT NULL CHECK (account LIKE 'xtb:%'),
		token TEXT NOT NULL CHECK (token LIKE 'xtb:%'),
		previous TEXT UNIQUE,
		representative TEXT NOT NULL CHECK (representative LIKE 'xtb:%'),
		balance REAL NOT NULL CHECK (balance >= 0),
		link TEXT UNIQUE,
		signature TEXT NOT NULL UNIQUE,
		hash TEXT NOT NULL UNIQUE,
		FOREIGN KEY (previous) REFERENCES accounts(hash),
		PRIMARY KEY (hash)
		);`
	s["createSwapsTable"] = `CREATE TABLE IF NOT EXISTS swaps(
		action TEXT NOT NULL CHECK (action IN ('offer', 'commit', 'refund-left', 'refund-right')),
		account TEXT NOT NULL CHECK (account LIKE 'xtb:%'),
		token TEXT NOT NULL CHECK (token LIKE 'xtb:%'),
		id TEXT NOT NULL,
		previous TEXT UNIQUE,
		left TEXT UNIQUE NOT NULL,
		right TEXT UNIQUE,
		refund_left TEXT CHECK (refund_left LIKE 'xtb:%'),
		refund_right TEXT CHECK (refund_right LIKE 'xtb:%'),
		counterparty TEXT NOT NULL CHECK (counterparty LIKE 'xtb:%'),
		want TEXT NOT NULL CHECK (want LIKE 'xtb:%'),
		quantity REAL NOT NULL CHECK (quantity >= 0),
		executor TEXT CHECK (executor LIKE 'xtb:%'),
		fee REAL,
		signature TEXT NOT NULL UNIQUE,
		hash TEXT NOT NULL UNIQUE,
		FOREIGN KEY (previous) REFERENCES swaps(hash),
		PRIMARY KEY (hash)
		);`
	s["createOrdersTable"] = `CREATE TABLE IF NOT EXISTS orders(
		action TEXT NOT NULL CHECK (action IN ('create-order', 'accept-order', 'refund-order')),
		account TEXT NOT NULL CHECK (account LIKE 'xtb:%'),
		token TEXT NOT NULL CHECK (token LIKE 'xtb:%'),
		id TEXT NOT NULL,
		previous TEXT,
		balance REAL NOT NULL CHECK (balance >= 0),
		quote TEXT NOT NULL,
		price REAL NOT NULL CHECK (price >= 0),
		link TEXT NOT NULL,
		partial INTEGER NOT NULL,
		executor TEXT CHECK (executor LIKE 'xtb:%'),
		fee REAL,
		signature TEXT NOT NULL UNIQUE,
		hash TEXT NOT NULL UNIQUE,
		FOREIGN KEY (previous) REFERENCES orders(hash),
		PRIMARY KEY (hash)
		);`
	s["createConfirmsTable"] = `CREATE TABLE IF NOT EXISTS confirms(
		previous TEXT,
		addr TEXT NOT NULL CHECK (addr LIKE 'xtb:%'),
		head TEXT NOT NULL,
		account TEXT NOT NULL CHECK (account LIKE 'xtb:%'),
		signature TEXT NOT NULL UNIQUE,
		hash TEXT NOT NULL UNIQUE,
		FOREIGN KEY (previous) REFERENCES confirms(hash),
		PRIMARY KEY (hash)
		);`

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if err2 := tx.Rollback(); err2 != nil {
				fmt.Println(err2)
			}
			return
		}
		err = tx.Commit()
	}()

	for n, stmnt := range s {
		_, err = tx.Exec(stmnt)
		if err != nil {
			return fmt.Errorf("db: error executing statement %s: %s", n, err.Error())
		}
	}

	return nil
}

// Close releases all resources used by this database
func (m *DB) Close() error {
	return m.db.Close()
}

// InsertAccountBlock inserts the specified block into the database
func (m *DB) InsertAccountBlock(b *tradeblocks.AccountBlock) error {
	var previousOrNil interface{}
	if b.Previous != "" {
		previousOrNil = b.Previous
	} else {
		previousOrNil = nil
	}
	_, err := m.db.Exec(`INSERT INTO accounts (
		action,
		account,
		token,
		previous,
		representative,
		balance,
		link,
		signature,
		hash
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		b.Action,
		b.Account,
		b.Token,
		previousOrNil,
		b.Representative,
		b.Balance,
		b.Link,
		b.Signature,
		b.Hash())
	return err
}

// GetAccountBlocks gets all account blocks
func (m *DB) GetAccountBlocks() ([]*tradeblocks.AccountBlock, error) {
	rows, err := m.db.Query(`SELECT
		action,
		account,
		token,
		previous,
		representative,
		balance,
		link,
		signature
		FROM accounts`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*tradeblocks.AccountBlock
	for rows.Next() {
		b, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

// GetAccountBlock gets a block with the specified parameters
func (m *DB) GetAccountBlock(hash string) (*tradeblocks.AccountBlock, error) {
	row := m.db.QueryRow(`SELECT
		action,
		account,
		token,
		previous,
		representative,
		balance,
		link,
		signature
		FROM accounts WHERE hash = $1`, hash)
	return scanAccount(row)
}

func scanAccount(s scanner) (*tradeblocks.AccountBlock, error) {
	var b tradeblocks.AccountBlock
	var previous sql.NullString
	err := s.Scan(&b.Action, &b.Account, &b.Token, &previous, &b.Representative, &b.Balance, &b.Link, &b.Signature)
	if previous.Valid {
		b.Previous = previous.String
	}
	return &b, err
}

// InsertSwapBlock inserts the specified block into the database
func (m *DB) InsertSwapBlock(b *tradeblocks.SwapBlock) error {
	var previous interface{}
	if b.Previous != "" {
		previous = b.Previous
	} else {
		previous = nil
	}
	var right interface{}
	if b.Right != "" {
		right = b.Right
	} else {
		right = nil
	}
	var refundLeft interface{}
	if b.RefundLeft != "" {
		refundLeft = b.RefundLeft
	} else {
		refundLeft = nil
	}
	var refundRight interface{}
	if b.RefundRight != "" {
		refundRight = b.RefundRight
	} else {
		refundRight = nil
	}
	var executor interface{}
	if b.Executor != "" {
		executor = b.Executor
	} else {
		executor = nil
	}
	var fee interface{}
	if b.Fee != 0 {
		fee = b.Fee
	} else {
		fee = nil
	}
	_, err := m.db.Exec(`INSERT INTO swaps (
		action,
		account,
		token,
		id,
		previous,
		left,
		right,
		refund_left,
		refund_right,
		counterparty,
		want,
		quantity,
		executor,
		fee,
		signature,
		hash
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`,
		b.Action,
		b.Account,
		b.Token,
		b.ID,
		previous,
		b.Left,
		right,
		refundLeft,
		refundRight,
		b.Counterparty,
		b.Want,
		b.Quantity,
		executor,
		fee,
		b.Signature,
		b.Hash())
	return err
}

// GetSwapBlock gets a block with the specified parameters
func (m *DB) GetSwapBlock(hash string) (*tradeblocks.SwapBlock, error) {
	row := m.db.QueryRow(`SELECT
		action,
		account,
		token,
		id,
		previous,
		left,
		right,
		refund_left,
		refund_right,
		counterparty,
		want,
		quantity,
		executor,
		fee,
		signature
		FROM swaps WHERE hash = $1`, hash)
	return scanSwap(row)
}

// GetSwapBlocks gets all swap blocks
func (m *DB) GetSwapBlocks() ([]*tradeblocks.SwapBlock, error) {
	rows, err := m.db.Query(`SELECT
		action,
		account,
		token,
		id,
		previous,
		left,
		right,
		refund_left,
		refund_right,
		counterparty,
		want,
		quantity,
		executor,
		fee,
		signature
		FROM swaps`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*tradeblocks.SwapBlock
	for rows.Next() {
		b, err := scanSwap(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

func scanSwap(s scanner) (*tradeblocks.SwapBlock, error) {
	var b tradeblocks.SwapBlock
	var previous sql.NullString
	var right sql.NullString
	var refundLeft sql.NullString
	var refundRight sql.NullString
	var executor sql.NullString
	var fee sql.NullFloat64
	err := s.Scan(&b.Action,
		&b.Account,
		&b.Token,
		&b.ID,
		&previous,
		&b.Left,
		&right,
		&refundLeft,
		&refundRight,
		&b.Counterparty,
		&b.Want,
		&b.Quantity,
		&executor,
		&fee,
		&b.Signature)
	if previous.Valid {
		b.Previous = previous.String
	}
	if right.Valid {
		b.Right = right.String
	}
	if refundLeft.Valid {
		b.RefundLeft = refundLeft.String
	}
	if refundRight.Valid {
		b.RefundRight = refundRight.String
	}
	if executor.Valid {
		b.Executor = executor.String
	}
	if fee.Valid {
		b.Fee = fee.Float64
	}
	return &b, err
}

// InsertOrderBlock inserts the specified block into the database
func (m *DB) InsertOrderBlock(b *tradeblocks.OrderBlock) error {
	var previous interface{}
	if b.Previous != "" {
		previous = b.Previous
	} else {
		previous = nil
	}
	var executor interface{}
	if b.Executor != "" {
		executor = b.Executor
	} else {
		executor = nil
	}
	var fee interface{}
	if b.Fee != 0 {
		fee = b.Fee
	} else {
		fee = nil
	}
	_, err := m.db.Exec(`INSERT INTO orders (
		action,
		account,
		token,
		id,
		previous,
		balance,
		quote,
		price,
		link,
		partial,
		executor,
		fee,
		signature,
		hash
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		b.Action,
		b.Account,
		b.Token,
		b.ID,
		previous,
		b.Balance,
		b.Quote,
		b.Price,
		b.Link,
		b.Partial,
		executor,
		fee,
		b.Signature,
		b.Hash())
	return err
}

// GetOrderBlock gets a block with the specified parameters
func (m *DB) GetOrderBlock(hash string) (*tradeblocks.OrderBlock, error) {
	row := m.db.QueryRow(`SELECT
		action,
		account,
		token,
		id,
		previous,
		balance,
		quote,
		price,
		link,
		partial,
		executor,
		fee,
		signature
		FROM orders WHERE hash = $1`, hash)
	return scanOrder(row)
}

// GetOrderBlocks gets all order blocks
func (m *DB) GetOrderBlocks() ([]*tradeblocks.OrderBlock, error) {
	rows, err := m.db.Query(`SELECT
		action,
		account,
		token,
		id,
		previous,
		balance,
		quote,
		price,
		link,
		partial,
		executor,
		fee,
		signature
		FROM orders`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*tradeblocks.OrderBlock
	for rows.Next() {
		b, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

func scanOrder(s scanner) (*tradeblocks.OrderBlock, error) {
	var b tradeblocks.OrderBlock
	var previous sql.NullString
	var executor sql.NullString
	var fee sql.NullFloat64
	err := s.Scan(
		&b.Action,
		&b.Account,
		&b.Token,
		&b.ID,
		&previous,
		&b.Balance,
		&b.Quote,
		&b.Price,
		&b.Link,
		&b.Partial,
		&executor,
		&fee,
		&b.Signature)
	if previous.Valid {
		b.Previous = previous.String
	}
	if executor.Valid {
		b.Executor = executor.String
	}
	if fee.Valid {
		b.Fee = fee.Float64
	}
	return &b, err
}

// InsertConfirmBlock inserts the specified block into the database
func (m *DB) InsertConfirmBlock(b *tradeblocks.ConfirmBlock) error {
	var previous interface{}
	if b.Previous != "" {
		previous = b.Previous
	} else {
		previous = nil
	}
	_, err := m.db.Exec(`INSERT INTO confirms (
		previous,
		addr,
		head,
		account,
		signature,
		hash
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		previous,
		b.Addr,
		b.Head,
		b.Account,
		b.Signature,
		b.Hash())
	return err
}

// GetConfirmBlock gets a block with the specified parameters
func (m *DB) GetConfirmBlock(hash string) (*tradeblocks.ConfirmBlock, error) {
	row := m.db.QueryRow(`SELECT
		previous,
		addr,
		head,
		account,
		signature
		FROM confirms WHERE hash = $1`, hash)
	return scanConfirm(row)
}

// GetConfirmBlocks gets all confirm blocks
func (m *DB) GetConfirmBlocks() ([]*tradeblocks.ConfirmBlock, error) {
	rows, err := m.db.Query(`SELECT
		previous,
		addr,
		head,
		account,
		signature
		FROM confirms`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*tradeblocks.ConfirmBlock
	for rows.Next() {
		b, err := scanConfirm(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

func scanConfirm(s scanner) (*tradeblocks.ConfirmBlock, error) {
	var b tradeblocks.ConfirmBlock
	var previous sql.NullString
	err := s.Scan(
		&previous,
		&b.Addr,
		&b.Head,
		&b.Account,
		&b.Signature)
	if previous.Valid {
		b.Previous = previous.String
	}
	return &b, err
}
