package blockchain

import (
	"encoding/hex"
	"fmt"

	"github.com/boltdb/bolt"
)

type UTXOSet struct {
	blockchain *Blockchain
}

const utxoBucket = "chainstate"

func NewUTXOSet(bc *Blockchain) *UTXOSet {
	return &UTXOSet{bc}
}

func (u UTXOSet) Reindex() error {
	db := u.blockchain.db
	bucketName := []byte(utxoBucket)

	err := db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket(bucketName)
		_, err := tx.CreateBucket(bucketName)
		return err
	})
	if err != nil {
		return err
	}

	utxo := u.blockchain.FindUTXO()

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return fmt.Errorf("Bucket %s does not exist", utxoBucket)
		}

		for txID, outs := range utxo {
			key, _ := hex.DecodeString(txID)
			serializedOuts, _ := outs.Serialize()
			if err = b.Put(key, serializedOuts); err != nil {
				return err
			}
		}
		return nil
	})

	return nil
}

func (u UTXOSet) Update(block *Block) error {
	db := u.blockchain.db

	err := db.Update(func(dbTX *bolt.Tx) error {
		b := dbTX.Bucket([]byte(utxoBucket))
		if b == nil {
			return fmt.Errorf("Bucket %s does not exist", utxoBucket)
		}

		for _, tx := range block.Transactions {
			if IsCoinbaseTX(tx) == false {
				for _, in := range tx.Vin {
					outsBytes := b.Get(in.TxID)
					outs, err := DeserializeOutputs(outsBytes)
					if err != nil {
						return err
					}

					updatedOuts := TXOutputs{}
					for outIdx, out := range outs.Outputs {
						if outIdx != in.Vout {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						if err = b.Delete(in.TxID); err != nil {
							return err
						}
					} else {
						outsBytes, err = updatedOuts.Serialize()
						if err != nil {
							return err
						}
						if err = b.Put(in.TxID, outsBytes); err != nil {
							return err
						}
					}
				}
			}

			newOuts := TXOutputs{}
			for _, out := range tx.Vout {
				newOuts.Outputs = append(newOuts.Outputs, out)
			}
			outsBytes, err := newOuts.Serialize()
			if err != nil {
				return err
			}
			if err = b.Put(tx.ID, outsBytes); err != nil {
				return err
			}
		}

		return nil
	})
	return err
}

func (u UTXOSet) FindSpendableOutputs(publicKeyHash []byte, amount int) (int, map[string][]int, error) {
	unspentOutputs := make(map[string][]int)
	accumulatedValue := 0

	db := u.blockchain.db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		if b == nil {
			return fmt.Errorf("Bucket %s does not exist", utxoBucket)
		}
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs, err := DeserializeOutputs(v)
			if err != nil {
				return fmt.Errorf("Corrupted data in utxo: %v", err)
			}

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(publicKeyHash) && accumulatedValue < amount {
					accumulatedValue += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}

		return nil
	})

	return accumulatedValue, unspentOutputs, err
}

func (u UTXOSet) FindUnspentOutputs(publicKeyHash []byte) ([]TXOutput, error) {
	var unspentOutputs []TXOutput

	db := u.blockchain.db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		if b == nil {
			return fmt.Errorf("Bucket %s does not exist", utxoBucket)
		}
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs, err := DeserializeOutputs(v)
			if err != nil {
				return fmt.Errorf("Corrupted data in utxo: %v", err)
			}

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(publicKeyHash) {
					unspentOutputs = append(unspentOutputs, out)
				}
			}
		}

		return nil
	})

	return unspentOutputs, err
}
