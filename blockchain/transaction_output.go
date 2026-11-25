package blockchain

import (
	"bytes"
)

type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

func NewTXOutput(value int, address string) *TXOutput {
	out := TXOutput{value, nil}
	out.Lock([]byte(address))

	return &out
}

func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := GetPubKeyHashFromAddress(address)
	out.PubKeyHash = pubKeyHash
}

func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Equal(out.PubKeyHash, pubKeyHash)
}
