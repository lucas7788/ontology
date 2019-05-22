package ledgerstore

import (
	"encoding/binary"
	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/overlaydb"
)

type MockDB struct {
	store common.PersistStore
	db    map[string]string
}

//Put(key []byte, value []byte) error      //Put the key-value pair to store
//Get(key []byte) ([]byte, error)          //Get the value if key in store
//Has(key []byte) (bool, error)            //Whether the key is exist in store
//Delete(key []byte) error                 //Delete the key in store
//NewBatch()                               //Start commit batch
//BatchPut(key []byte, value []byte)       //Put a key-value pair to batch
//BatchDelete(key []byte)                  //Delete the key in batch
//BatchCommit() error                      //Commit batch to store
//Close() error                            //Close store
//NewIterator(prefix []byte) StoreIterator

func (self *MockDB) Get(key []byte) ([]byte, error) {
	val, ok := self.db[string(key)]
	if ok == false {
		return nil, common.ErrNotFound
	}
	return []byte(val), nil
}

func (self *MockDB) BatchPut(key []byte, value []byte) {
	self.db[string(key)] = string(value)
}

func (self *MockDB) BatchDelete(key []byte) {
	delete(self.db, string(key))
}
func (self *MockDB) Put(key []byte, value []byte) error {
	return self.store.Put(key, value)
}
func (self *MockDB) Has(key []byte) (bool, error) {
	return self.store.Has(key)
}

func (self *MockDB) BatchCommit() error {
	return self.store.BatchCommit()
}
func (self *MockDB) Close() error {
	return self.store.Close()
}
func (self *MockDB) Delete(key []byte) error {
	return self.store.Delete(key)
}

func (self *MockDB) NewBatch() {
	self.store.NewBatch()
}

func (self *MockDB) NewIterator(prefix []byte) common.StoreIterator {
	return self.store.NewIterator(prefix)
}

func (self *MockDB) NewOverlayDB(height uint32) *overlaydb.OverlayDB {
	//get before execute data
	key := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(key[:], uint32(height))
	dataBytes, err := self.store.Get(key)
	if err != nil {
		return nil
	}
	source := common2.NewZeroCopySource(dataBytes)
	data := make(map[string]string)
	for {
		key, _, _, eof := source.NextVarBytes()
		if eof {
			break
		}
		value, _, _, eof := source.NextVarBytes()
		if eof {
			break
		}
		data[string(key)] = string(value)
	}
	self.db = data
	return overlaydb.NewOverlayDB(self)
}
