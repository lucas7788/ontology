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

package overlaydb

import (
	"crypto/sha256"
	"fmt"
	comm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/store/common"
	"github.com/syndtr/goleveldb/leveldb/util"
	"os"
)

type OverlayDB struct {
	store     common.PersistStore
	memdb     *MemDB
	ReadCache *MemDB
	dbErr     error
}

var IS_SHOW = false
var IS_SHOW_TEST = false

const initCap = 4 * 1024
const initkvNum = 128

func NewOverlayDB(store common.PersistStore) *OverlayDB {
	return &OverlayDB{
		store:     store,
		memdb:     NewMemDB(initCap, initkvNum),
		ReadCache: NewMemDB(initCap, initkvNum),
	}
}

func (self *OverlayDB) Reset() {
	self.memdb.Reset()
}

func (self *OverlayDB) Error() error {
	return self.dbErr
}

func (self *OverlayDB) SetError(err error) {
	self.dbErr = err
}

// if key is deleted, value == nil
func (self *OverlayDB) Get(key []byte) (value []byte, err error) {
	if comm.ToHexString(key) == "058367e1a4a25772ca49013d6f3a13ac46838516c34c55434b595f4845524f5f524f554e445f015f524f554e445f454e44494e475f424c4f434b5f484549474854" {
		fmt.Printf("")
		IS_SHOW_TEST = false
	}
	var unknown bool
	value, unknown = self.memdb.Get(key)
	if unknown == false {
		if IS_SHOW {
			fmt.Fprintf(os.Stderr, "*************key:%x, val: %x\n", key, value)
		}
		return value, nil
	}

	value, err = self.store.Get(key)
	if err != nil {
		if err == common.ErrNotFound {
			if IS_SHOW {
				fmt.Fprintf(os.Stderr, "*************key:%x, val: %x\n", key, value)
			}
			return nil, nil
		}
		self.dbErr = err
		if IS_SHOW {
			fmt.Fprintf(os.Stderr, "*************key:%x, val: %x\n", key, value)
		}
		return nil, err
	}
	if IS_SHOW {
		fmt.Fprintf(os.Stderr, "*************key:%x, val: %x\n", key, value)
	}
	self.ReadCache.Put(key, value)
	return
}
func (self *OverlayDB) GetReadCache() *MemDB {
	return self.ReadCache
}

func (self *OverlayDB) Put(key []byte, value []byte) {
	if comm.ToHexString(key) == "058367e1a4a25772ca49013d6f3a13ac46838516c34c55434b595f4845524f5f524f554e445f015f524f554e445f454e44494e475f424c4f434b5f484549474854" {
		fmt.Println("")
	}
	if IS_SHOW {
		fmt.Fprintf(os.Stderr, "PUT*************key:%x, val: %x\n", key, value)
	}
	self.memdb.Put(key, value)
}

func (self *OverlayDB) Delete(key []byte) {
	self.memdb.Delete(key)
}

func (self *OverlayDB) CommitTo() {
	self.memdb.ForEach(func(key, val []byte) {
		if len(val) == 0 {
			self.store.BatchDelete(key)
		} else {
			self.store.BatchPut(key, val)
		}
	})
}

func (self *OverlayDB) GetWriteSet() *MemDB {
	return self.memdb
}

func (self *OverlayDB) ChangeHash() comm.Uint256 {
	stateDiff := sha256.New()
	self.memdb.ForEach(func(key, val []byte) {
		stateDiff.Write(key)
		stateDiff.Write(val)
	})

	var hash comm.Uint256
	stateDiff.Sum(hash[:0])
	return hash
}

// param key is referenced by iterator
func (self *OverlayDB) NewIterator(key []byte) common.StoreIterator {
	prefixRange := util.BytesPrefix(key)
	backIter := self.store.NewIterator(key)
	memIter := self.memdb.NewIterator(prefixRange)

	return NewJoinIter(memIter, backIter, self.ReadCache)
}
