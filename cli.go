package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"example.com/blockchainApp/blockchain"
)

type CLI struct {
	nodeID string
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
	reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
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
	case "reindexutxo":
		err := reindexUTXOCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
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

	cli.nodeID = os.Getenv("NODE_ID")
	if cli.nodeID == "" {
		cli.nodeID = "oschdi"
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

	if reindexUTXOCmd.Parsed() {
		cli.reindexUTXO()
	}

	if sendCmd.Parsed() {
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}

func (cli *CLI) createBlockchain(address string) {
	client := blockchain.NewClient(cli.nodeID)
	defer client.Close()

	client.CreateBlockchain(address)
	fmt.Println("Done!")
}

func (cli *CLI) getBalance(address string) {
	client := blockchain.NewClient(cli.nodeID)
	defer client.Close()

	balance, err := client.GetBalance(address)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) createWallet() {
	client := blockchain.NewClient(cli.nodeID)
	defer client.Close()

	fmt.Printf("Your new address: %s", client.CreateWallet())
}

func (cli *CLI) listAddresses() {
	client := blockchain.NewClient(cli.nodeID)
	defer client.Close()

	for _, address := range client.ListAddresses() {
		fmt.Printf("%s\n", address)
	}
}

func (cli *CLI) printChain() {
	client := blockchain.NewClient(cli.nodeID)
	defer client.Close()

	for _, block := range client.Blocks() {
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("PoW: %t\n", blockchain.NewProofOfWork(block).Validate())

		for txIdx, tx := range block.Transactions {
			fmt.Printf("Transaction %d:\n", txIdx)
			fmt.Printf("  ID: %x\n", tx.ID)
			for inIdx, in := range tx.Vin {
				fmt.Printf("  Input %d: TxID %x, Vout %d, PubKey %x\n", inIdx, in.TxID, in.Vout, in.PubKey)
			}
			for outIdx, out := range tx.Vout {
				fmt.Printf("  Output %d: Value %d, PubKeyHash %x\n", outIdx, out.Value, out.PubKeyHash)
			}
		}
		fmt.Println()
	}
}

func (cli *CLI) reindexUTXO() {
	client := blockchain.NewClient(cli.nodeID)
	defer client.Close()

	client.ReindexUTXO()
}

func (cli *CLI) send(from string, to string, amount int) {
	client := blockchain.NewClient(cli.nodeID)
	defer client.Close()

	if err := client.Send(from, to, amount); err != nil {
		log.Panic(err)
	}
}
