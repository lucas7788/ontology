package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/store/ledgerstore"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	"github.com/ontio/ontology/smartcontract/storage"
	"github.com/syndtr/goleveldb/leveldb"
	_ "net/http/pprof"
	"os"
	"sync"
	"time"
)

type ExecuteInfo struct {
	Height    uint32
	ReadSet   *overlaydb.MemDB
	WriteSet  *overlaydb.MemDB
	GasTable  map[string]uint64
	BlockInfo *types.Block
}

type Task interface {
	ImplementTask()
}

type implTask struct{}

func (self implTask) ImplementTask() {}

type FinishedTask struct {
	implTask
}

type ExecuteTask struct {
	implTask
	executeInfo *ExecuteInfo
}

func main() {
	//go func() {
	//	http.ListenAndServe("localhost:10000", nil)
	//}()
	runMode := flag.String("name", "checkall", "run mode")
	blockHeight := flag.Int("blockHeight", 0, "run mode")
	flag.Parse()
	if *runMode == "checkone" {
		fmt.Println("checkone")
		checkOneBlock()
	} else if *runMode == "updatedata" {
		fmt.Println("saveBlockToReadWriteSet")
		//1989103  2050774
		saveBlockToReadWriteSet()
	} else if *runMode == "changeDbToFile" {
		changeDbToFile()
	} else {
		fmt.Println("checkAllBlock")
		checkAllBlock(uint32(*blockHeight))
	}
}

func changeDbToFile() {
	dbDir := "./Chain/ontology"

	ledgerStore, err := ledgerstore.NewLedgerStore(dbDir, 3000000)
	if err != nil {
		fmt.Println("NewLedgerStore err:", err)
		return
	}
	initLedgerStore(ledgerStore)

	modkDBPath2 := fmt.Sprintf("%s%s%s", dbDir, string(os.PathSeparator), "states"+"mockdb2")
	levelDB2, err := ledgerstore.OpenLevelDB(modkDBPath2)
	if err != nil && err.Error() != "leveldb: not found" {
		fmt.Println("modkDBPath2 err: ", err)
		return
	}
	currentHeightBytes, err := levelDB2.Get([]byte("currentHeight"), nil)
	if err != nil && err.Error() != "leveldb: not found" {
		fmt.Println("Get currentHeight err:", err)
		return
	}
	currentHeight := uint32(0)
	if currentHeightBytes != nil {
		currentHeight = binary.LittleEndian.Uint32(currentHeightBytes)
		fmt.Println("&&& currentHeight:", currentHeight)
	}
	fileName := "readWriteSetAndBlock.txt"
	var f *os.File
	if checkFileIsExist(fileName) {
		f, err = os.OpenFile(fileName, os.O_APPEND, 0666)
		if err != nil {
			log.Errorf("OpenFile err: %s\n", err)
			return
		}
	} else {
		f, err = os.Create(fileName)
		if err != nil {
			log.Errorf("Create err: %s\n", err)
			return
		}
	}
	defer f.Close()
	writer := bufio.NewWriter(f)
	defer writer.Flush()
	currBlockHeight := ledgerStore.GetCurrentBlockHeight()
	for i := uint32(1294239); i < currBlockHeight; i++ {
		key := make([]byte, 4, 4)
		binary.LittleEndian.PutUint32(key[:], i)
		val, err := levelDB2.Get(key, nil)
		if err != nil {
			panic(fmt.Errorf("err: %s, height:%d", err, i))
			return
		}
		serialization.WriteVarBytes(writer, val)
		if i%10000 == 0 {
			log.Infof("height: %d\n", i)
		}
	}
	log.Infof("finished\n")
}

