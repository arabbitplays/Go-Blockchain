package blockchain

import (
	"time"
	"bytes"
	"encoding/gob"
	"fmt"
)

type Block struct {
	Timestamp int64
	Data []byte
	PrevBlockHash []byte
	Hash []byte
	Nonce int
}

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	
	block.Hash = hash[:]
	block.Nonce = nonce 
	return block
}

func (block *Block) Serialize() ([]byte, error) {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(block) 
	if (err != nil) {
		return nil, fmt.Errorf("encoding block failed: %v", err)
	}

	return result.Bytes(), err
}

func DeserializeBlock(d []byte) (*Block, error) {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block) // TODO handle errors
	if (err != nil) {
		return nil, fmt.Errorf("decoding block failed: %v", err)
	}

	return &block, err
}
