package blockchain

import (
	"fmt"
	"log"
)

type Client struct {
	nodeID     string
	blockchain *Blockchain
	utxoSet    *UTXOSet
}

func NewClient(nodeID string) *Client {
	client := &Client{nodeID: nodeID}
	if !dbExists(fmt.Sprintf(dbFile, nodeID)) {
		return client
	}

	bc := Open(nodeID)
	client.blockchain = bc
	client.utxoSet = NewUTXOSet(bc)

	return client
}

func (client *Client) Close() {
	if client.blockchain == nil {
		return
	}

	client.blockchain.Close()
	client.blockchain = nil
	client.utxoSet = nil
}

func (client *Client) CreateBlockchain(address string) {
	client.blockchain = Create(client.nodeID, address)
	client.utxoSet = NewUTXOSet(client.blockchain)
	client.utxoSet.Reindex()
}

func (client *Client) GetBalance(address string) (int, error) {
	if !ValidateAddress(address) {
		return 0, fmt.Errorf("ERROR: Address is not valid")
	}

	publicKeyHash := GetPubKeyHashFromAddress([]byte(address))
	unspentOutputs, err := client.utxoSet.FindUnspentOutputs(publicKeyHash)
	if err != nil {
		return 0, err
	}

	balance := 0
	for _, out := range unspentOutputs {
		balance += out.Value
	}

	return balance, nil
}

func (client *Client) CreateWallet() string {
	walletRepo := NewWalletRepo(client.nodeID)
	wallet := walletRepo.CreateNewWallet()

	return string(wallet.GetAddress())
}

func (client *Client) ListAddresses() []string {
	walletRepo := NewWalletRepo(client.nodeID)
	return walletRepo.GetAddresses()
}

func (client *Client) ReindexUTXO() {
	client.utxoSet.Reindex()
}

func (client *Client) Blocks() []*Block {
	iter := client.blockchain.Iterator()
	var blocks []*Block

	for {
		block := iter.Next()
		blocks = append(blocks, block)

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return blocks
}

func (client *Client) Send(from string, to string, amount int) error {
	if !ValidateAddress(from) {
		return fmt.Errorf("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		return fmt.Errorf("ERROR: Recipient address is not valid")
	}

	tx, err := NewUTXOTransaction(client.nodeID, from, to, amount, client.blockchain)
	if err != nil {
		return err
	}

	cbtx := NewCoinbaseTX(client.nodeID, from, "")
	block, err := client.blockchain.MineBlock([]*Transaction{cbtx, tx})
	if err != nil {
		return err
	}

	// TODO bundle these into one transaction or add tip marker to utxo set
	err = client.blockchain.AddBlock(block)
	if err != nil {
		log.Panic(err)
	}
	err = client.utxoSet.Update(block)
	if err != nil {
		log.Panic(err)
	}

	return nil
}
