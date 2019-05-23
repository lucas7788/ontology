package ledgerstore

import (
	"encoding/binary"
	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type MockDBStore struct {
	store common.PersistStore
}

func NewMockDBStore(store *leveldbstore.LevelDBStore) *MockDBStore {
	return &MockDBStore{
		store: store,
	}
}

func (self *MockDBStore) Put(key []byte, value []byte) error {
	return self.store.Put(key, value)
}

func (self *MockDBStore) Get(key []byte) ([]byte, error) {
	return self.store.Get(key)
}

func (self *MockDBStore) Has(key []byte) (bool, error) {
	return self.store.Has(key)
}

func (self *MockDBStore) Delete(key []byte) error {
	return self.store.Delete(key)
}

func (self *MockDBStore) NewBatch() {
	self.store.NewBatch()
}

func (self *MockDBStore) BatchPut(key []byte, value []byte) {
	self.store.BatchPut(key, value)
}

func (self *MockDBStore) BatchDelete(key []byte) {}

func (self *MockDBStore) BatchCommit() error {
	return self.store.BatchCommit()
}

func (self *MockDBStore) Close() {
	self.store.Close()
}

func (self *MockDBStore) NewIterator(prefix []byte) common.StoreIterator {
	return nil
}

type MockDB struct {
	db *overlaydb.MemDB
}

func NewMockDB() *MockDB {
	return &MockDB{db: overlaydb.NewMemDB(16*1024, 16)}
}

func (self *MockDB) NewIterator(prefix []byte) common.StoreIterator {
	prefixRange := util.BytesPrefix(prefix)
	return self.db.NewIterator(prefixRange)
}

func (self *MockDB) Put(key []byte, value []byte) error {
	self.db.Put(key, value)
	return nil
}

func (self *MockDB) Get(key []byte) ([]byte, error) {
	value, unknow := self.db.Get(key)
	if unknow == false {
		return value, nil
	}
	return nil, nil
}

func (self *MockDB) Has(key []byte) (bool, error) {
	_, unknow := self.db.Get(key)
	if unknow {
		return false, nil
	}
	return true, nil
}

func (self *MockDB) Delete(key []byte) error {
	self.db.Delete(key)
	return nil
}

func (self *MockDB) NewBatch() {}

func (self *MockDB) BatchPut(key []byte, value []byte) {}

func (self *MockDB) BatchDelete(key []byte) {
}

func (self *MockDB) BatchCommit() error {
	return nil
}

func (self *MockDB) Close() error {
	return nil
}

func NewOverlayDB(height uint32, store *MockDBStore) *overlaydb.OverlayDB {
	//get before execute data
	key := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(key[:], uint32(height))
	dataBytes, err := store.Get(key)
	if err != nil {
		return nil
	}
	source := common2.NewZeroCopySource(dataBytes)
	mockDB := NewMockDB()
	l, eof := source.NextUint32()
	if eof {
		return nil
	}
	for i := uint32(0); i < l; i++ {
		key, _, irregular, eof := source.NextVarBytes()
		if eof || irregular {
			break
		}
		value, _, _, eof := source.NextVarBytes()
		if eof {
			break
		}
		mockDB.db.Put(key, value)
	}
	return overlaydb.NewOverlayDB(mockDB)
}
