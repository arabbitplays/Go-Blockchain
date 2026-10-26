package main

import (
	"example.com/blockchainApp/blockchain"
)

func main() {
	bc := blockchain.NewBlockchain()
	defer bc.Destroy()

	cli := blockchain.NewCLI(bc)
	cli.Run()
}
