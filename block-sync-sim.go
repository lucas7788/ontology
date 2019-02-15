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

func syncHash(conn net.Conn, info *BlockHashInfo) {
	for {
		hash, _ := info.GetHash(info.GetKnownHeight())
		req := msgpack.NewHeadersReq(hash)
		WriteMessage(conn, req)
		time.Sleep(time.Second)
	}
}

func syncBlock(conn net.Conn, info *BlockHashInfo) {
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

func readLoop(conn net.Conn, info *BlockHashInfo) {
	for {
		msg, _, err := types.ReadMessage(conn)
		checkerr(err)
		switch message := msg.(type) {
		case *types.BlkHeader:
			//fmt.Printf("received headers. len:%d\n", len(message.BlkHdr))
			for _, header := range message.BlkHdr {
				info.SetHash(header.Height, header.Hash())
			}
		case *types.Block:
			//fmt.Printf("received block. height:%d\n", message.Blk.Header.Height)
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

	go heartBeatService(conn)
	go syncHash(conn, &info)
	go syncBlock(conn, &info)
	go readLoop(conn, &info)

	waitToExit()
}
