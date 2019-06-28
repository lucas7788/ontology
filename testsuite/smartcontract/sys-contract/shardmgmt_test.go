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

package TestContracts

import (
	"fmt"
	"testing"

	"bytes"
	"github.com/ontio/ont-ipfs/thirdparty/assert"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/chainmgr"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/testsuite"
	"github.com/ontio/ontology/testsuite/common"
	"math/big"
)

func init() {
	TestConsts.TestRootDir = "../../"
}

func Test_ShardMgmtInit(t *testing.T) {

	// 1. create root chain
	shardID := common.NewShardIDUnchecked(config.DEFAULT_SHARD_ID)
	TestCommon.CreateChain(t, "test", shardID, 0)

	// 2. build shard-mgmt init tx

	tx := TestCommon.CreateAdminTx(t, shardID, utils.ShardMgmtContractAddress, shardmgmt.INIT_NAME, nil)

	// 3. create new block
	blk := TestCommon.CreateBlock(t, ledger.GetShardLedger(shardID), []*types.Transaction{tx})

	// 4. add block
	TestCommon.ExecBlock(t, shardID, blk)
	TestCommon.SubmitBlock(t, shardID, blk)

	// 5. query db
	state := TestCommon.GetShardStateFromLedger(t, ledger.GetShardLedger(shardID), shardID)
	fmt.Printf("%v", state)
}

func Test_ShardCreate(t *testing.T) {
	// 1. create root chain
	shardID := common.NewShardIDUnchecked(config.DEFAULT_SHARD_ID)
	TestCommon.CreateChain(t, "test", shardID, 0)

	shardName := chainmgr.GetShardName(shardID)
	user := shardName + "_peerOwner" + fmt.Sprintf("%d", 0)
	users := make([]string, 0)
	for i:=0;i<7;i++ {
		users = append(users, shardName + "_peerOwner" + fmt.Sprintf("%d", i))
	}
	acc := TestCommon.GetAccount(user)
	accounts := make([]*account.Account, 0)
	accounts = append(accounts, acc)
	pks := make([]keypair.PublicKey, 0)
	pks = append(pks, acc.PublicKey)
	for i := 1; i < 7; i++ {
		acc := TestCommon.GetAccount(shardName + "_peerOwner" + fmt.Sprintf("%d", i))
		pks = append(pks, acc.PublicKey)
		accounts = append(accounts, acc)
	}

	withdrawOng(t, shardID, pks, accounts)

	// 2. build shard-mgmt init tx

	tx := TestCommon.CreateAdminTx(t, shardID, utils.ShardMgmtContractAddress, shardmgmt.INIT_NAME, nil)

	//build SetCreateShardFee tx
	txSetCreateShardFee := getSetCreateFeeTx(t, shardID)

	// build CreateShard tx
	txCreate := getCreateShardtx(t, shardID, acc, user)

	// 3. create new block
	blk := TestCommon.CreateBlock(t, ledger.GetShardLedger(shardID), []*types.Transaction{tx, txSetCreateShardFee, txCreate})

	// 4. add block
	TestCommon.ExecBlock(t, shardID, blk)
	TestCommon.SubmitBlock(t, shardID, blk)

	// 5. query db
	state := TestCommon.GetShardStateFromLedger(t, ledger.GetShardLedger(shardID), shardID)
	fmt.Printf("%v\n", state)

	ledger0 := ledger.GetShardLedger(shardID)
	notify, err := ledger0.GetEventNotifyByTx(txCreate.Hash())
	assert.Nil(err, t)
	createShardNotifyState := notify.Notify[0].States
	fmt.Printf("createShard event: %v\n", createShardNotifyState)
	//get new shard state
	stat := createShardNotifyState.(map[string]interface{})
	newShardId := common.NewShardIDUnchecked(uint64(stat["ToShard"].(float64)))
	state = TestCommon.GetShardStateFromLedger(t, ledger.GetShardLedger(shardID), newShardId)
	fmt.Printf("new shard state: %v\n", state)

	//build ConfigShard tx
	txConfigShard := getConfigShardTx(t, newShardId, user)

	//build ApplyJoinShard tx

	txApplyJoinShard1 := getApplyJoinShardTx(t, newShardId, users[0], accounts[0])
	txApplyJoinShard2 := getApplyJoinShardTx(t, newShardId, users[1], accounts[1])
	txApplyJoinShard3 := getApplyJoinShardTx(t, newShardId, users[2], accounts[2])
	txApplyJoinShard4 := getApplyJoinShardTx(t, newShardId, users[3], accounts[3])
	txApplyJoinShard5 := getApplyJoinShardTx(t, newShardId, users[4], accounts[4])
	txApplyJoinShard6 := getApplyJoinShardTx(t, newShardId, users[5], accounts[5])
	txApplyJoinShard7 := getApplyJoinShardTx(t, newShardId, users[6], accounts[6])

	// build ApproveJoinShard tx

	txApproveJoinShard := getApproveJoinShardTx(t, newShardId, user, accounts)

	//build JoinShardTx
	txJoinShard1 := getJoinShardTx(t, newShardId, users[0], accounts[0])
	txJoinShard2 := getJoinShardTx(t, newShardId, users[1], accounts[1])
	txJoinShard3 := getJoinShardTx(t, newShardId, users[2], accounts[2])
	txJoinShard4 := getJoinShardTx(t, newShardId, users[3], accounts[3])
	txJoinShard5 := getJoinShardTx(t, newShardId, users[4], accounts[4])
	txJoinShard6 := getJoinShardTx(t, newShardId, users[5], accounts[5])
	txJoinShard7 := getJoinShardTx(t, newShardId, users[6], accounts[6])

	//build activiteTx
	activiteTx := getActivateShardTx(t, newShardId, user)

	//  create new block
	blk = TestCommon.CreateBlock(t, ledger.GetShardLedger(shardID), []*types.Transaction{txConfigShard, txApplyJoinShard1,txApplyJoinShard2,
	txApplyJoinShard3,txApplyJoinShard4,txApplyJoinShard5,txApplyJoinShard6,txApplyJoinShard7, txApproveJoinShard, txJoinShard1,txJoinShard2,
		txJoinShard3,txJoinShard4,txJoinShard5,txJoinShard6,txJoinShard7,activiteTx})

	//  add block
	TestCommon.ExecBlock(t, shardID, blk)
	TestCommon.SubmitBlock(t, shardID, blk)
	checkNotify(t, "txConfigShard", txConfigShard.Hash(), ledger.GetShardLedger(shardID))

	state = TestCommon.GetShardStateFromLedger(t, ledger.GetShardLedger(shardID), newShardId)
	fmt.Printf("new shard state: %v\n", state)

	checkNotify(t, "txJoinShard1", txJoinShard1.Hash(), ledger.GetShardLedger(shardID))
	checkNotify(t, "txJoinShard2", txJoinShard2.Hash(), ledger.GetShardLedger(shardID))
	checkNotify(t, "txJoinShard3", txJoinShard3.Hash(), ledger.GetShardLedger(shardID))
	checkNotify(t, "txJoinShard4", txJoinShard4.Hash(), ledger.GetShardLedger(shardID))
	checkNotify(t, "txJoinShard5", txJoinShard5.Hash(), ledger.GetShardLedger(shardID))
	checkNotify(t, "txJoinShard6", txJoinShard6.Hash(), ledger.GetShardLedger(shardID))
	checkNotify(t, "txJoinShard7", txJoinShard7.Hash(), ledger.GetShardLedger(shardID))
	checkNotify(t, "activiteTx", activiteTx.Hash(), ledger.GetShardLedger(shardID))

}

