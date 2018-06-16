package db

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3" // sqlite driver
)

// DB represents a database
type DB struct {
	db *sql.DB
}

// NewDB connects to the specified data source
func NewDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, err
	}
	return &DB{
		db: db,
	}, nil
}

func (db *DB) init() (err error) {
	s := make(map[string]string)
	s["createAccountsTable"] = `CREATE TABLE IF NOT EXISTS accounts(
		action TEXT NOT NULL CHECK (action IN ('open', 'issue', 'send', 'receive')),
		account TEXT NOT NULL CHECK (account LIKE xtb:%),
		token TEXT NOT NULL CHECK (token LIKE xtb:%),
		previous TEXT UNIQUE,
			FOREIGN KEY (previous) REFERENCES accounts(hash),
		representative TEXT NOT NULL CHECK (representative LIKE xtb:%),
			FOREIGN KEY (representative) REFERENCES accounts(account),
		balance REAL NOT NULL CHECK (balance >= 0),
		link TEXT UNIQUE,
		signature TEXT NOT NULL UNIQUE,
		hash TEXT NOT NULL UNIQUE,
		PRIMARY KEY (account, token)
		);`
	s["createSwapsTable"] = `CREATE TABLE IF NOT EXISTS swaps(
		action TEXT NOT NULL CHECK (action IN ('offer', 'commit', 'refund-left', 'refund-right')),
		account TEXT NOT NULL CHECK (account LIKE xtb:%),
			FOREIGN KEY (account) REFERENCES accounts(account),
		token TEXT NOT NULL CHECK (token LIKE xtb:%),
			FOREIGN KEY (token) REFERENCES accounts(token),
		id TEXT NOT NULL,
		previous TEXT UNIQUE,
			FOREIGN KEY (previous) REFERENCES swaps(hash),
		left TEXT UNIQUE NOT NULL,
		right TEXT UNIQUE,
		refund_left TEXT CHECK (refund_left LIKE xtb:%),
		refund_right TEXT CHECK (refund_right LIKE xtb:%),
		counterparty TEXT NOT NULL CHECK (counterparty LIKE xtb:%),
			FOREIGN KEY (counterparty) REFERENCES accounts(account),
		want TEXT NOT NULL CHECK (want LIKE xtb:%),
			FOREIGN KEY (want) REFERENCES accounts(token),
		quantity REAL NOT NULL CHECK (quantity >= 0),
		executor TEXT CHECK (executor LIKE xtb:%),
			FOREIGN KEY executor REFERENCES orders(executor),
		fee REAL,
		signature TEXT NOT NULL UNIQUE,
		hash TEXT NOT NULL UNIQUE,
		PRIMARY KEY (account, id)
		);`
	s["createOrdersTable"] = `CREATE TABLE IF NOT EXISTS orders(
		action TEXT NOT NULL CHECK (action IN ('create-order', 'accept-order', 'refund-order')),
		account TEXT NOT NULL CHECK (account LIKE xtb:%),
			FOREIGN KEY (account) REFERENCES accounts(account),
		token TEXT NOT NULL CHECK (token LIKE xtb:%),
			FOREIGN KEY (token) REFERENCES accounts(token),
		id TEXT NOT NULL,
		previous TEXT,
			FOREIGN KEY (previous) REFERENCES orders(hash),
		balance REAL NOT NULL CHECK (balance >= 0),
		quote TEXT NOT NULL,
			FOREIGN KEY (quote) REFERENCES accounts(account),
		price REAL NOT NULL CHECK (price >= 0),
		link TEXT NOT NULL,
		partial INTEGER NOT NULL,
		executor TEXT CHECK (executor LIKE xtb:%),
			FOREIGN KEY executor REFERENCES accounts(account),
		fee REAL,
		signature TEXT NOT NULL UNIQUE,
		hash TEXT NOT NULL UNIQUE,
		PRIMARY KEY (account, id)`

	tx, err := db.db.Begin()
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
func (db *DB) Close() error {
	return db.Close()
}
