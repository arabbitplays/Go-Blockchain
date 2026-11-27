package blockchain

import (
	"encoding/hex"

	"github.com/boltdb/bolt"
)

type UTXOSet struct {
	blockchain *Blockchain
}

const utxoBucket = "chainstate"

func NewUTXOSet(bc *Blockchain) *UTXOSet {
	return &UTXOSet{bc}
}

func (u UTXOSet) Reindex() {
	db := u.blockchain.db
	bucketName := []byte(utxoBucket)

	db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket(bucketName)
		tx.CreateBucket(bucketName)
		return nil
	})

	utxo := u.blockchain.FindUTXO()

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txID, outs := range utxo {
			key, _ := hex.DecodeString(txID)
			serializedOuts, _ := outs.Serialize()
			b.Put(key, serializedOuts)
		}
		return nil
	})
}

func (u UTXOSet) Update(block *Block) {
	db := u.blockchain.db

	db.Update(func(dbTX *bolt.Tx) error {
		b := dbTX.Bucket([]byte(utxoBucket))

		for _, tx := range block.Transactions {
			if IsCoinbaseTX(tx) == false {
				for _, in := range tx.Vin {
					outsBytes := b.Get(in.TxID)
					outs, _ := DeserializeOutputs(outsBytes)

					updatedOuts := TXOutputs{}
					for outIdx, out := range outs.Outputs {
						if outIdx != in.Vout {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						b.Delete(in.TxID)
					} else {
						outsBytes, _ = updatedOuts.Serialize()
						b.Put(in.TxID, outsBytes)
					}
				}
			}

			newOuts := TXOutputs{}
			for _, out := range tx.Vout {
				newOuts.Outputs = append(newOuts.Outputs, out)
			}
			outsBytes, _ := newOuts.Serialize()
			b.Put(tx.ID, outsBytes)
		}

		return nil
	})
}

func (u UTXOSet) FindSpendableOutputs(publicKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulatedValue := 0

	db := u.blockchain.db

	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs, _ := DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(publicKeyHash) && accumulatedValue < amount {
					accumulatedValue += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}

		return nil
	})

	return accumulatedValue, unspentOutputs
}

func (u UTXOSet) FindUnspentOutputs(publicKeyHash []byte) []TXOutput {
	var unspentOutputs []TXOutput

	db := u.blockchain.db

	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs, _ := DeserializeOutputs(v)

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(publicKeyHash) {
					unspentOutputs = append(unspentOutputs, out)
				}
			}
		}

		return nil
	})

	return unspentOutputs
}