func withdrawOng(t *testing.T, shard common.ShardID, pks []keypair.PublicKey, accounts []*account.Account) {
	multiAddress, err := types.AddressFromMultiPubKeys(pks, 5)
	if err != nil {
		t.Fatalf("AddressFromMultiPubKeys err: %s", err)
	}

	param := make([]ont.State, 0)
	for i := 0; i < 7; i++ {
		s := ont.State{
			From:  multiAddress,
			To:    accounts[i].Address,
			Value: uint64(200000),
		}
		param = append(param, s)
	}
	tr := ont.Transfers{
		States: param,
	}
	transferOngTx := TestCommon.CreateAdminTx2(t, shard, utils.OntContractAddress, ont.TRANSFER_NAME, []interface{}{tr})

	transferFromTx := make([]*types.Transaction, 0)
	transferFromTx = append(transferFromTx, transferOngTx)
	for i := 0; i < 7; i++ {
		trans := ont.TransferFrom{
			Sender: multiAddress,
			From:   utils.OntContractAddress,
			To:     accounts[i].Address,
			Value:  uint64(1000000000000),
		}
		withdrawOngTx := TestCommon.CreateAdminTx2(t, shard, utils.OngContractAddress, ont.TRANSFERFROM_NAME, []interface{}{trans})
		transferFromTx = append(transferFromTx, withdrawOngTx)
	}

	ldg := ledger.GetShardLedger(shard)
	blk := TestCommon.CreateBlock(t, ldg, transferFromTx)

	//  add block
	TestCommon.ExecBlock(t, shard, blk)
	TestCommon.SubmitBlock(t, shard, blk)

	for i := 0; i < 7; i++ {
		balance, err := ldg.GetStorageItem(utils.OngContractAddress, accounts[i].Address[:])
		if err != nil {
			t.Fatalf("GetStorageItem err : %s", err)
		}
		ba := common.BigIntFromNeoBytes(balance)
		fmt.Printf("i: %d, balance : %d\n", i, ba)
	}
}

