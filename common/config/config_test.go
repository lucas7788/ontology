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
package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetDebugOption(t *testing.T) {
	_ = os.Setenv("debug_option", `{
		"OpenFilesCacheCapacity": 500,
		"WriteBuffer":            4194304,
		"BlockCacheCapacity":     8388608
	}`)
	opt := GetDebugOption()
	assert.Equal(t, opt.WriteBuffer, 4194304)
	assert.Equal(t, opt.BlockCacheCapacity, 8388608)

}
