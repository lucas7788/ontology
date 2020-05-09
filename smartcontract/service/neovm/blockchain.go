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

package neovm

import (
	"fmt"
	"os"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	vm "github.com/ontio/ontology/vm/neovm"
)

// BlockChainGetHeight put blockchain's height to vm stack
func BlockChainGetHeight(service *NeoVmService, engine *vm.Executor) error {
	blockHeight := service.Store.GetCurrentBlockHeight()
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d, blockHeight:%d\n",
		"BlockChainGetHeight", service.Height, blockHeight)
	err := engine.EvalStack.PushUint32(service.Store.GetCurrentBlockHeight())
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetHeight] GetHeight error!.")
	}
	return nil
}

// BlockChainGetHeader put blockchain's header to vm stack
func BlockChainGetHeader(service *NeoVmService, engine *vm.Executor) error {
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d\n", "BlockChainGetHeader", service.Height)
	var (
		header *types.Header
		err    error
	)
	data, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	l := len(data)
	if l <= 5 {
		b := common.BigIntFromNeoBytes(data)
		height := uint32(b.Int64())
		hash := service.Store.GetBlockHash(height)
		header, err = service.Store.GetHeaderByHash(hash)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetHeader] GetHeader error!.")
		}
	} else if l == 32 {
		hash, _ := common.Uint256ParseFromBytes(data)
		header, err = service.Store.GetHeaderByHash(hash)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetHeader] GetHeader error!.")
		}
	} else {
		return errors.NewErr("[BlockChainGetHeader] data invalid.")
	}
	headerHash := header.Hash()
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d, headerHash:%s\n",
		"BlockChainGetHeader", service.Height, headerHash.ToHexString())
	err = engine.EvalStack.PushAsInteropValue(header)
	if err != nil {
		return errors.NewErr("[BlockChainGetHeader] PushAsInteropValue error.")
	}
	return nil
}

// BlockChainGetBlock put blockchain's block to vm stack
func BlockChainGetBlock(service *NeoVmService, engine *vm.Executor) error {
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d\n", "BlockChainGetBlock", service.Height)
	data, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	var block *types.Block
	l := len(data)
	if l <= 5 {
		b := common.BigIntFromNeoBytes(data)
		height := uint32(b.Int64())
		var err error
		block, err = service.Store.GetBlockByHeight(height)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetBlock] GetBlock error!.")
		}
	} else if l == 32 {
		hash, err := common.Uint256ParseFromBytes(data)
		if err != nil {
			return err
		}
		block, err = service.Store.GetBlockByHash(hash)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetBlock] GetBlock error!.")
		}
	} else {
		return errors.NewErr("[BlockChainGetBlock] data invalid.")
	}
	blockHash := block.Hash()
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d, blockHash:%s\n",
		"BlockChainGetBlock", service.Height, blockHash.ToHexString())
	err = engine.EvalStack.PushAsInteropValue(block)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetBlock] PushAsInteropValue error!.")
	}
	return nil
}

// BlockChainGetTransaction put blockchain's transaction to vm stack
func BlockChainGetTransaction(service *NeoVmService, engine *vm.Executor) error {
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d\n", "BlockChainGetTransaction", service.Height)
	d, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	hash, err := common.Uint256ParseFromBytes(d)
	if err != nil {
		return err
	}
	t, _, err := service.Store.GetTransaction(hash)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetTransaction] GetTransaction error!")
	}
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d,txhash:%s\n",
		"BlockChainGetTransaction", service.Height, hash.ToHexString())
	err = engine.EvalStack.PushAsInteropValue(t)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetTransaction] PushAsInteropValue error!")
	}
	return nil
}

// BlockChainGetContract put blockchain's contract to vm stack
func BlockChainGetContract(service *NeoVmService, engine *vm.Executor) error {
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d\n", "BlockChainGetContract", service.Height)
	b, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	address, err := common.AddressParseFromBytes(b)
	if err != nil {
		return err
	}
	item, err := service.Store.GetContractState(address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetContract] GetContract error!")
	}
	err = engine.EvalStack.PushAsInteropValue(item)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetContract] PushAsInteropValue error!")
	}
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d, contractAddress:%s\n",
		"BlockChainGetContract", service.Height, address.ToHexString())
	return nil
}

// BlockChainGetTransactionHeight put transaction in block height to vm stack
func BlockChainGetTransactionHeight(service *NeoVmService, engine *vm.Executor) error {
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d\n", "BlockChainGetTransactionHeight", service.Height)
	d, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	hash, err := common.Uint256ParseFromBytes(d)
	if err != nil {
		return err
	}
	_, h, err := service.Store.GetTransaction(hash)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetTransactionHeight] GetTransaction error!")
	}
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d,txHash:%s,txHeight:%d\n",
		"BlockChainGetTransactionHeight", service.Height, hash.ToHexString(), h)
	err = engine.EvalStack.PushUint32(h)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetTransactionHeight] PushInt64 error!")
	}
	return nil
}
