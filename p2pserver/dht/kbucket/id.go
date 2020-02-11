/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

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
