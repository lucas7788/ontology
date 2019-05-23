package main

import (
	"bytes"
	"fmt"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/core/store/ledgerstore"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	"github.com/ontio/ontology/smartcontract/storage"
	"math"
	"os"
	"strconv"
)

var Block_Height = uint32(505)

func main() {
	ledgerstore.MOCKDBSTORE = false

	dbDir := "./Chain/ontology"
	modkDBPath := fmt.Sprintf("%s%s%s", dbDir, string(os.PathSeparator), "states"+"mockdb")
	store, err := leveldbstore.NewLevelDBStore(modkDBPath)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	ledgerStore, err := ledgerstore.NewLedgerStore(dbDir, 3000000)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	initLedgerStore(ledgerStore)

	mockDBStore := ledgerstore.NewMockDBStore(store)
	overlay := ledgerstore.NewOverlayDB(Block_Height, mockDBStore)

	hash := ledgerStore.GetBlockHash(Block_Height)
	block, _ := ledgerStore.GetBlockByHash(hash)

	if block.Header.Height != 0 {
		config := &smartcontract.Config{
			Time:   block.Header.Timestamp,
			Height: block.Header.Height,
			Tx:     &types.Transaction{},
		}

		err = refreshGlobalParam(config, storage.NewCacheDB(ledgerStore.GetStateStore().NewOverlayDB()), ledgerStore)
		if err != nil {
			fmt.Println("err:", err)
			return
		}
	}

	cache := storage.NewCacheDB(overlay)
	for _, tx := range block.Transactions {
		cache.Reset()
		_, e := handleTransaction(ledgerStore, overlay, cache, block, tx)
		if e != nil {
			err = e
			fmt.Println("err:", err)
			return
		}
	}

	writeSet := overlay.GetWriteSet()
	fmt.Fprintf(os.Stderr, "diff hash at height:%d, hash:%x\n", block.Header.Height, writeSet.Hash())

	memdbAfterExecute := ledgerstore.NewOverlayDBAfterExecutor(Block_Height, mockDBStore)
	fmt.Fprintf(os.Stderr, "diff hash at height:%d, hash:%x\n", block.Header.Height, memdbAfterExecute.Hash())

}

func initLedgerStore(ledgerStore *ledgerstore.LedgerStoreImp) {
	bookKeepers, err := config.DefConfig.GetBookkeepers()
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	genesisConfig := config.DefConfig.Genesis
	genesisBlock, err := genesis.BuildGenesisBlock(bookKeepers, genesisConfig)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	ledgerStore.InitLedgerStoreWithGenesisBlock(genesisBlock, bookKeepers)
}

func handleTransaction(ledgerStore *ledgerstore.LedgerStoreImp, overlay *overlaydb.OverlayDB, cache *storage.CacheDB, block *types.Block, tx *types.Transaction) (*event.ExecuteNotify, error) {
	txHash := tx.Hash()
	notify := &event.ExecuteNotify{TxHash: txHash, State: event.CONTRACT_STATE_FAIL}
	stateStore := ledgerstore.StateStore{}
	switch tx.TxType {
	case types.Deploy:
		err := stateStore.HandleDeployTransaction(ledgerStore, overlay, cache, tx, block, notify)
		if overlay.Error() != nil {
			return nil, fmt.Errorf("HandleDeployTransaction tx %s error %s", txHash.ToHexString(), overlay.Error())
		}
		if err != nil {
			log.Debugf("HandleDeployTransaction tx %s error %s", txHash.ToHexString(), err)
		}
	case types.Invoke:
		err := stateStore.HandleInvokeTransaction(ledgerStore, overlay, cache, tx, block, notify)
		if overlay.Error() != nil {
			return nil, fmt.Errorf("HandleInvokeTransaction tx %s error %s", txHash.ToHexString(), overlay.Error())
		}
		if err != nil {
			log.Debugf("HandleInvokeTransaction tx %s error %s", txHash.ToHexString(), err)
		}
	}

	return notify, nil
}

func refreshGlobalParam(config *smartcontract.Config, cache *storage.CacheDB, store store.LedgerStore) error {
	bf := new(bytes.Buffer)
	if err := utils.WriteVarUint(bf, uint64(len(neovm.GAS_TABLE_KEYS))); err != nil {
		return fmt.Errorf("write gas_table_keys length error:%s", err)
	}
	for _, value := range neovm.GAS_TABLE_KEYS {
		if err := serialization.WriteString(bf, value); err != nil {
			return fmt.Errorf("serialize param name error:%s", value)
		}
	}

	sc := smartcontract.SmartContract{
		Config:  config,
		CacheDB: cache,
		Store:   store,
		Gas:     math.MaxUint64,
	}

	service, _ := sc.NewNativeService()
	result, err := service.NativeCall(utils.ParamContractAddress, "getGlobalParam", bf.Bytes())
	if err != nil {
		return err
	}
	params := new(global_params.Params)
	if err := params.Deserialize(bytes.NewBuffer(result.([]byte))); err != nil {
		return fmt.Errorf("deserialize global params error:%s", err)
	}
	neovm.GAS_TABLE.Range(func(key, value interface{}) bool {
		n, ps := params.GetParam(key.(string))
		if n != -1 && ps.Value != "" {
			pu, err := strconv.ParseUint(ps.Value, 10, 64)
			if err != nil {
				log.Errorf("[refreshGlobalParam] failed to parse uint %v\n", ps.Value)
			} else {
				neovm.GAS_TABLE.Store(key, pu)

			}
		}
		return true
	})
	return nil
}
