package blockchain

import (
	"fmt"
	"os"
	"log"
	"github.com/boltdb/bolt"
)

type Blockchain struct {
	tip []byte
	db *bolt.DB
}

const dbFile = "blockchain_%s.db"
const nodeId = "oschdi"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

func (bc *Blockchain) AddBlock(transactions []*Transaction) {
	var lastHash []byte

	// TODO error handling -> how do transactions work here?
	_ = bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})

	newBlock := NewBlock(transactions, lastHash)

	_ = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock, _ := newBlock.Serialize()
		err := b.Put(newBlock.Hash, encodedBlock)
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		bc.tip = newBlock.Hash

		return nil
	})
}

func newGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

func CreateBlockchain(address string) *Blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeId)
	if dbExists(dbFile) {
		fmt.Println("Blockchain already exists!")
		os.Exit(1)
	}

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	var tip []byte

	coinbase := NewCoinbaseTX(address, genesisCoinbaseData)
	genesis := newGenesisBlock(coinbase)

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		encodedBlock, _ := genesis.Serialize()
		err = b.Put(genesis.Hash, encodedBlock)
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}

		tip = genesis.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}
 
func NewBlockchain() *Blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeId)
	if dbExists(dbFile) == false {
		fmt.Println("Blockchain does not exist, create it first!")
		os.Exit(1)
	}

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	var tip []byte

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

func (bc *Blockchain) Destroy() {
	bc.db.Close()
}

func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}
