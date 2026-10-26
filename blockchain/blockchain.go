package blockchain

type Blockchain struct {
	Blocks []*Block
}

func (chain *Blockchain) AddBlock(data string) {
	prevBlock := chain.Blocks[len(chain.Blocks) - 1]
	block := NewBlock(data, prevBlock.Hash)
	chain.Blocks = append(chain.Blocks, block)
}

func newGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}
 
func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{newGenesisBlock()}}
}
