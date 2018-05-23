package web

import (
	"encoding/json"
	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
	"io/ioutil"
	"os"
	"path/filepath"
)

type blockStorage struct {
	blockstore *app.BlockStore
	dir        string
}

func newBlockStorage(blockstore *app.BlockStore, dir string) *blockStorage {
	return &blockStorage{
		blockstore: blockstore,
		dir:        dir,
	}
}

func (s *blockStorage) save() error {
	// Create storage directory
	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return err
	}
	// Serialize each block and write it to the directory
	for hash, block := range s.blockstore.AccountBlocks {
		if err := s.saveBlock(hash, block); err != nil {
			return err
		}
	}
	return nil
}

func (s *blockStorage) saveBlock(hash string, block *tradeblocks.AccountBlock) error {
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

func (s *blockStorage) load() error {
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

func (s *blockStorage) loadBlock(hash string) error {
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
