package blockchain
 
import (
	"log"
	"github.com/boltdb/bolt"
)

type BlockchainIterator struct {
	currentHash []byte
	db *bolt.DB
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	iter := &BlockchainIterator{bc.tip, bc.db}
	return iter;
}

func (iter *BlockchainIterator) Next() *Block {
	var block *Block

	// TODO error handling with transactions?
	_ = iter.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(iter.currentHash)

		var err error
		block, err = DeserializeBlock(encodedBlock) 
		if err != nil {
			log.Panic(err)
		}
		
		return nil;
	})

	iter.currentHash = block.PrevBlockHash
	 
	return block
}