func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func saveBlockToReadWriteSet() {
	dbDir := "./Chain/ontology"

	ledgerstore.MOCKDBSTORE = false

	modkDBPath := fmt.Sprintf("%s%s%s", dbDir, string(os.PathSeparator), "states"+"mockdb")
	levelDB, err := ledgerstore.OpenLevelDB(modkDBPath)
	if err != nil {
		fmt.Println("modkDBPath err: ", err)
		return
	}

	modkDBPath2 := fmt.Sprintf("%s%s%s", dbDir, string(os.PathSeparator), "states"+"mockdb2")
	levelDB2, err := ledgerstore.OpenLevelDB(modkDBPath2)
	if err != nil && err.Error() != "leveldb: not found" {
		fmt.Println("modkDBPath2 err: ", err)
		return
	}

	ledgerStore, err := ledgerstore.NewLedgerStore(dbDir, 3000000)
	if err != nil {
		fmt.Println("NewLedgerStore err:", err)
		return
	}
	initLedgerStore(ledgerStore)

	currentHeight := uint32(0)
	if levelDB2 != nil {
		currentHeightBytes, err := levelDB2.Get([]byte("currentHeight"), nil)
		if err != nil && err.Error() != "leveldb: not found" {
			fmt.Println("Get currentHeight err:", err)
			return
		}
		if currentHeightBytes != nil {
			currentHeight = binary.LittleEndian.Uint32(currentHeightBytes)
			fmt.Println("&&& currentHeight:", currentHeight)
		}
	}
	//1294239
	//currentHeight = 1000000

	currentBlockHeight := ledgerStore.GetCurrentBlockHeight()
	var wg = new(sync.WaitGroup)
	wg.Add(10)
	for i := uint32(0); i < 10; i++ {
		go updateData(levelDB, levelDB2, ledgerStore, i, currentBlockHeight, wg, currentHeight)
	}
	wg.Wait()
	fmt.Println("currentBlockHeight:", currentBlockHeight)
	fmt.Println("end")
}

func updateData(levelDB, levelDB2 *leveldb.DB, ledgerStore *ledgerstore.LedgerStoreImp, offset uint32, currentBlockHeight uint32, wg *sync.WaitGroup, currentHeight uint32) {
	defer wg.Done()
	sink := common.NewZeroCopySink(nil)
	blockSink := common.NewZeroCopySink(nil)
	for i := uint32(currentHeight / 10); 10*i+offset < currentBlockHeight; i++ {

		//read WriteSet
		key := make([]byte, 4, 4)
		binary.LittleEndian.PutUint32(key[:], 10*i+offset)

		v, err := levelDB2.Get(key, nil)
		if err != nil {
			fmt.Errorf("levelDB2.Get: %s, height: %d", err, 10*i+offset)
		}
		if v != nil {
			fmt.Println("has value height:", 10*i+offset)
			continue
		}

		dataBytes, err := levelDB.Get(key, nil)
		if err != nil {
			fmt.Printf("err:%s, height:%d", err, 10*i+offset)
			panic(10*i + offset)
			return
		}
		sink.Reset()
		sink.WriteVarBytes(dataBytes)
		blockHash := ledgerStore.GetBlockHash(10*i + offset)

		block, err := ledgerStore.GetBlockByHash(blockHash)
		if err != nil {
			return
		}
		blockSink.Reset()
		block.Serialization(blockSink)
		sink.WriteVarBytes(blockSink.Bytes())
		levelDB2.Put(key, sink.Bytes(), nil)
		currentHeight = 10*i + offset
		height := make([]byte, 4, 4)
		binary.LittleEndian.PutUint32(height[:], currentHeight)
		levelDB2.Put([]byte("currentHeight"), height, nil)
		fmt.Println("updateData currentHeight:", currentHeight)
	}
}

func checkOneBlock() {
	blockHeight := uint32(534300)
	blockHeight = uint32(1294201)
	blockHeight = uint32(80003)
	blockHeight = uint32(0)
	ledgerstore.MOCKDBSTORE = false

	dbDir := "./Chain/ontology"

	modkDBPath := fmt.Sprintf("%s%s%s", dbDir, string(os.PathSeparator), "states"+"mockdb2")
	levelDB, err := ledgerstore.OpenLevelDB(modkDBPath)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}

	ledgerStore, err := ledgerstore.NewLedgerStore(dbDir, 3000000)
	initLedgerStore(ledgerStore)

	executeInfo, err := getExecuteInfoByHeight(blockHeight, levelDB, ledgerStore)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	if executeInfo == nil {
		fmt.Println("executeInfo is nil:")
		return
	}
	execute(executeInfo, ledgerStore)
}

