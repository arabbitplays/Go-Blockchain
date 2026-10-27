package blockchain

import (
	"fmt"
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

	return &tx
}
