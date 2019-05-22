package main

import (
	"fmt"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/store/ledgerstore"
	"os"
)

func main() {
	initLedger()
	blockStore, err := ledgerstore.NewBlockStore(fmt.Sprintf("%s%s%s", "./Chain", string(os.PathSeparator), ledgerstore.DBDirBlock), true)
	if err != nil {
		return
	}
	hash,_ := blockStore.GetBlockHash(1)
	fmt.Println("hash:", hash)
	block,_ := ledger.DefLedger.GetBlockByHash(hash)
	ledger.DefLedger.ExecuteBlock(block)
}
func initLedger()  {
	var err error
	dbDir := "./Chain"
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
