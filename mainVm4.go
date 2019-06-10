package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
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
	"io"
	_ "net/http/pprof"
	"os"
	"sync"
	"time"
	"flag"
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

type ReadTask struct {
	implTask
	blockHeight   uint32
	dataBytes []byte
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

	currentBlockHeight := ledgerStore.GetCurrentBlockHeight()


	modkDBPath2 := fmt.Sprintf("%s%s%s", dbDir, string(os.PathSeparator), "states"+"mockdb2")
	levelDB2, err := ledgerstore.OpenLevelDB(modkDBPath2)
	if err != nil && err.Error() != "leveldb: not found" {
		fmt.Println("modkDBPath2 err: ", err)
		return
	}
	var wg = new(sync.WaitGroup)
	wg.Add(9)
    for i:=uint32(0);i<9;i++ {
    	go writeFile(i, levelDB2, wg, currentBlockHeight)
	}
	wg.Wait()
    log.Info("finished")
}

func writeFile(offset uint32, levelDB2 *leveldb.DB, wg *sync.WaitGroup,currentBlockHeight uint32) {
	defer wg.Done()

	fileName := fmt.Sprintf("readWriteSetAndBlock%d.txt", offset)
	var startHeight uint32
	var f *os.File
	var writer *bufio.Writer
	var err error
	if checkFileIsExist(fileName) {
		f, err = os.OpenFile(fileName, os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			log.Errorf("OpenFile err: %s\n", err)
			return
		}
		key := make([]byte, 4, 4)
		f.ReadAt(key, 0)
		startHeight = binary.LittleEndian.Uint32(key)
		writer = bufio.NewWriter(f)
	} else {
		f, err = os.Create(fileName)
		if err != nil {
			log.Errorf("Create err: %s\n", err)
			return
		}
		writer = bufio.NewWriter(f)
		startHeight = 0
		serialization.WriteUint32(writer, startHeight)
	}
	fmt.Println("startHeight:", startHeight)
	defer func() {
		writer.Flush()
		f.Sync()
		f.Close()
	}()

	fmt.Printf("startHeight: %d, endHeight: %d\n", startHeight+500000*offset, startHeight+500000*offset+500000)

	for i:=uint32(startHeight+500000*offset);i<currentBlockHeight && i<startHeight+500000*offset+500000;i++ {
		key := make([]byte, 4, 4)
		binary.LittleEndian.PutUint32(key[:], i)
		val, err := levelDB2.Get(key, nil)
		if err != nil || val == nil{
			panic(fmt.Errorf("err: %s, height:%d", err, i))
			return
		}
		err = serialization.WriteVarBytes(writer, val)
		if err != nil {
			log.Error("WriteVarBytes err: %s, i:%d", err, i)
			panic("")
		}
		if i%100000 == 0 {
			log.Infof("height: %d", i)
		}
		f.WriteAt(key,0)
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


	ledgerStore, err := ledgerstore.NewLedgerStore(dbDir, 3000000)
	if err != nil {
		log.Errorf("NewLedgerStore err: %s", err)
		return
	}
	initLedgerStore(ledgerStore)

	fileName := "readWriteSetAndBlock.txt"
	var f *os.File
	f, err = os.OpenFile(fileName, os.O_APPEND, 0666)
	if err != nil {
		log.Errorf("OpenFile err: %s\n", err)
		return
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	key := make([]byte, 4,4)
	reader.Read(key)
	currentBlockHeight := binary.LittleEndian.Uint32(key)
	fmt.Println("currentBlockHeight:", currentBlockHeight)

	for i:=uint32(0);i<blockHeight;i++ {
		_, err = serialization.ReadVarBytes(reader)
		if err != nil {
			panic(fmt.Errorf("err: %s\n", err))
		}
	}
	dataBytes, err := serialization.ReadVarBytes(reader)
	if err != nil {
		panic(fmt.Errorf("err: %s\n", err))
	}
	executeInfo, err := getExecuteInfoByHeight(dataBytes)
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

func checkAllBlock(startHeight uint32) {

	ledgerstore.MOCKDBSTORE = false

	dbDir := "./Chain/ontology"

	ledgerStore, err := ledgerstore.NewLedgerStore(dbDir, 3000000)
	if err != nil {
		log.Errorf("NewLedgerStore err: %s\n", err)
		return
	}
	initLedgerStore(ledgerStore)

	start := time.Now()

	currentBlockHeight := ledgerStore.GetCurrentBlockHeight()


	fmt.Println("currentBlockHeight:", currentBlockHeight)

	//one goruntine read
	fmt.Println("start: ", start)
	var wg = new(sync.WaitGroup)

	wg.Add(18)

	ch := make(chan Task, 100)
	for i:=uint32(0);i<9;i++ {
		go readFile(i,currentBlockHeight, wg, ch)
	}

	for i:=uint32(0);i<9;i++ {
		go executeCh(i, ch, ledgerStore, wg)
	}
	wg.Wait()

	fmt.Println("checkAllBlock Current BlockHeight: ", currentBlockHeight)
	fmt.Println("start: ", start)
	fmt.Println("end: ", time.Now())
}
func readFile(offset uint32,currentBlockHeight uint32, wg *sync.WaitGroup,ch chan<- Task) error {
	defer wg.Done()
	fileName := fmt.Sprintf("readWriteSetAndBlock%d.txt", offset)
	var f *os.File
	var err error
	f, err = os.OpenFile(fileName, os.O_APPEND, 0666)
	if err != nil {
		log.Errorf("OpenFile err: %s\n", err)
		return fmt.Errorf("OpenFile err: %s\n", err)
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	key := make([]byte, 4,4)
	reader.Read(key)
	currentBlockHeight = binary.LittleEndian.Uint32(key)

    fmt.Printf("startHeight: %d, endHeight: %d\n", offset*500000, currentBlockHeight)

	for i:=offset*500000; i<currentBlockHeight;i++  {
		dataBytes, err := serialization.ReadVarBytes(reader)
		if err != nil || io.EOF == err {
			ch <- &FinishedTask{}
			fmt.Printf("ReadString err: %s, height: %d, offset: %d\n", err, i, reader.Size())
			return fmt.Errorf("ReadString err: %s, height: %d, offset: %d\n", err, i, reader.Size())
		}
		ch <- &ReadTask{
			blockHeight:i,
			dataBytes:dataBytes,
		}
	}
	ch <- &FinishedTask{}
	fmt.Printf("readFile finished, offset: %d\n", offset)
	return nil
}

func executeCh(offset uint32, ch <- chan Task, ledgerStore *ledgerstore.LedgerStoreImp, wg *sync.WaitGroup) error {
	defer wg.Done()
	for {
		data,ok := <-ch
		if !ok {
			return nil
		}
		switch t:=data.(type) {
		case *ReadTask:
			executeInfo, err := getExecuteInfoByHeight(t.dataBytes)
			if err != nil {
				return err
			}
			execute(executeInfo, ledgerStore)
		case *FinishedTask:
			fmt.Printf("executeCh finished, offset: %d\n", offset)
			return nil
		}
	}
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

func sendExecuteInfoToCh(ch chan<- Task, readChan <-chan Task, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		task, ok:=  <- readChan
		if !ok {
			ch <- &FinishedTask{}
			return
		}
		switch t:=task.(type){
		case *ReadTask:
			executeInfo, err := getExecuteInfoByHeight(t.dataBytes)
			if err != nil {
				fmt.Printf("databytes: %x\n", t.dataBytes)
				panic(fmt.Errorf("getExecuteInfoByHeight err:%s, height: %d", err, t.blockHeight))
			}
			if executeInfo == nil {
				ch <- &FinishedTask{}
				return
			} else {
				ch <- &ExecuteTask{
					executeInfo: executeInfo,
				}
			}
		case *FinishedTask:
			ch <- &FinishedTask{}
			return
		}
	}
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
	if executeInfo.Height % 100000 == 0 {
		fmt.Println("execute blockHeight: ", executeInfo.Height)
	}
	if executeInfo.Height %1000000 == 0 {
		fmt.Println("time: ", time.Now())
	}

	//fmt.Fprintf(os.Stderr, "diff hash at height:%d, hash:%x\n", block.Header.Height, writeSet.Hash())
	//
	//fmt.Fprintf(os.Stderr, "diff hash at height:%d, hash:%x\n", block.Header.Height, executeInfo.WriteSet.Hash())
}

func getExecuteInfoByHeight(dataBytes []byte) (*ExecuteInfo, error) {

	//get gasTable

	source := common.NewZeroCopySource(dataBytes)
	readWriteSetBytes, _, irregular, eof := source.NextVarBytes()
	if eof || irregular {
		return nil, fmt.Errorf("readWriteSetBytes, eof or irregular error")
	}
	blockBytes, _, irregular, eof := source.NextVarBytes()
	if eof || irregular {
		return nil, fmt.Errorf("blockBytes eof or irregular error")
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
	return &ExecuteInfo{Height: block.Header.Height, ReadSet: readSetDB, WriteSet: writeSetDB, GasTable: m, BlockInfo: block}, nil
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

func handleTransaction(ledgerStore *ledgerstore.LedgerStoreImp, overlay *overlaydb.OverlayDB,gasTable map[string]uint64, cache *storage.CacheDB, block *types.Block, tx *types.Transaction) (*event.ExecuteNotify, error) {
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
		err := stateStore.HandleInvokeTransaction(ledgerStore, overlay,gasTable, cache, tx, block, notify)
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
