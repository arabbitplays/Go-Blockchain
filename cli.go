package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"example.com/blockchainApp/blockchain"
)

type CLI struct {
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  listaddresses - Lists all addresses from the wallet file")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  reindexutxo - Rebuilds the UTXO set")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT -mine - Send AMOUNT of coins from FROM address to TO. Mine on the same node, when -mine is set.")
	fmt.Println("  startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env. var. -miner enables mining")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) Run() {
	cli.validateArgs()

	// nodeID := os.Getenv("NODE_ID")
	// if nodeID == "" {
	// 	fmt.Printf("NODE_ID env. var is not set!")
	// 	os.Exit(1)
	// }

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	// reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	// startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	// sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")
	// startNodeMiner := startNodeCmd.String("miner", "", "Enable mining mode and send reward to ADDRESS")

	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	// case "reindexutxo":
	// 	err := reindexUTXOCmd.Parse(os.Args[2:])
	// 	if err != nil {
	// 		log.Panic(err)
	// 	}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	// case "startnode":
	// 	err := startNodeCmd.Parse(os.Args[2:])
	// 	if err != nil {
	// 		log.Panic(err)
	// 	}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if createBlockchainCmd.Parsed() {
		cli.createBlockchain(*createBlockchainAddress)
	}

	if getBalanceCmd.Parsed() {
		cli.getBalance(*getBalanceAddress)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}

func (cli *CLI) createBlockchain(address string) {
	bc := blockchain.Create(address)
	defer bc.Close()

	fmt.Println("Done!")
}

func (cli *CLI) getBalance(address string) {
	if !blockchain.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}

	bc := blockchain.Open()
	defer bc.Close()

	unspentOutputs := bc.FindUnspentOutputs(address)

	balance := 0
	for _, out := range unspentOutputs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) createWallet() {
	walletRepo := blockchain.NewWalletRepo()
	wallet := walletRepo.CreateNewWallet()

	fmt.Printf("Your new address: %s", wallet.GetAddress())
}

func (cli *CLI) listAddresses() {
	walletRepo := blockchain.NewWalletRepo()
	addresses := walletRepo.GetAddresses()

	for _, adr := range addresses {
		fmt.Printf("%s\n", adr)
	}
}

func (cli *CLI) send(from string, to string, amount int) {
	if !blockchain.ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !blockchain.ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := blockchain.Open()
	defer bc.Close()

	tx, err := blockchain.NewUTXOTransaction(from, to, amount, bc)
	if err != nil {
		fmt.Printf("Failed to send: %v\n", err)
		os.Exit(1)
	}

	block, err := bc.MineBlock([]*blockchain.Transaction{tx})
	if err != nil {
		fmt.Printf("Failed to send: %v\n", err)
		os.Exit(1)
	}
	bc.AddBlock(block)
}

func (cli *CLI) printChain() {
	// bc := blockchain.NewBlockchain()
	// defer bc.Destroy()
	//
	// bci := bc.Iterator()
	// for {
	// 	block := bci.Next()
	//
	// 	fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
	// 	//fmt.Printf("Data: %s\n", block.Data)
	// 	fmt.Printf("Hash: %x\n", block.Hash)
	// 	pow := blockchain.NewProofOfWork(block)
	//
	// 	for txIdx, tx := range(block.Transactions) {
	// 		fmt.Printf("Transaction %d:\n", txIdx)
	// 		for _, in := range(tx.Vin) {
	// 			fmt.Printf("In: %s\n", in.ScriptSig)
	// 		}
	//
	// 		for _, out := range(tx.Vout) {
	// 			fmt.Printf("Out: %d %s\n", out.Value, out.ScriptPubKey)
	// 		}
	// 		fmt.Println()
	// 	}
	//
	// 	fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
	// 	fmt.Println()
	//
	// 	if len(block.PrevBlockHash) == 0 {
	// 		break
	// 	}
	// }
}
