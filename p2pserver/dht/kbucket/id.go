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
	"crypto/sha256"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	ks "github.com/ontio/ontology/p2pserver/dht/kbucket/keyspace"
	"io"
)

var Difficulty = 18 //bit

type KadId = common.Address

type KadKeyId struct {
	PublicKey keypair.PublicKey
	Id        KadId
}

func (this *KadKeyId) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(keypair.SerializePublicKey(this.PublicKey))
	sink.WriteAddress(this.Id)
}

func (this *KadKeyId) Deserialization(source *common.ZeroCopySource) error {
	data, _, irregular, eof := source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	pub, err := keypair.DeserializePublicKey(data)
	if err != nil {
		return err
	}
	this.PublicKey = pub
	this.Id, eof = source.NextAddress()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func KadKeyIdFromPubkey(pubKey keypair.PublicKey) *KadKeyId {
	addr := types.AddressFromPubKey(pubKey)
	return &KadKeyId{
		PublicKey: pubKey,
		Id:        addr,
	}
}

func KadIdFromBytes(addr []byte) (KadId, error) {
	return common.AddressParseFromBytes(addr)
}

func GenerateRandomId() *KadKeyId {
	var acc *account.Account
	for {
		acc = account.NewAccount("")
		if ValidatePublicKey(acc.PublicKey) {
			break
		}
	}
	addr := types.AddressFromPubKey(acc.PublicKey)
	return &KadKeyId{
		PublicKey: acc.PublicKey,
		Id:        addr,
	}
}

func ValidatePublicKey(pubKey keypair.PublicKey) bool {
	pub := keypair.SerializePublicKey(pubKey)
	res := sha256.Sum256(pub)
	hash := sha256.Sum256(res[:])
	limit := Difficulty >> 3
	for i := 0; i < limit; i++ {
		if hash[i] != 0 {
			return false
		}
	}
	diff := Difficulty - limit*8
	if diff != 0 {
		x := hash[limit] >> uint8(8-diff)
		if x != 0 {
			return false
		}
	}
	return true
}

func xor(a, b KadId) KadId {
	id, _ := KadIdFromBytes(ks.XOR(a[:], b[:]))
	return id
}

// Closer returns true if a is closer to key than b is
func Closer(a, b, key KadId) bool {
	adist := xor(a, key)
	bdist := xor(b, key)

	return less(adist, bdist)
}

func less(id, other KadId) bool {
	a := ks.Key{Space: ks.XORKeySpace, Bytes: id[:]}
	b := ks.Key{Space: ks.XORKeySpace, Bytes: other[:]}
	return a.Less(b)
}

// CommonPrefixLen(cpl) calculate two ID's xor prefix 0
func CommonPrefixLen(a, b KadId) int {
	return ks.ZeroPrefixLen(ks.XOR(a[:], b[:]))
}
