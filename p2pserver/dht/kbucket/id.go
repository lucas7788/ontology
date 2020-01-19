package kbucket

import (
	"strconv"

	"github.com/minio/sha256-simd"
	ks "github.com/ontio/ontology/p2pserver/dht/kbucket/keyspace"
)

// ID is used in dht
type ID []byte

// ConvertPeerID convert ontology peerid to dht id in bucket
func ConvertPeerID(id uint64) ID {
	hash := sha256.Sum256([]byte(strconv.FormatUint(id, 10)))
	return hash[:]
}

func xor(a, b ID) ID {
	return ID(ks.XOR(a, b))
}

// Closer returns true if a is closer to key than b is
func Closer(a, b, key uint64) bool {
	aid := ConvertPeerID(a)
	bid := ConvertPeerID(b)
	tgt := ConvertPeerID(key)
	adist := xor(aid, tgt)
	bdist := xor(bid, tgt)

	return adist.less(bdist)
}

func (id ID) less(other ID) bool {
	a := ks.Key{Space: ks.XORKeySpace, Bytes: id}
	b := ks.Key{Space: ks.XORKeySpace, Bytes: other}
	return a.Less(b)
}

// CommonPrefixLen calculate cpl
func CommonPrefixLen(a, b ID) int {
	return ks.ZeroPrefixLen(ks.XOR(a, b))
}
