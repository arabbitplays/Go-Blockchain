package blockchain

import (
	"fmt"
	"errors"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"log"
	"bytes"
)

const subsidy = 10

type TXInput struct {
	TxId []byte
	Vout int
	ScriptSig string
}

type TXOutput struct {
	Value int
	ScriptPubKey string
}

type Transaction struct {
	ID []byte
	Vin []TXInput
	Vout []TXOutput
}

func (tx *Transaction) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	
	err  := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:] 
}

func IsCoinbaseTX(tx *Transaction) bool {
	return len(tx.Vin) == 1 && tx.Vin[0].Vout == -1
}

func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to: %s", to)
	}

	txIn := TXInput{[]byte{}, -1, data}
	txOut := TXOutput{subsidy, to}
	tx := Transaction{nil, []TXInput{txIn}, []TXOutput{txOut}}
	tx.ID = tx.Hash()

	return &tx
}

func NewUTXOTransaction(from string, to string, amount int, bc *Blockchain) (*Transaction, error) {
	var inputs []TXInput
	var outputs []TXOutput

	acc, spendableOutputs := bc.FindSpendableOutputs(from, amount)

	if acc < amount {
		msg := fmt.Sprintf("Balance %d of account '%s' is not enough to create transaction with amount %d", acc, from, amount)
		return nil, errors.New(msg)
	}

	for txid, outIndices := range spendableOutputs {
		txId, _ := hex.DecodeString(txid)

		for _, outIdx := range outIndices {
			in := TXInput{txId, outIdx, from}
			inputs = append(inputs, in)
		}
	}

	spentOutput := TXOutput{amount, to}
	outputs = append(outputs, spentOutput)

	if acc > amount {
		change := acc - amount
		changeOutput := TXOutput{change, from}
		outputs = append(outputs, changeOutput)
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()

	return &tx, nil
}

func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	spendableOutputs := make(map[string][]int)
	unspentTXs := bc.findUnspentTransactions(address)

	accumulatedValue := 0

	Work:
	for _, tx := range(unspentTXs) {
		txId := hex.EncodeToString(tx.ID)
		for outIdx, out := range(tx.Vout) {
			if out.CanBeUnlockedWith(address) {
				accumulatedValue += out.Value
				spendableOutputs[txId] = append(spendableOutputs[txId], outIdx)

				if accumulatedValue >= amount {
					break Work
				}
			}
		}
	} 

	return accumulatedValue, spendableOutputs
}

func (bc *Blockchain) FindUnspentOutputs(address string) []TXOutput {
	var UTXO []TXOutput
	unspentTXs := bc.findUnspentTransactions(address)

	for _, tx := range(unspentTXs) {
		for _, out := range(tx.Vout) {
			if out.CanBeUnlockedWith(address) {
				UTXO = append(UTXO, out)
			}
		}
	} 

	return UTXO
}

func (bc *Blockchain) findUnspentTransactions(address string) []Transaction {
	iter := bc.Iterator()
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)

	for {
		block := iter.Next()

		for _, tx := range(block.Transactions) {
			txId := hex.EncodeToString(tx.ID)

			Outputs:
			for outIdx, out := range(tx.Vout) {
				if spentTXOs[txId] != nil {
					for _, spentOut := range(spentTXOs[txId]) {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if IsCoinbaseTX(tx) == false {
				for _, in := range(tx.Vin) {
					if in.CanUnlockOutputWith(address) {
						inTxId := hex.EncodeToString(in.TxId)
						spentTXOs[inTxId] = append(spentTXOs[inTxId], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

// returns true if this input can unlock an output that needs the given data to be unlocked
func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}

