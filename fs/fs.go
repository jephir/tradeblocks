package fs

import (
	"encoding/json"
	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
	"io/ioutil"
	"os"
	"path/filepath"
)

// BlockStorage saves and loads blocks on the filesystem
type BlockStorage struct {
	blockstore *app.BlockStore2
	dir        string
}

// NewBlockStorage returns a new storage adapter for the specified blockstore and data directory
func NewBlockStorage(blockstore *app.BlockStore2, dir string) *BlockStorage {
	return &BlockStorage{
		blockstore: blockstore,
		dir:        dir,
	}
}

// Save saves all blocks to the filesystem
func (s *BlockStorage) Save() error {
	// Create storage directory
	if err := s.createStorageDir(); err != nil {
		return err
	}
	// Serialize each block and write it to the directory
	var err error
	s.blockstore.AccountBlocks(func(sequence int, b *tradeblocks.AccountBlock) bool {
		hash := b.Hash()
		err = s.SaveAccountBlock(hash, b)
		if err != nil {
			return false
		}
		return true
	})
	s.blockstore.SwapBlocks(func(sequence int, b *tradeblocks.SwapBlock) bool {
		hash := b.Hash()
		err = s.SaveSwapBlock(hash, b)
		if err != nil {
			return false
		}
		return true
	})
	s.blockstore.OrderBlocks(func(sequence int, b *tradeblocks.OrderBlock) bool {
		hash := b.Hash()
		err = s.SaveOrderBlock(hash, b)
		if err != nil {
			return false
		}
		return true
	})
	return err
}

func (s *BlockStorage) createStorageDir() error {
	if err := os.MkdirAll(accountsDir(s.dir), 0700); err != nil {
		return err
	}
	if err := os.MkdirAll(swapsDir(s.dir), 0700); err != nil {
		return err
	}
	if err := os.MkdirAll(ordersDir(s.dir), 0700); err != nil {
		return err
	}
	return nil
}

// SaveAccountBlock saves the specified block with the specified hash to the filesystem
func (s *BlockStorage) SaveAccountBlock(hash string, block *tradeblocks.AccountBlock) error {
	p := filepath.Join(accountsDir(s.dir), hash)
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(block); err != nil {
		return err
	}
	return f.Close()
}

// SaveSwapBlock saves the specified block with the specified hash to the filesystem
func (s *BlockStorage) SaveSwapBlock(hash string, block *tradeblocks.SwapBlock) error {
	p := filepath.Join(swapsDir(s.dir), hash)
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(block); err != nil {
		return err
	}
	return f.Close()
}

// SaveOrderBlock saves the specified block with the specified hash to the filesystem
func (s *BlockStorage) SaveOrderBlock(hash string, block *tradeblocks.OrderBlock) error {
	p := filepath.Join(ordersDir(s.dir), hash)
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(block); err != nil {
		return err
	}
	return f.Close()
}

// Load loads all blocks from the filesystem
func (s *BlockStorage) Load() error {
	if err := s.loadAccounts(); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := s.loadSwaps(); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := s.loadOrders(); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *BlockStorage) loadAccounts() error {
	dir := accountsDir(s.dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		p := filepath.Join(dir, file.Name())
		if err := s.loadAccountBlock(p); err != nil {
			return err
		}
	}
	return nil
}

func (s *BlockStorage) loadSwaps() error {
	dir := swapsDir(s.dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		p := filepath.Join(dir, file.Name())
		if err := s.loadSwapBlock(p); err != nil {
			return err
		}
	}
	return nil
}

func (s *BlockStorage) loadOrders() error {
	dir := ordersDir(s.dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		p := filepath.Join(dir, file.Name())
		if err := s.loadOrderBlock(p); err != nil {
			return err
		}
	}
	return nil
}

func (s *BlockStorage) loadAccountBlock(p string) error {
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()
	var b tradeblocks.AccountBlock
	if err := json.NewDecoder(f).Decode(&b); err != nil {
		return err
	}
	s.blockstore.AddAccountBlock(&b)
	return nil
}

func (s *BlockStorage) loadSwapBlock(p string) error {
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()
	var b tradeblocks.SwapBlock
	if err := json.NewDecoder(f).Decode(&b); err != nil {
		return err
	}
	s.blockstore.AddSwapBlock(&b)
	return nil
}

func (s *BlockStorage) loadOrderBlock(p string) error {
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()
	var b tradeblocks.OrderBlock
	if err := json.NewDecoder(f).Decode(&b); err != nil {
		return err
	}
	s.blockstore.AddOrderBlock(&b)
	return nil
}

// Dir returns the working directory of this storage
func (s *BlockStorage) Dir() string {
	return s.dir
}

func accountsDir(root string) string {
	return filepath.Join(root, "accounts")
}

func swapsDir(root string) string {
	return filepath.Join(root, "swaps")
}

func ordersDir(root string) string {
	return filepath.Join(root, "orders")
}
