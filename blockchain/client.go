package blockchain

import (
	"fmt"
	"log"
	"os"
)

type Client struct {
	nodeID     string
	blockchain *Blockchain
	utxoSet    *UTXOSet
}

func NewClient(nodeID string) *Client {
	bc := Open()
	utxoSet := NewUTXOSet(bc)
	defer bc.Close()

	client := Client{nodeID, bc, utxoSet}
	return &client
}

func (client *Client) GetBalance(address string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}

	publicKeyHash := GetPubKeyHashFromAddress([]byte(address))
	unspentOutputs := client.utxoSet.FindUnspentOutputs(publicKeyHash)

	balance := 0
	for _, out := range unspentOutputs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (client *Client) CreateWallet() {
	walletRepo := NewWalletRepo()
	wallet := walletRepo.CreateNewWallet()

	fmt.Printf("Your new address: %s", wallet.GetAddress())
}

func (client *Client) ListAddresses() {
	walletRepo := NewWalletRepo()
	addresses := walletRepo.GetAddresses()

	for _, adr := range addresses {
		fmt.Printf("%s\n", adr)
	}
}

func (client *Client) ReindexUTXO() {
	client.utxoSet.Reindex()
}

func (client *Client) Send(from string, to string, amount int) {
	if !ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	tx, err := NewUTXOTransaction(from, to, amount, client.blockchain)
	if err != nil {
		fmt.Printf("Failed to send: %v\n", err)
		os.Exit(1)
	}
	cbtx := NewCoinbaseTX(from, "")

	block, err := client.blockchain.MineBlock([]*Transaction{cbtx, tx})
	if err != nil {
		fmt.Printf("Failed to send: %v\n", err)
		os.Exit(1)
	}
	client.blockchain.AddBlock(block)
	client.utxoSet.Update(block)
}
