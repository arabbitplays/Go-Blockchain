package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
)

const subsidy = 10

type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

func (tx *Transaction) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(tx)
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

func NewCoinbaseTX(to string, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to: %s", to)
	}

	walletRepo := NewWalletRepo()
	wallet := walletRepo.GetWallet(to)

	txIn := TXInput{[]byte{}, -1, nil, []byte(data)}
	txOut := TXOutput{subsidy, HashPubKey(wallet.PublicKey)}
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

	walletRepo := NewWalletRepo()
	wallet := walletRepo.GetWallet(from)

	for txIDHex, outIndices := range spendableOutputs {
		txID, _ := hex.DecodeString(txIDHex)

		for _, outIdx := range outIndices {
			in := TXInput{txID, outIdx, nil, wallet.PublicKey}
			inputs = append(inputs, in)
		}
	}

	spentOutput := *NewTXOutput(amount, to)
	outputs = append(outputs, spentOutput)

	if acc > amount {
		change := acc - amount
		changeOutput := *NewTXOutput(change, from)
		outputs = append(outputs, changeOutput)
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()

	return &tx, nil
}

func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	spendableOutputs := make(map[string][]int)
	pubKeyHash := GetPubKeyHashFromAddress([]byte(address))
	unspentTXs := bc.findUnspentTransactions(pubKeyHash)

	accumulatedValue := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) {
				accumulatedValue += out.Value
				spendableOutputs[txID] = append(spendableOutputs[txID], outIdx)

				if accumulatedValue >= amount {
					break Work
				}
			}
		}
	}

	return accumulatedValue, spendableOutputs
}

func (bc *Blockchain) FindUnspentOutputs(address string) []TXOutput {
	var unspentOutputs []TXOutput
	pubKeyHash := GetPubKeyHashFromAddress([]byte(address))
	unspentTXs := bc.findUnspentTransactions(pubKeyHash)

	for _, tx := range unspentTXs {
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) {
				unspentOutputs = append(unspentOutputs, out)
			}
		}
	}

	return unspentOutputs
}

func (bc *Blockchain) findUnspentTransactions(pubKeyHash []byte) []Transaction {
	iter := bc.Iterator()
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				if out.IsLockedWithKey(pubKeyHash) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if IsCoinbaseTX(tx) == false {
				for _, in := range tx.Vin {
					if in.UsesKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.TxID)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
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
