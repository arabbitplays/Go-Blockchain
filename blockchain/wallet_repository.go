package blockchain

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"log"
	"os"
)

type WalletRepo struct {
	nodeID  string
	Wallets map[string]*Wallet
}

const walletFile = "wallet_%s.dat"

func NewWalletRepo(nodeID string) *WalletRepo {
	var repo WalletRepo
	repo.nodeID = nodeID
	repo.Wallets = make(map[string]*Wallet)

	repo.loadFromFile()

	return &repo
}

func (repo *WalletRepo) GetAddresses() []string {
	var addresses []string
	for address := range repo.Wallets {
		addresses = append(addresses, address)
	}
	return addresses
}

func (repo *WalletRepo) CreateNewWallet() *Wallet {
	wallet := NewWallet()
	address := fmt.Sprintf("%s", wallet.GetAddress())

	repo.Wallets[address] = wallet
	repo.saveToFile()

	return wallet
}

func (repo *WalletRepo) GetWallet(address string) Wallet {
	return *repo.Wallets[address]
}

func (repo *WalletRepo) loadFromFile() error {
	walletFile := fmt.Sprintf(walletFile, repo.nodeID)
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	fileContent, err := os.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	var loadedRepo WalletRepo
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&loadedRepo)
	if err != nil {
		log.Panic(err)
	}

	repo.Wallets = loadedRepo.Wallets

	return nil
}

func (repo WalletRepo) saveToFile() {
	var content bytes.Buffer

	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(repo)
	if err != nil {
		log.Panic(err)
	}

	walletFile := fmt.Sprintf(walletFile, repo.nodeID)
	os.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}

func (repo WalletRepo) Destroy() {
	walletFile := fmt.Sprintf(walletFile, repo.nodeID)
	os.Remove(walletFile)
}
