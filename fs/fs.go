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
	blockstore *app.BlockStore
	dir        string
}

// NewBlockStorage returns a new storage adapter for the specified blockstore and data directory
func NewBlockStorage(blockstore *app.BlockStore, dir string) *BlockStorage {
	return &BlockStorage{
		blockstore: blockstore,
		dir:        dir,
	}
}

// Save saves all blocks to the filesystem
func (s *BlockStorage) Save() error {
	// Create storage directory
	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return err
	}
	// Serialize each block and write it to the directory
	for hash, block := range s.blockstore.AccountBlocks {
		if err := s.SaveBlock(hash, block); err != nil {
			return err
		}
	}
	return nil
}

// SaveBlock saves the specified block with the specified hash to the filesystem
func (s *BlockStorage) SaveBlock(hash string, block *tradeblocks.AccountBlock) error {
	p := filepath.Join(s.dir, hash)
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
	files, err := ioutil.ReadDir(s.dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if err := s.loadBlock(file.Name()); err != nil {
			return err
		}
	}
	return nil
}

func (s *BlockStorage) loadBlock(hash string) error {
	p := filepath.Join(s.dir, hash)
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()
	var b tradeblocks.AccountBlock
	if err := json.NewDecoder(f).Decode(&b); err != nil {
		return err
	}
	s.blockstore.AccountBlocks[hash] = &b
	return nil
}

// Dir returns the working directory of this storage
func (s *BlockStorage) Dir() string {
	return s.dir
}