func checkAllBlock(blockHeight uint32) {
	var wg = new(sync.WaitGroup)

	ledgerstore.MOCKDBSTORE = false

	dbDir := "./Chain/ontology"

	modkDBPath := fmt.Sprintf("%s%s%s", dbDir, string(os.PathSeparator), "states"+"mockdb2")
	levelDB, err := ledgerstore.OpenLevelDB(modkDBPath)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}

	ledgerStore, err := ledgerstore.NewLedgerStore(dbDir, 3000000)
	initLedgerStore(ledgerStore)

	start := time.Now()

	ch := make(chan Task, 100)
	currentBlockHeight := ledgerStore.GetCurrentBlockHeight()
	wg.Add(8)
	for i := uint32(0); i < 4; i++ {
		go sendExecuteInfoToCh(ch, i, currentBlockHeight, levelDB, wg, ledgerStore, blockHeight)
	}

	for i := 0; i < 4; i++ {
		go handleExecuteInfo(ch, ledgerStore, wg)
	}

	//log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
	wg.Wait()
	fmt.Println("checkAllBlock Current BlockHeight: ", ledgerStore.GetCurrentBlockHeight())
	fmt.Println("start: ", start)
	fmt.Println("end: ", time.Now())
}

func handleExecuteInfo(ch <-chan Task, ledgerStore *ledgerstore.LedgerStoreImp, wg *sync.WaitGroup) {
	for {
		task, ok := <-ch
		if !ok {
			wg.Done()
			return
		}
		switch t := task.(type) {
		case *FinishedTask:
			wg.Done()
			return
		case *ExecuteTask:
			execute(t.executeInfo, ledgerStore)
		}
	}
}

func sendExecuteInfoToCh(ch chan<- Task, offset uint32, currentBlockHeight uint32, levelDB *leveldb.DB, wg *sync.WaitGroup, ledgerStore *ledgerstore.LedgerStoreImp, startHeight uint32) {
	defer wg.Done()
	for i := uint32(startHeight / 4); 4*i+offset < currentBlockHeight; i++ {
		executeInfo, err := getExecuteInfoByHeight(4*i+offset, levelDB, ledgerStore)
		if err != nil {
			fmt.Println("err:", err)
			panic(fmt.Errorf("getExecuteInfoByHeight err:%s, height:%d", err, 4*i+offset))
			return
		}
		ch <- &ExecuteTask{
			executeInfo: executeInfo,
		}
	}
	ch <- &FinishedTask{}
}

func execute(executeInfo *ExecuteInfo, ledgerStore *ledgerstore.LedgerStoreImp) {

	overlay := overlaydb.NewOverlayDB(ledgerstore.NewMockDBWithMemDB(executeInfo.ReadSet))

	refreshGlobalParam(executeInfo.GasTable)
	cache := storage.NewCacheDB(overlay)
	//overlaydb.IS_SHOW = false
	//neovm.PrintOpcode = false
	//index := 0
	for _, tx := range executeInfo.BlockInfo.Transactions {
		cache.Reset()
		//fmt.Fprintf(os.Stderr, "begin transaction, index:%d\n", index)
		_, e := handleTransaction(ledgerStore, overlay, executeInfo.GasTable, cache, executeInfo.BlockInfo, tx)
		//fmt.Fprintf(os.Stderr, "end transaction, index:%d\n", index)
		//index++
		if e != nil {
			fmt.Println("err:", e)
			return
		}
	}
	//overlaydb.IS_SHOW = false

	writeSet := overlay.GetWriteSet()
	//fmt.Printf("hash:  %x", writeSet.Hash())
	//fmt.Println("*****************")
	//fmt.Println("*****************")
	//fmt.Printf("hash:  %x", executeInfo.WriteSet.Hash())

	if !bytes.Equal(writeSet.Hash(), executeInfo.WriteSet.Hash()) {

		//writeSet.Hash()
		//fmt.Println("**********************")
		//executeInfo.WriteSet.Hash()

		//tempMap := make(map[string]string)
		//writeSet.ForEach(func(key, val []byte) {
		//	tempMap[common.ToHexString(key)] = common.ToHexString(val)
		//})
		//executeInfo.WriteSet.ForEach(func(key, val []byte) {
		//	if tempMap[common.ToHexString(key)] != common.ToHexString(val) {
		//		fmt.Printf("key:%x, value: %x\n", key, val)
		//	}
		//})

		fmt.Printf("blockheight:%d, writeSet.Hash:%x, executeInfo.WriteSet.Hash:%x\n", executeInfo.Height, writeSet.Hash(), executeInfo.WriteSet.Hash())
		panic(executeInfo.Height)
	}
	if executeInfo.Height%10000 == 0 {
		fmt.Println("execute blockHeight: ", executeInfo.Height)
	}

	//fmt.Fprintf(os.Stderr, "diff hash at height:%d, hash:%x\n", block.Header.Height, writeSet.Hash())
	//
	//fmt.Fprintf(os.Stderr, "diff hash at height:%d, hash:%x\n", block.Header.Height, executeInfo.WriteSet.Hash())
}

