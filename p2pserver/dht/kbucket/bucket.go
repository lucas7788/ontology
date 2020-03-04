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
	"bytes"
	"container/list"
	"sync"
)

// Bucket holds a list of peers.
type Bucket struct {
	lk   sync.RWMutex
	list *list.List
}

type KPId struct {
	KId KadId
	PId uint64
}

func newBucket() *Bucket {
	b := new(Bucket)
	b.list = list.New()
	return b
}

func (b *Bucket) Peers() []*KPId {
	b.lk.RLock()
	defer b.lk.RUnlock()
	ps := make([]*KPId, 0, b.list.Len())
	for e := b.list.Front(); e != nil; e = e.Next() {
		id := e.Value.(*KPId)
		ps = append(ps, id)
	}
	return ps
}

func (b *Bucket) Has(id KadId) bool {
	b.lk.RLock()
	defer b.lk.RUnlock()
	for e := b.list.Front(); e != nil; e = e.Next() {
		curr := e.Value.(*KPId)
		if bytes.Compare(curr.KId[:], id[:]) == 0 {
			return true
		}
	}
	return false
}

func (b *Bucket) Remove(id KadId) bool {
	b.lk.Lock()
	defer b.lk.Unlock()
	for e := b.list.Front(); e != nil; e = e.Next() {
		curr := e.Value.(*KPId)
		if bytes.Compare(curr.KId[:], id[:]) == 0 {
			return true
		}
	}
	return false
}

func (b *Bucket) MoveToFront(id *KPId) {
	b.lk.Lock()
	defer b.lk.Unlock()
	for e := b.list.Front(); e != nil; e = e.Next() {
		curr := e.Value.(*KPId)
		if bytes.Compare(curr.KId[:], id.KId[:]) == 0 {
			b.list.MoveToFront(e)
		}
	}
}

func (b *Bucket) PushFront(p *KPId) {
	b.lk.Lock()
	b.list.PushFront(p)
	b.lk.Unlock()
}

func (b *Bucket) PopBack() uint64 {
	b.lk.Lock()
	defer b.lk.Unlock()
	last := b.list.Back()
	b.list.Remove(last)
	return last.Value.(uint64)
}

func (b *Bucket) Len() int {
	b.lk.RLock()
	defer b.lk.RUnlock()
	return b.list.Len()
}

// Split splits a buckets peers into two buckets, the methods receiver will have
// peers with CPL equal to cpl, the returned bucket will have peers with CPL
// greater than cpl (returned bucket has closer peers)
// CPL ==> CommonPrefixLen
func (b *Bucket) Split(cpl int, target KadId) *Bucket {
	b.lk.Lock()
	defer b.lk.Unlock()

	out := list.New()
	newbuck := newBucket()
	newbuck.list = out
	e := b.list.Front()
	for e != nil {
		peerID := e.Value.(*KPId)
		peerCPL := CommonPrefixLen(peerID.KId, target)
		if peerCPL > cpl {
			cur := e
			out.PushBack(e.Value)
			e = e.Next()
			b.list.Remove(cur)
			continue
		}
		e = e.Next()
	}
	return newbuck
}
