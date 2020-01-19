package kbucket

import (
	"container/list"
	"sync"
)

// Bucket holds a list of peers.
type Bucket struct {
	lk   sync.RWMutex
	list *list.List
}

func newBucket() *Bucket {
	b := new(Bucket)
	b.list = list.New()
	return b
}

func (b *Bucket) Peers() []uint64 {
	b.lk.RLock()
	defer b.lk.RUnlock()
	ps := make([]uint64, 0, b.list.Len())
	for e := b.list.Front(); e != nil; e = e.Next() {
		id := e.Value.(uint64)
		ps = append(ps, id)
	}
	return ps
}

func (b *Bucket) Has(id uint64) bool {
	b.lk.RLock()
	defer b.lk.RUnlock()
	for e := b.list.Front(); e != nil; e = e.Next() {
		if e.Value.(uint64) == id {
			return true
		}
	}
	return false
}

func (b *Bucket) Remove(id uint64) bool {
	b.lk.Lock()
	defer b.lk.Unlock()
	for e := b.list.Front(); e != nil; e = e.Next() {
		if e.Value.(uint64) == id {
			b.list.Remove(e)
			return true
		}
	}
	return false
}

func (b *Bucket) MoveToFront(id uint64) {
	b.lk.Lock()
	defer b.lk.Unlock()
	for e := b.list.Front(); e != nil; e = e.Next() {
		if e.Value.(uint64) == id {
			b.list.MoveToFront(e)
		}
	}
}

func (b *Bucket) PushFront(p uint64) {
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
func (b *Bucket) Split(cpl int, target ID) *Bucket {
	b.lk.Lock()
	defer b.lk.Unlock()

	out := list.New()
	newbuck := newBucket()
	newbuck.list = out
	e := b.list.Front()
	for e != nil {
		peerID := ConvertPeerID(e.Value.(uint64))
		peerCPL := CommonPrefixLen(peerID, target)
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
