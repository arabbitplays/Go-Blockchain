package blockchain

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

type TXOutputs struct {
	Outputs []TXOutput
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

func (block TXOutputs) Serialize() ([]byte, error) {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(block)
	if err != nil {
		return nil, fmt.Errorf("encoding outputs failed: %v", err)
	}

	return result.Bytes(), err
}

func DeserializeOutputs(d []byte) (TXOutputs, error) {
	var block TXOutputs

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		return TXOutputs{}, fmt.Errorf("decoding block failed: %v", err)
	}

	return block, err
}