func getExecuteInfoByHeight(height uint32, levelDB *leveldb.DB, ledgerStore *ledgerstore.LedgerStoreImp) (*ExecuteInfo, error) {
	//get gasTable
	key := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(key[:], height)

	dataBytes, err := levelDB.Get(key, nil)

	if err != nil {
		return nil, fmt.Errorf("getExecuteInfoByHeight get databytes error: %s， height：%d", err, height)
	}
	source := common.NewZeroCopySource(dataBytes)
	readWriteSetBytes, _, irregular, eof := source.NextVarBytes()
	if eof || irregular {
		return nil, fmt.Errorf("eof or irregular error, height: %d", height)
	}
	blockBytes, _, irregular, eof := source.NextVarBytes()
	if eof || irregular {
		return nil, fmt.Errorf("eof or irregular error,height: %d", height)
	}
	source = common.NewZeroCopySource(readWriteSetBytes)
	l, eof := source.NextUint32()
	if eof {
		return nil, fmt.Errorf("gastable length is wrong: %d", l)
	}

	m := make(map[string]uint64)
	for i := uint32(0); i < l; i++ {
		key, _, irregular, eof := source.NextVarBytes()
		if irregular || eof {
			return nil, fmt.Errorf("update gastable NextVarBytes error")
		}
		val, eof := source.NextUint64()
		if eof {
			return nil, fmt.Errorf("update gastable NextUint64 error")
		}
		m[string(key)] = val
	}
	//get readSet
	l, eof = source.NextUint32()
	if eof {
		return nil, fmt.Errorf("readset NextUint32 error: %d", l)
	}
	readSetDB := overlaydb.NewMemDB(16*1024, 16)
	for i := uint32(0); i < l; i++ {
		key, _, irregular, eof := source.NextVarBytes()
		if eof || irregular {
			break
		}
		value, _, _, eof := source.NextVarBytes()
		if eof {
			break
		}
		readSetDB.Put(key, value)
	}

	// get writeSet
	l, eof = source.NextUint32()
	if eof {
		return nil, fmt.Errorf("writeset NextUint32 error: %d", l)
	}
	writeSetDB := overlaydb.NewMemDB(16*1024, 16)
	for i := uint32(0); i < l; i++ {
		key, _, irregular, eof := source.NextVarBytes()
		if eof || irregular {
			break
		}
		value, _, _, eof := source.NextVarBytes()
		if eof {
			break
		}
		writeSetDB.Put(key, value)
	}

	block, err := types.BlockFromRawBytes(blockBytes)
	if err != nil {
		return nil, err
	}
	return &ExecuteInfo{Height: height, ReadSet: readSetDB, WriteSet: writeSetDB, GasTable: m, BlockInfo: block}, nil
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

func handleTransaction(ledgerStore *ledgerstore.LedgerStoreImp, overlay *overlaydb.OverlayDB, gasTable map[string]uint64, cache *storage.CacheDB, block *types.Block, tx *types.Transaction) (*event.ExecuteNotify, error) {
	txHash := tx.Hash()
	notify := &event.ExecuteNotify{TxHash: txHash, State: event.CONTRACT_STATE_FAIL}
	stateStore := ledgerstore.StateStore{}
	switch tx.TxType {
	case types.Deploy:
		err := stateStore.HandleDeployTransaction(ledgerStore, overlay, gasTable, cache, tx, block, notify)
		if overlay.Error() != nil {
			return nil, fmt.Errorf("HandleDeployTransaction tx %s error %s", txHash.ToHexString(), overlay.Error())
		}
		if err != nil {
			log.Debugf("HandleDeployTransaction tx %s error %s", txHash.ToHexString(), err)
		}
	case types.Invoke:
		err := stateStore.HandleInvokeTransaction(ledgerStore, overlay, gasTable, cache, tx, block, notify)
		if overlay.Error() != nil {
			return nil, fmt.Errorf("HandleInvokeTransaction tx %s error %s", txHash.ToHexString(), overlay.Error())
		}
		if err != nil {
			log.Debugf("HandleInvokeTransaction tx %s error %s", txHash.ToHexString(), err)
		}
	}

	return notify, nil
}

func refreshGlobalParam(gasTable map[string]uint64) {
	for k, v := range gasTable {
		neovm.GAS_TABLE.Store(k, v)
	}
}
