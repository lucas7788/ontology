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

package TestCommon

import (
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/chainmgr/xshard"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

func GetShardStateFromLedger(t *testing.T, lgr *ledger.Ledger, shardID common.ShardID) *shardstates.ShardState {
	state, err := xshard.GetShardState(lgr, shardID)
	if err != nil {
		t.Fatalf(err.Error())
	}
	return state
}

func GetStorageItem(t *testing.T, lgr *ledger.Ledger, contract common.Address, key []byte) []byte {
	item , err := lgr.GetStorageItem(contract, key)
	if err != nil {
		t.Fatalf("GetStorageItem err : %s", err)
	}
	return item
}

var VBFT_CONFIG = config.VBFTConfig{
	N:                    uint32(7),
	C:                    uint32(2),
	K:                    uint32(7),
	L:                    uint32(112),
	BlockMsgDelay:        uint32(5000),
	HashMsgDelay:         uint32(10000),
	PeerHandshakeTimeout: uint32(10),
	MaxBlockChangeView:   uint32(5),
	MinInitStake:         uint32(10000),
	AdminOntID:           "did:ont:ARiwjLzjzLKZy8V43vm6yUcRG9b56DnZtY",
	VrfValue:             "1c9810aa9822e511d5804a9c4db9dd08497c31087b0daafa34d768a3253441fa20515e2f30f81741102af0ca3cefc4818fef16adb825fbaa8cad78647f3afb590e",
	VrfProof:             "c57741f934042cb8d8b087b44b161db56fc3ffd4ffb675d36cd09f83935be853d8729f3f5298d12d6fd28d45dde515a4b9d7f67682d182ba5118abf451ff1988",
	Peers:                []*config.VBFTPeerStakeInfo{},
}
