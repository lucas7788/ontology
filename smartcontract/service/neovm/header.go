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

	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	vm "github.com/ontio/ontology/vm/neovm"
)

// HeaderGetHash put header's hash to vm stack
func HeaderGetHash(service *NeoVmService, engine *vm.Executor) error {
	txHash := service.Tx.Hash()
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d,txHash:%s\n",
		"HeaderGetHash", service.Height, txHash.ToHexString())
	d, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	var data *types.Header
	if b, ok := d.Data.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.Data.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetHash] Wrong type!")
	}
	h := data.Hash()
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d,serviceTxHash:%s, hash:%s\n",
		"HeaderGetHash", service.Height, txHash.ToHexString(), h.ToHexString())
	return engine.EvalStack.PushBytes(h.ToArray())
}

// HeaderGetVersion put header's version to vm stack
func HeaderGetVersion(service *NeoVmService, engine *vm.Executor) error {
	d, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	var data *types.Header
	if b, ok := d.Data.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.Data.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetVersion] Wrong type!")
	}
	return engine.EvalStack.PushInt64(int64(data.Version))
}

// HeaderGetPrevHash put header's prevblockhash to vm stack
func HeaderGetPrevHash(service *NeoVmService, engine *vm.Executor) error {
	txHash := service.Tx.Hash()
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d,txHash:%s\n",
		"HeaderGetPrevHash", service.Height, txHash.ToHexString())
	d, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	var data *types.Header
	if b, ok := d.Data.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.Data.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetPrevHash] Wrong type!")
	}
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d,txHash:%s,preHash:%s\n",
		"HeaderGetPrevHash", service.Height, txHash.ToHexString(), data.PrevBlockHash.ToHexString())
	return engine.EvalStack.PushBytes(data.PrevBlockHash.ToArray())
}

// HeaderGetMerkleRoot put header's merkleroot to vm stack
func HeaderGetMerkleRoot(service *NeoVmService, engine *vm.Executor) error {
	txHash := service.Tx.Hash()
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d, txHash:%s\n",
		"HeaderGetMerkleRoot", service.Height, txHash.ToHexString())
	d, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	var data *types.Header
	if b, ok := d.Data.(*types.Block); ok {
		data = b.Header
		blockHash := b.Hash()
		fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d,txHash:%s,blockHash:%s\n",
			"HeaderGetMerkleRoot", service.Height, txHash.ToHexString(), blockHash.ToHexString())
	} else if h, ok := d.Data.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetMerkleRoot] Wrong type!")
	}
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d,txHash:%s, data.TransactionsRoot:%s\n",
		"HeaderGetMerkleRoot", service.Height, txHash.ToHexString(), data.TransactionsRoot.ToHexString())
	return engine.EvalStack.PushBytes(data.TransactionsRoot.ToArray())
}

// HeaderGetIndex put header's height to vm stack
func HeaderGetIndex(service *NeoVmService, engine *vm.Executor) error {
	txHash := service.Tx.Hash()
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d,txHash:%s\n",
		"HeaderGetIndex", service.Height, txHash.ToHexString())
	d, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	var data *types.Header
	if b, ok := d.Data.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.Data.(*types.Header); ok {
		data = h
	} else {
		return fmt.Errorf("[HeaderGetIndex] Wrong type")
	}
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d,serviceTxHash:%s, Height:%d\n",
		"HeaderGetIndex", service.Height, txHash.ToHexString(), data.Height)
	return engine.EvalStack.PushUint32(data.Height)
}

// HeaderGetTimestamp put header's timestamp to vm stack
func HeaderGetTimestamp(service *NeoVmService, engine *vm.Executor) error {
	txHash := service.Tx.Hash()
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d, txHash:%s\n",
		"HeaderGetTimestamp", service.Height, txHash.ToHexString())
	d, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	var data *types.Header
	if b, ok := d.Data.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.Data.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetTimestamp] Wrong type")
	}
	headerHash := data.Hash()
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d, txHash:%s, timestamp:%d, headerHash:%s\n",
		"HeaderGetTimestamp", service.Height, txHash.ToHexString(), data.Timestamp, headerHash.ToHexString())
	return engine.EvalStack.PushUint32(data.Timestamp)
}

// HeaderGetConsensusData put header's consensus data to vm stack
func HeaderGetConsensusData(service *NeoVmService, engine *vm.Executor) error {
	txHash := service.Tx.Hash()
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d, txHash:%s\n",
		"HeaderGetConsensusData", service.Height, txHash.ToHexString())
	d, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	var data *types.Header
	if b, ok := d.Data.(*types.Block); ok {
		data = b.Header
		blockHash := b.Hash()
		fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d,txHash:%s,ConsensusData:%d, blockHash:%s\n",
			"HeaderGetConsensusData", service.Height, txHash.ToHexString(), data.ConsensusData, blockHash.ToHexString())
	} else if h, ok := d.Data.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetConsensusData] Wrong type")
	}
	headerHash := data.Hash()
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d,ConsensusData:%d, headerHash:%s\n",
		"HeaderGetConsensusData", service.Height, data.ConsensusData, headerHash.ToHexString())
	return engine.EvalStack.PushUint64(data.ConsensusData)
}

// HeaderGetNextConsensus put header's consensus to vm stack
func HeaderGetNextConsensus(service *NeoVmService, engine *vm.Executor) error {
	txHash := service.Tx.Hash()
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d,txHash:%s\n",
		"HeaderGetNextConsensus", service.Height, txHash.ToHexString())
	d, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	var data *types.Header
	if b, ok := d.Data.(*types.Block); ok {
		data = b.Header
		blockHash := b.Hash()
		fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d,txHash:%s, blockHash:%s\n",
			"HeaderGetNextConsensus", service.Height, txHash.ToHexString(), blockHash.ToHexString())
	} else if h, ok := d.Data.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetNextConsensus] Wrong type")
	}
	headerHash := data.Hash()
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d,txHash:%s, headerHash:%s\n",
		"HeaderGetNextConsensus", service.Height, txHash.ToHexString(), headerHash.ToHexString())
	return engine.EvalStack.PushBytes(data.NextBookkeeper[:])
}
