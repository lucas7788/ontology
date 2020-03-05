package netserver

import (
	"fmt"
	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/dht/kbucket"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	"github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/peer"
	"net"
	"strconv"
	"strings"
	"time"
)

func HandshakeClient(netServer *NetServer, conn net.Conn) error {
	addr := conn.RemoteAddr().String()
	log.Debugf("[p2p]peer %s connect with %s with %s",
		conn.LocalAddr().String(), conn.RemoteAddr().String(),
		conn.RemoteAddr().Network())

	// 1. send version
	version := msgpack.NewVersion(netServer, ledger.DefLedger.GetCurrentBlockHeight())
	sink := common2.NewZeroCopySink(nil)
	types.WriteMessage(sink, version)
	err := send(conn, sink.Bytes())
	if err != nil {
		netServer.RemoveFromOutConnRecord(addr)
		log.Warn(err)
		return err
	}

	// 2. read version
	msg, _, err := types.ReadMessage(conn)
	if err != nil {
		return err
	}
	if msg.CmdType() != common.VERSION_TYPE {
		return fmt.Errorf("")
	}
	// 3. update kadId
	versionRaw := msg.(*types.Version)
	remotePeer, err := versionHandle(netServer, versionRaw, conn)
	if err != nil {
		return err
	}
	if remotePeer == nil {
		return fmt.Errorf("remote peer is nil")
	}

	if versionRaw.P.SoftVersion > "v1.9.0" && false {
		log.Info("*******come in dht*******")
		msg := msgpack.NewUpdateKadKeyId(netServer)
		sink.Reset()
		types.WriteMessage(sink, msg)
		err = send(conn, sink.Bytes())
		if err != nil {
			return err
		}
		// 4. read kadkeyid
		msg, _, err = types.ReadMessage(conn)
		if err != nil {
			return err
		}
		if msg.CmdType() != common.UPDATE_KADID_TYPE {
			return fmt.Errorf("")
		}
		kadKeyId := msg.(*types.UpdateKadId)

		if !kbucket.ValidatePublicKey(kadKeyId.KadKeyId.PublicKey) {
			return fmt.Errorf("validate publickey failed")
		}
		if !netServer.UpdateDHT(kadKeyId.KadKeyId.Id) {
			log.Errorf("[HandshakeClient] UpdateDHT failed, kadId: %s", kadKeyId.KadKeyId.Id.ToHexString())
			return fmt.Errorf("[HandshakeClient] UpdateDHT failed, kadId: %s", kadKeyId.KadKeyId.Id.ToHexString())
		}
		remotePeer.SetKId(kadKeyId.KadKeyId.Id)
	}

	// 5. send ack
	ack := msgpack.NewVerAck()
	sink.Reset()
	types.WriteMessage(sink, ack)
	err = send(conn, sink.Bytes())
	if err != nil {
		netServer.RemoveFromOutConnRecord(addr)
		log.Warn(err)
		return err
	}

	netServer.AddOutConnRecord(addr)
	netServer.AddPeerAddress(addr, remotePeer)
	remotePeer.Link.SetAddr(addr)
	remotePeer.Link.SetConn(conn)
	remotePeer.Link.SetID(remotePeer.GetID())
	remotePeer.AttachChan(netServer.NetChan)
	netServer.AddNbrNode(remotePeer)
	log.Infof("remotePeer.GetId():%d,addr: %s, link id: %d", remotePeer.GetID(), addr, remotePeer.Link.GetID())
	go remotePeer.Link.Rx()
	remotePeer.SetState(common.ESTABLISH)
	return nil
}

func HandshakeServer(netServer *NetServer, conn net.Conn) error {
	// 1. read version
	msg, _, err := types.ReadMessage(conn)
	if err != nil {
		log.Errorf("[HandshakeServer] ReadMessage failed, error: %s", err)
		return fmt.Errorf("[HandshakeServer] ReadMessage failed, error: %s", err)
	}
	if msg.CmdType() != common.VERSION_TYPE {
		return fmt.Errorf("[HandshakeServer] expected version message")
	}
	version := msg.(*types.Version)
	remotePeer, err := versionHandle(netServer, version, conn)
	if err != nil || remotePeer == nil {
		return err
	}
	// 2. send version
	sink := common2.NewZeroCopySink(nil)
	ver := msgpack.NewVersion(netServer, ledger.DefLedger.GetCurrentBlockHeight())
	types.WriteMessage(sink, ver)
	err = send(conn, sink.Bytes())
	if err != nil {
		log.Errorf("[HandshakeServer] WriteMessage send failed, error: %s", err)
	}

	// 3. read update kadkey id
	if version.P.SoftVersion > "v1.9.0" && false {
		msg, _, err := types.ReadMessage(conn)
		if err != nil {
			log.Errorf("[HandshakeServer] ReadMessage failed, error: %s", err)
			return fmt.Errorf("[HandshakeServer] ReadMessage failed, error: %s", err)
		}
		if msg.CmdType() != common.UPDATE_KADID_TYPE {
			return fmt.Errorf("[HandshakeServer] expected update kadkeyid message")
		}
		kadkeyId := msg.(*types.UpdateKadId)
		remotePeer := netServer.GetPeerFromAddr(conn.RemoteAddr().String())
		remotePeer.SetKId(kadkeyId.KadKeyId.Id)
		netServer.dht.Update(kadkeyId.KadKeyId.Id)
		if !kbucket.ValidatePublicKey(kadkeyId.KadKeyId.PublicKey) {
			log.Errorf("[HandshakeServer] ValidatePublicKey failed, kadId:%s", kadkeyId.KadKeyId.Id.ToHexString())
			return fmt.Errorf("[HandshakeServer] ValidatePublicKey failed, kadId:%s", kadkeyId.KadKeyId.Id.ToHexString())
		}
		netServer.dht.Update(kadkeyId.KadKeyId.Id)
		// 4. send update kadkey id
		msg = msgpack.NewUpdateKadKeyId(netServer)
		sink.Reset()
		types.WriteMessage(sink, msg)
		err = send(conn, sink.Bytes())
		if err != nil {
			return err
		}
		remotePeer.SetKId(kadkeyId.KadKeyId.Id)
	}

	// 5. read version ack
	msg, _, err = types.ReadMessage(conn)
	if err != nil {
		log.Errorf("[HandshakeServer] ReadMessage failed, error: %s", err)
		return fmt.Errorf("[HandshakeServer] ReadMessage failed, error: %s", err)
	}
	if msg.CmdType() != common.VERACK_TYPE {
		return fmt.Errorf("[HandshakeServer] expected version ack message")
	}

	netServer.AddNbrNode(remotePeer)

	addr := conn.RemoteAddr().String()
	netServer.AddInConnRecord(addr)

	netServer.AddPeerAddress(addr, remotePeer)

	remotePeer.Link.SetAddr(addr)
	remotePeer.Link.SetConn(conn)
	remotePeer.AttachChan(netServer.NetChan)
	go remotePeer.Link.Rx()
	return nil
}

