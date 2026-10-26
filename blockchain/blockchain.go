package blockchain

import (
	"github.com/boltdb/bolt"
)

type Blockchain struct {
	tip []byte
	db *bolt.DB
}

const dbFile = "blockchain_%s.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

func (bc *Blockchain) AddBlock(data string) {
	var lastHash []byte

	// TODO error handling
	_ = bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})

	newBlock := NewBlock(data, lastHash)

	_ = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		_ = b.Put(newBlock.Hash, newBlock.Serialize())
		_ = b.Put([]byte("l"), newBlock.Hash)
		bc.tip = newBlock.Hash

		return nil
	})
}

func newGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}
 
func NewBlockchain() *Blockchain {
	var tip []byte
	// TODO error handling
	db, _ := bolt.Open(dbFile, 0600, nil)

	_ = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			genesis := newGenesisBlock()
			b, _ := tx.CreateBucket([]byte(blocksBucket))
			_ = b.Put(genesis.Hash, genesis.Serialize())
			_ = b.Put([]byte("l"), genesis.Hash)
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}

		return nil
	})

	bc := Blockchain{tip, db}

	return &bc
}

func (bc *Blockchain) Destroy() {
	bc.db.Close()
}
