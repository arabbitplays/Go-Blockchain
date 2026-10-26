package blockchain
 
import (
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

	_ = iter.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(iter.currentHash)
		block = DeserializeBlock(encodedBlock) 
		
		return nil;
	})

	iter.currentHash = block.PrevBlockHash
	 
	return block
}
