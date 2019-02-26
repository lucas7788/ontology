package main

import (
	"errors"
	"fmt"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/constants"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	"github.com/ontio/ontology/p2pserver/message/types"
)

type BlockHashInfo struct {
	hashes     map[uint32]common.Uint256
	currHeight uint32
	lock       sync.RWMutex
}

func NewBlockHashInfo() BlockHashInfo {
	hashes := make(map[uint32]common.Uint256)
	hashes[0] = common.UINT256_EMPTY
	return BlockHashInfo{
		hashes:     hashes,
		currHeight: 0,
	}
}

func (self *BlockHashInfo) GetKnownHeight() uint32 {
	return self.currHeight
}

func (self *BlockHashInfo) GetHash(height uint32) (common.Uint256, error) {
	self.lock.RLock()
	defer self.lock.RUnlock()
	hash, ok := self.hashes[height]
	if ok == false {
		return hash, errors.New("hash not found")
	}

	return hash, nil
}

func (self *BlockHashInfo) SetHash(height uint32, hash common.Uint256) {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.hashes[height] = hash
	if height == self.currHeight+1 {
		self.currHeight += 1
	}
}

func heartBeatService(conn net.Conn) {
	var periodTime uint
	periodTime = 3
	t := time.NewTicker(time.Second * (time.Duration(periodTime)))
	for {
		select {
		case <-t.C:
			ping := msgpack.NewPingMsg(0)
			WriteMessage(conn, ping)
		}
	}
}

func waitToExit() {
	exit := make(chan bool, 0)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for sig := range sc {
			log.Infof("Ontology received exit signal:%v.", sig.String())
			close(exit)
			break
		}
	}()
	<-exit
}

func WriteMessage(conn net.Conn, msg types.Message) {
	sink := common.NewZeroCopySink(nil)
	err := types.WriteMessage(sink, msg)
	checkerr(err)
	_, err = conn.Write(sink.Bytes())
	checkerr(err)
}

func checkerr(err error) {
	if err != nil {
		panic(err)
	}
}

func (self *BlockSync) syncHash() {
	conn := self.conn
	info := self.info
	for {
		hash, _ := info.GetHash(info.GetKnownHeight())
		req := msgpack.NewHeadersReq(hash)
		WriteMessage(conn, req)
		time.Sleep(time.Second)
	}
}

func (self *BlockSync) syncBlock() {
	conn := self.conn
	info := self.info
	numBlocks := config.GetDebugOption().NumBlockPerSecond
	for {
		if info.GetKnownHeight() != 0 {
			for i := 0; i < numBlocks; i++ {
				pick := rand.Uint32() % info.GetKnownHeight()
				hash, _ := info.GetHash(pick)
				req := msgpack.NewBlkDataReq(hash)
				WriteMessage(conn, req)
			}
		}
		time.Sleep(time.Second)
	}
}

type BlockSync struct {
	info       *BlockHashInfo
	conn       net.Conn
	blockCount int64
	totalBytes int64
	lock       sync.Mutex
}

func (self *BlockSync) IncBlockCount(size int64) (int64, int64) {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.blockCount += 1
	self.totalBytes += size
	return self.blockCount, self.totalBytes
}

func (self *BlockSync) readLoop() {
	conn := self.conn
	info := self.info
	for {
		msg, size, err := types.ReadMessage(conn)
		checkerr(err)
		switch message := msg.(type) {
		case *types.BlkHeader:
			//fmt.Printf("received headers. len:%d\n", len(message.BlkHdr))
			for _, header := range message.BlkHdr {
				info.SetHash(header.Height, header.Hash())
			}
		case *types.Block:
			count, tb := self.IncBlockCount(int64(size))
			if count%1000 == 0 {
				fmt.Printf("received block:%d, size:%d\n", count, tb)
			}
		}
	}
}

func main() {
	config.DefConfig.P2PNode.NetworkMagic = constants.NETWORK_MAGIC_MAINNET
	info := NewBlockHashInfo()
	conn, err := net.Dial("tcp", "127.0.0.1:20338")
	checkerr(err)

	var version types.Version
	version.P = types.VersionPayload{
		Nonce:       rand.Uint64(),
		IsConsensus: false,
		TimeStamp:   time.Now().UnixNano(),
	}
	WriteMessage(conn, &version)

	msg, _, err := types.ReadMessage(conn)
	checkerr(err)
	fmt.Printf("receive msg:%v", msg)
	WriteMessage(conn, msgpack.NewVerAck(false))

	blockSync := BlockSync{
		info: &info,
		conn: conn,
	}

	go heartBeatService(conn)
	go blockSync.syncHash()
	go blockSync.syncBlock()
	go blockSync.readLoop()

	waitToExit()
}