func send(conn net.Conn, rawPacket []byte) error {
	nByteCnt := len(rawPacket)
	log.Tracef("[p2p]TX buf length: %d\n", nByteCnt)

	nCount := nByteCnt / common.PER_SEND_LEN
	if nCount == 0 {
		nCount = 1
	}
	_ = conn.SetWriteDeadline(time.Now().Add(time.Duration(nCount*common.WRITE_DEADLINE) * time.Second))
	_, err := conn.Write(rawPacket)
	if err != nil {
		log.Infof("[handshake]error sending messge to %s :%s", conn.LocalAddr(), err.Error())
		return err
	}
	return nil
}

func versionHandle(p2p *NetServer, version *types.Version, conn net.Conn) (*peer.Peer, error) {
	log.Infof("remoteAddr: %s, localAddr: %s", conn.RemoteAddr().String(), conn.LocalAddr().String())
	remoteAddr := conn.RemoteAddr().String()
	addrIp, err := common.ParseIPAddr(remoteAddr)
	if err != nil {
		log.Warn(err)
		return nil, err
	}
	nodeAddr := addrIp + ":" +
		strconv.Itoa(int(version.P.SyncPort))
	if config.DefConfig.P2PNode.ReservedPeersOnly && len(config.DefConfig.P2PNode.ReservedCfg.ReservedPeers) > 0 {
		found := false
		for _, addr := range config.DefConfig.P2PNode.ReservedCfg.ReservedPeers {
			if strings.HasPrefix(remoteAddr, addr) {
				log.Debug("[versionHandle]peer in reserved list", remoteAddr)
				found = true
				break
			}
		}
		if !found {
			log.Debug("[versionHandle]peer not in reserved list,close", remoteAddr)
			return nil, fmt.Errorf("the remote addr: %s not in ReservedPeers", remoteAddr)
		}
	}
	if version.P.Nonce == p2p.GetID() {
		p2p.RemoveFromInConnRecord(remoteAddr)
		p2p.RemoveFromOutConnRecord(remoteAddr)
		log.Warn("[versionHandle]the node handshake with itself:", remoteAddr)
		p2p.SetOwnAddress(nodeAddr)
		return nil, fmt.Errorf("[versionHandle]the node handshake with itself: %s", remoteAddr)
	}

	// Obsolete node
	p := p2p.GetPeer(version.P.Nonce)
	if p != nil {
		ipOld, err := common.ParseIPAddr(p.GetAddr())
		if err != nil {
			log.Warn("[versionHandle]exist peer %d ip format is wrong %s", version.P.Nonce, p.GetAddr())
			return nil, fmt.Errorf("[versionHandle]exist peer %d ip format is wrong %s", version.P.Nonce, p.GetAddr())
		}
		ipNew, err := common.ParseIPAddr(remoteAddr)
		if err != nil {
			log.Warn("[versionHandle]connecting peer %d ip format is wrong %s, close", version.P.Nonce, remoteAddr)
			return nil, fmt.Errorf("[versionHandle]connecting peer %d ip format is wrong %s, close", version.P.Nonce, remoteAddr)
		}
		if ipNew == ipOld {
			//same id and same ip
			n, delOK := p2p.DelNbrNode(version.P.Nonce)
			if delOK {
				log.Infof("[versionHandle]peer reconnect %d", version.P.Nonce, remoteAddr)
				// Close the connection and release the node source
				n.Close()
				if p2p.pid != nil {
					input := &common.RemovePeerID{
						ID: version.P.Nonce,
					}
					p2p.pid.Tell(input)
				}
			}
		} else {
			log.Warnf("[versionHandle]same peer id from different addr: %s, %s close latest one", ipOld, ipNew)
			return nil, nil
		}
	}

	remotePeer := peer.NewPeer()
	if version.P.Cap[common.HTTP_INFO_FLAG] == 0x01 {
		remotePeer.SetHttpInfoState(true)
	} else {
		remotePeer.SetHttpInfoState(false)
	}
	remotePeer.SetHttpInfoPort(version.P.HttpInfoPort)

	remotePeer.UpdateInfo(time.Now(), version.P.Version,
		version.P.Services, version.P.SyncPort, version.P.Nonce,
		version.P.Relay, version.P.StartHeight, version.P.SoftVersion)

	if p2p.pid != nil {
		input := &common.AppendPeerID{
			ID: version.P.Nonce,
		}
		p2p.pid.Tell(input)
	}
	return remotePeer, nil
}
