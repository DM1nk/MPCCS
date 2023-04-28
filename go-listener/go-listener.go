package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	client, err := ethclient.Dial("wss://goerli.infura.io/ws/v3/ed3535da156a4e16b540b78d07126601")
	if err != nil {
		log.Fatal(err)
	}

	contractAddress := common.HexToAddress("0x07865c6E87B9F70255377e024ace6630C1Eaa37F")
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}

	logs := make(chan types.Log)

	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case vLog := <-logs:
			fmt.Println(vLog) // pointer to log event
			fmt.Println(vLog.BlockNumber, vLog.TxHash.Hex(), vLog.Topics, vLog.Data)
		}
	}
}
