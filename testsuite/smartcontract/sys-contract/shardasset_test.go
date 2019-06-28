package TestContracts

import (
	"fmt"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/chainmgr"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/shardasset/oep4"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/testsuite/common"
	"math/big"
	"testing"
)

func TestInit(t *testing.T) {
	// 1. create root chain
	shardID := common.NewShardIDUnchecked(config.DEFAULT_SHARD_ID)
	TestCommon.CreateChain(t, "test", shardID, 0)

	// 2. build shard-asset init tx
	shardName := chainmgr.GetShardName(shardID)
	user := shardName + "_peerOwner" + fmt.Sprintf("%d", 0)
	acc := TestCommon.GetAccount(user)

	//tx := TestCommon.CreateNativeTx(t, user, utils.ShardAssetAddress, oep4.INIT, nil)
	//build RegisterParam tx
	registerTx := getRegisterTx(t, acc, user)

	//build getAssetId tx
	getAssetIdTx := TestCommon.CreateNativeTx2(t, user, utils.ShardAssetAddress, oep4.ASSET_ID, []interface{}{acc.Address})

	//build Migrate tx
	migrateTx := TestCommon.CreateNativeTx2(t, user, utils.ShardAssetAddress, oep4.MIGRATE, []interface{}{acc.Address})

	// 3. create new block
	blk := TestCommon.CreateBlock(t, ledger.GetShardLedger(shardID), []*types.Transaction{registerTx, getAssetIdTx, migrateTx})

	// 4. add block
	TestCommon.ExecBlock(t, shardID, blk)
	TestCommon.SubmitBlock(t, shardID, blk)

	// 5. query db
	assetBytes := utils.GetUint64Bytes(uint64(0))
	temp := []byte(oep4.KEY_OEP4_SHARD_SUPPLY)
	temp = append(temp, assetBytes...)
	state := TestCommon.GetStorageItem(t, ledger.GetShardLedger(shardID), utils.ShardAssetAddress, temp)
	fmt.Println("state:", state)

	checkNotify(t, "registerTx", registerTx.Hash(), ledger.GetShardLedger(shardID))

}

func getRegisterTx(t *testing.T, acc *account.Account, user string) *types.Transaction {
	registerParam := oep4.RegisterParam{
		TotalSupply: big.NewInt(int64(1000)),
		Account:     acc.Address,
	}
	return TestCommon.CreateNativeTx2(t, user, utils.ShardAssetAddress, oep4.REGISTER, []interface{}{registerParam})
}
