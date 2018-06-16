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
		FOREIGN KEY (representative) REFERENCES accounts(account),
		PRIMARY KEY (account, token)
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
		FOREIGN KEY (account) REFERENCES accounts(account),
		FOREIGN KEY (token) REFERENCES accounts(token),
		FOREIGN KEY (previous) REFERENCES swaps(hash),
		FOREIGN KEY (counterparty) REFERENCES accounts(account),
		FOREIGN KEY (want) REFERENCES accounts(token),
		FOREIGN KEY (executor) REFERENCES orders(executor),
		PRIMARY KEY (account, id)
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
		FOREIGN KEY (account) REFERENCES accounts(account),
		FOREIGN KEY (token) REFERENCES accounts(token),
		FOREIGN KEY (previous) REFERENCES orders(hash),
		FOREIGN KEY (quote) REFERENCES accounts(account),
		FOREIGN KEY (executor) REFERENCES accounts(account),
		PRIMARY KEY (account, id)
		);`
	s["createConfirmsTable"] = `CREATE TABLE IF NOT EXISTS confirms(
		previous TEXT,
		addr TEXT NOT NULL CHECK (addr LIKE 'xtb:%'),
		head TEXT NOT NULL,
		account TEXT NOT NULL CHECK (account LIKE 'xtb:%'),
		signature TEXT NOT NULL UNIQUE,
		hash TEXT NOT NULL UNIQUE,
		FOREIGN KEY (previous) REFERENCES confirms(hash),
		FOREIGN KEY (account) REFERENCES accounts(account),
		PRIMARY KEY (account, addr)
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
