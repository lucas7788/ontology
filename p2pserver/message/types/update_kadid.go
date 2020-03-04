package types

import (
	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/dht/kbucket"
)

type UpdateKadId struct {
	//TODO remove this legecy field when upgrade network layer protocal
	KadKeyId *kbucket.KadKeyId
}

//Serialize message payload
func (this *UpdateKadId) Serialization(sink *common2.ZeroCopySink) {
	this.KadKeyId.Serialization(sink)
}

func (this *UpdateKadId) Deserialization(source *common2.ZeroCopySource) error {
	return this.KadKeyId.Deserialization(source)
}

func (this *UpdateKadId) CmdType() string {
	return common.UPDATE_KADID_TYPE
}
