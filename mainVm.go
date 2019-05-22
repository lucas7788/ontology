package main

import (
	"fmt"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/ledger"
)

func main() {
	initLedger()

	hash := ledger.DefLedger.GetBlockHash(0)
	fmt.Println("hash:", hash)
	block, _ := ledger.DefLedger.GetBlockByHash(hash)
	ledger.DefLedger.GetHeaderByHeight(0)
	ledger.DefLedger.ExecuteBlock(block)
}
func initLedger() {
	var err error
	dbDir := "./Chain/ontology"
	ledger.DefLedger, err = ledger.NewLedger(dbDir, 3000000)
	if err != nil {
		return
	}

	bookKeepers, err := config.DefConfig.GetBookkeepers()
	if err != nil {
		return
	}
	genesisConfig := config.DefConfig.Genesis
	genesisBlock, err := genesis.BuildGenesisBlock(bookKeepers, genesisConfig)
	if err != nil {
		return
	}
	err = ledger.DefLedger.Init(bookKeepers, genesisBlock)
	if err != nil {
		return
	}
}
