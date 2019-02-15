package main

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	msgCommon "github.com/ontio/ontology/p2pserver/common"
	connL "github.com/ontio/ontology/p2pserver/link"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	"github.com/ontio/ontology/p2pserver/message/types"
	ns "github.com/ontio/ontology/p2pserver/net/netserver"
	"os"
	"os/signal"
	osSync "sync"
	"syscall"
	"time"
)


const (
	SYNC_NEXT_BLOCK_TIMES        = 3          //Request times of next height block
	SYNC_NEXT_BLOCKS_HEIGHT      = 2          //for current block height plus next
)

var curHeaderBlockHash = common.UINT256_EMPTY
var currHeaderHeight    = uint32(0)
var headerIndex   map[uint32]common.Uint256
var headerIndexlock  osSync.RWMutex

func hookChan(channel chan *types.MsgPayload, link* connL.Link) {
	for {
		select {
		case data, ok := <-channel:
			if ok {
				msgType := data.Payload.CmdType()
                switch(msgType){
				case msgCommon.VERSION_TYPE:
					log.Info("Receive version type message")
					verAck := msgpack.NewVerAck(false)
					err := link.Tx(verAck)
					if err != nil {
						log.Warn(err)
						return
					}
				case msgCommon.VERACK_TYPE:
					log.Info("Receive version ck type message")
					go heartBeatService(link)
					/*msgHeaderReq := msgpack.NewHeadersReq(curHeaderBlockHash)
					err := link.Tx( msgHeaderReq)
					if err != nil {
						log.Warn("[p2p]failed to send a new headersReq:s", err)
					}*/
					go syncBlockTimer(link)
				case msgCommon.PONG_TYPE:
					pongMsg := data.Payload.(*types.Pong)
					log.Infof("Receive pong message, remote height=%d", pongMsg.Height)

				case msgCommon.HEADERS_TYPE:
					blkHeader := data.Payload.(*types.BlkHeader)
					if len(blkHeader.BlkHdr) > 0 {
						log.Infof("Header receive height:%d - %d", blkHeader.BlkHdr[0].Height, blkHeader.BlkHdr[len(blkHeader.BlkHdr)-1].Height)
						curHeaderBlockHash    =  blkHeader.BlkHdr[len(blkHeader.BlkHdr)-1].Hash()
						currHeaderHeight =  blkHeader.BlkHdr[len(blkHeader.BlkHdr)-1].Height

						headerIndexlock.Lock()
						for _, h := range blkHeader.BlkHdr {
							headerIndex[h.Height] = h.Hash()
						}
						headerIndexlock.Unlock()
						/*msgHeaderReq := msgpack.NewHeadersReq(curHeaderBlockHash)
						err := link.Tx(msgHeaderReq)
						if err != nil {
							log.Warn("[p2p]failed to send a new headersReq:s", err)
						}*/
					}else {
						log.Info("Header receive, the len is 0")
					}
				case msgCommon.BLOCK_TYPE:
					block := data.Payload.(*types.Block)
					log.Infof("Receive block msg, block height=%d", block.Blk.Header.Height)
				}
			}
		}
	}
}

func heartBeatService(link* connL.Link) {
	var periodTime uint
	periodTime = 3
	t := time.NewTicker(time.Second * (time.Duration(periodTime)))
	for {
		select {
		case <-t.C:
			ping := msgpack.NewPingMsg(0)
			go link.Tx(ping)
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

func syncHeader(link* connL.Link) {
	msgHeaderReq := msgpack.NewHeadersReq(curHeaderBlockHash)
	err := link.Tx( msgHeaderReq)
	if err != nil {
		log.Warn("[p2p]failed to send a new headersReq:s", err)
	}
}

func syncBlock(link* connL.Link) {
	curBlockHeight := uint32(0)
	curHeaderHeight := currHeaderHeight
	count := int(curHeaderHeight - curBlockHeight)
	if count <= 0 {
		return
	}

	counter := 1
	i := uint32(0)
	reqTimes := 1
	for {
		if counter > count {
			break
		}
		i++
		nextBlockHeight := curBlockHeight + i

		headerIndexlock.Lock()
		nextBlockHash := headerIndex[nextBlockHeight]
		headerIndexlock.Unlock()

		if nextBlockHash == common.UINT256_EMPTY {
			return
		}

		if nextBlockHeight <= curBlockHeight+SYNC_NEXT_BLOCKS_HEIGHT {
			reqTimes = SYNC_NEXT_BLOCK_TIMES
		}
		for t := 0; t < reqTimes; t++ {
			msg := msgpack.NewBlkDataReq(nextBlockHash)
			err := link.Tx(msg)
			if err != nil {
				log.Warnf("[p2p]syncBlock Height:%d ReqBlkData error:%s", nextBlockHeight, err)
				return
			}
		}
		counter++
		reqTimes = 1
	}
}

func sync(link* connL.Link){
	syncHeader(link)
	syncBlock(link)
}


func syncBlockTimer(link* connL.Link) {
	go sync(link)
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			go sync(link)
		}
	}
}

func main() {

	headerIndex = make(map[uint32]common.Uint256)

	server := ns.NewNetServer()

	conn, err := ns.NTLSDial("127.0.0.1:20338")
	if err != nil {
		log.Infof("[p2p]connect failed:%s", err.Error())
		return
	}
	addr := conn.RemoteAddr().String()

	recvmsgChan := server.GetMsgChan(false)

	link :=  connL.NewLink()
	link.SetAddr(addr)
	link.SetConn(conn)
	link.SetChan(recvmsgChan)
	go link.Rx()

	go hookChan(recvmsgChan, link)

	version := msgpack.NewVersion(server, false, 0)
	err = link.Tx(version)
	if err != nil {
		log.Warn(err)
		return
	}

	waitToExit()
}
