package dht

import (
	"context"
	"time"

	"github.com/ontio/ontology/common/log"
	kb "github.com/ontio/ontology/p2pserver/dht/kbucket"
)

// Pool size is the number of nodes used for group find/set RPC calls
var PoolSize = 6

// K is the maximum number of requests to perform before returning failure.
var KValue = 20

// Alpha is the concurrency factor for asynchronous requests.
var AlphaValue = 3

type DHT struct {
	self  uint64
	birth time.Time // When this peer started up

	ctx context.Context

	bucketSize   int
	routingTable *kb.RoutingTable // Array of routing tables for differently distanced nodes

	AutoRefresh           bool
	RtRefreshQueryTimeout time.Duration
	RtRefreshPeriod       time.Duration
}

// Context return dht's context
func (dht *DHT) Context() context.Context {
	return dht.ctx
}

// RoutingTable return dht's routingTable
func (dht *DHT) RoutingTable() *kb.RoutingTable {
	return dht.routingTable
}

// New creates a new DHT with the specified host and options.
func New(ctx context.Context, localID uint64) (*DHT, error) {
	dht := makeDHT(ctx, localID, KValue)
	dht.RtRefreshPeriod = 10 * time.Second
	dht.RtRefreshQueryTimeout = 10 * time.Second

	return dht, nil
}

func makeDHT(ctx context.Context, localID uint64, bucketSize int) *DHT {
	self := kb.ConvertPeerID(localID)
	rt := kb.NewRoutingTable(bucketSize, self)

	rt.PeerAdded = func(p uint64) {
		log.Debugf("dht: peer: %d added to dht", p)
	}

	rt.PeerRemoved = func(p uint64) {
		log.Debugf("dht: peer: %d removed from dht", p)
	}

	dht := &DHT{
		self:         localID,
		ctx:          ctx,
		birth:        time.Now(),
		routingTable: rt,
		bucketSize:   bucketSize,

		AutoRefresh: true,
	}

	return dht
}

// Update signals the routingTable to Update its last-seen status
// on the given peer.
func (dht *DHT) Update(peer uint64) bool {
	err := dht.routingTable.Update(peer)

	return err == nil
}

func (dht *DHT) Remove(peer uint64) {
	dht.routingTable.Remove(peer)
}

func (dht *DHT) BetterPeers(id uint64, count int) []uint64 {
	closer := dht.routingTable.NearestPeers(kb.ConvertPeerID(id), count)
	filtered := make([]uint64, 0, len(closer))
	// don't include self and target id
	for _, curID := range closer {
		if curID == dht.self {
			continue
		}
		if curID == id {
			continue
		}
		filtered = append(filtered, curID)
	}

	return filtered
}
