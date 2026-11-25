package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
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
	tx.Sign(wallet.PrivateKey, bc.FindReferencedTransactions(&tx))

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

func (bc *Blockchain) FindReferencedTransactions(tx *Transaction) map[string]Transaction {
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Vin {
		refTX, _ := bc.FindTransaction(in.TxID)
		prevTXs[hex.EncodeToString(refTX.ID)] = refTX
	}

	return prevTXs
}

func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	iter := bc.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")
}

func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if IsCoinbaseTX(tx) {
		return
	}

	txCopy := tx.TrimmedCopy()

	for inIdx, inCopy := range txCopy.Vin {
		prevTX := prevTXs[hex.EncodeToString(inCopy.TxID)]

		txCopy.Vin[inIdx].PubKey = prevTX.Vout[inCopy.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inIdx].PubKey = nil

		r, s, _ := ecdsa.Sign(rand.Reader, &privateKey, txCopy.ID)
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vin[inIdx].Signature = signature
	}
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inIdx, in := range tx.Vin {
		prevTX := prevTXs[hex.EncodeToString(in.TxID)]

		// TODO why not HashPubKey(in.PubKey) ????
		txCopy.Vin[inIdx].PubKey = prevTX.Vout[in.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inIdx].PubKey = nil

		r := big.Int{}
		s := big.Int{}
		sigLen := len(in.Signature)
		r.SetBytes(in.Signature[:sigLen/2])
		s.SetBytes(in.Signature[sigLen/2:])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(in.PubKey)
		x.SetBytes(in.PubKey[:keyLen/2])
		y.SetBytes(in.PubKey[keyLen/2:])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}

		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
			return false
		}
	}

	return true
}

func (tx *Transaction) TrimmedCopy() *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, in := range tx.Vin {
		newIn := TXInput{in.TxID, in.Vout, nil, nil}
		inputs = append(inputs, newIn)
	}

	for _, out := range tx.Vout {
		newOut := TXOutput{out.Value, out.PubKeyHash}
		outputs = append(outputs, newOut)
	}

	txCopy := &Transaction{tx.ID, inputs, outputs}
	return txCopy
}