func getActivateShardTx(t *testing.T, shard common.ShardID, user string) *types.Transaction {
	param := shardmgmt.ActivateShardParam{
		ShardID: shard,
	}
	return TestCommon.CreateNativeTx2(t, user, utils.ShardMgmtContractAddress, shardmgmt.ACTIVATE_SHARD_NAME, []interface{}{param})
}

func getSetCreateFeeTx(t *testing.T, shardID common.ShardID) *types.Transaction {
	return TestCommon.CreateAdminTx2(t, shardID, utils.ShardMgmtContractAddress, shardmgmt.SET_CREATE_SHARD_FEE, []interface{}{common.BigIntToNeoBytes(big.NewInt(0))})
}
func getCreateShardtx(t *testing.T, shardID common.ShardID, acc *account.Account, user string) *types.Transaction {
	createShardParam := shardmgmt.CreateShardParam{ParentShardID: shardID, Creator: acc.Address}
	return TestCommon.CreateNativeTx2(t, user, utils.ShardMgmtContractAddress, shardmgmt.CREATE_SHARD_NAME, []interface{}{createShardParam})
}

func getConfigShardTx(t *testing.T, newShardId common.ShardID, user string) *types.Transaction {
	cfgBuff := new(bytes.Buffer)
	TestCommon.VBFT_CONFIG.Serialize(cfgBuff)
	configShardParam := shardmgmt.ConfigShardParam{
		ShardID:           newShardId,
		NetworkMin:        uint32(7),
		StakeAssetAddress: utils.OntContractAddress,
		GasAssetAddress:   utils.OngContractAddress,
		GasPrice:          uint64(0),
		GasLimit:          uint64(20000),
		VbftConfigData:    cfgBuff.Bytes(),
	}
	return TestCommon.CreateNativeTx2(t, user, utils.ShardMgmtContractAddress, shardmgmt.CONFIG_SHARD_NAME, []interface{}{configShardParam})
}
func getApplyJoinShardTx(t *testing.T, newShardId common.ShardID, user string, acc *account.Account) *types.Transaction {
	ApplyJoinShardParam := &shardmgmt.ApplyJoinShardParam{
		ShardId:    newShardId,
		PeerOwner:  acc.Address,
		PeerPubKey: common.ToHexString(keypair.SerializePublicKey(acc.PublicKey)),
	}
	return TestCommon.CreateNativeTx2(t, user, utils.ShardMgmtContractAddress, shardmgmt.APPLY_JOIN_SHARD_NAME, []interface{}{ApplyJoinShardParam})
}

func getApproveJoinShardTx(t *testing.T, newShardId common.ShardID, user string, acc []*account.Account) *types.Transaction {
	pks := make([]string, 0)
	for _,ac := range acc {
		pks = append(pks, common.ToHexString(keypair.SerializePublicKey(ac.PublicKey)))
	}
	approveJoinShardParam := shardmgmt.ApproveJoinShardParam{
		ShardId:    newShardId,
		PeerPubKey: pks,
	}
	return TestCommon.CreateNativeTx2(t, user, utils.ShardMgmtContractAddress, shardmgmt.APPROVE_JOIN_SHARD_NAME, []interface{}{approveJoinShardParam})
}

func getJoinShardTx(t *testing.T, newShardId common.ShardID, user string, acc *account.Account) *types.Transaction {
	joinShardParam := shardmgmt.JoinShardParam{
		ShardID:     newShardId,
		IpAddress:   "127.0.0.1:30338",
		PeerOwner:   acc.Address,
		PeerPubKey:  common.ToHexString(keypair.SerializePublicKey(acc.PublicKey)),
		StakeAmount: uint64(10000),
	}
	return TestCommon.CreateNativeTx2(t, user, utils.ShardMgmtContractAddress, shardmgmt.JOIN_SHARD_NAME, []interface{}{joinShardParam})
}

func checkNotify(t *testing.T, txName string, txHash common.Uint256, ldg *ledger.Ledger) {
	notify, err := ldg.GetEventNotifyByTx(txHash)
	if err != nil {
		t.Fatalf("GetEventNotifyByTx err: %s", err)
	}
	if notify == nil || notify.Notify == nil || len(notify.Notify) == 0 {
		t.Errorf("notify is nil: %s", txName)
	} else {
		fmt.Printf("txName: %s, notify: %v\n", txName, notify.Notify[0])
	}
}
