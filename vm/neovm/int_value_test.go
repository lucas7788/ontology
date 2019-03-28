package neovm

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	vmtypes "github.com/ontio/ontology/vm/neovm/types"
	"math/big"
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
)

func randInt64() *big.Int {
	buf := make([]byte, 8)
	_, _ = rand.Read(buf)
	r := binary.LittleEndian.Uint64(buf)
	right := big.NewInt(int64(r))
	return right
}

func genBBInt() (*big.Int, *big.Int) {
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	left := common.BigIntFromNeoBytes(buf)
	_, _ = rand.Read(buf)
	right := common.BigIntFromNeoBytes(buf)
	return left, right
}

func genBLInt() (*big.Int, *big.Int) {
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	left := common.BigIntFromNeoBytes(buf)
	right := randInt64()
	return left, right
}

func genLBInt() (*big.Int, *big.Int) {
	right, left := genBLInt()
	return left, right
}

func genLLInt() (*big.Int, *big.Int) {
	left := randInt64()
	right := randInt64()
	return left, right
}

type IntOp func(left, right *big.Int) ([]byte, error)

func compareIntOpInner(t *testing.T, left, right *big.Int, func1, func2 IntOp) {
	left2 := big.NewInt(0).Set(left)
	right2 := big.NewInt(0).Set(right)
	val1, err := func1(left, right)
	val2, err2 := func2(left2, right2)
	fmt.Println("val1:", val1)
	fmt.Println("val2:", val2)
	if err != nil || err2 != nil {
		return
	}
	assert.Equal(t, val1, val2)
}

func compareIntOp(t *testing.T, func1, func2 IntOp) {
	const N = 100000
	for i := 0; i < N; i++ {
		left := big.NewInt(0).SetInt64(-6615735770392487332)
		right := big.NewInt(0).SetInt64(6275592772156762005)
		compareIntOpInner(t, left, right, func1, func2)
		left, right = genLLInt()
		compareIntOpInner(t, left, right, func1, func2)
		left, right = genBLInt()
		compareIntOpInner(t, left, right, func1, func2)
		left, right = genLBInt()
		compareIntOpInner(t, left, right, func1, func2)
	}
}

func TestIntValue_Abs(t *testing.T) {
	compareIntOp(t, func(left, right *big.Int) ([]byte, error) {
		abs := big.NewInt(0).Abs(left)
		return common.BigIntToNeoBytes(abs), nil
	}, func(left, right *big.Int) ([]byte, error) {
		val, err := vmtypes.IntValFromBigInt(left)
		assert.Nil(t, err)
		val = val.Abs()

		return val.ToNeoBytes(), nil
	})
}

func TestIntValue_Add(t *testing.T) {
	compareIntOp(t, func(left, right *big.Int) ([]byte, error) {
		val := big.NewInt(0).Add(left, right)
		return common.BigIntToNeoBytes(val), nil
	}, func(left, right *big.Int) ([]byte, error) {
		lhs, err := vmtypes.IntValFromBigInt(left)
		if err != nil {
			return nil, err
		}
		rhs, err := vmtypes.IntValFromBigInt(right)
		if err != nil {
			return nil, err
		}
		val, err := lhs.Add(rhs)
		if err != nil {
			return nil, err
		}

		return val.ToNeoBytes(), nil
	})
}

func TestIntValue_Sub(t *testing.T) {
	compareIntOp(t, func(left, right *big.Int) ([]byte, error) {
		val := big.NewInt(0).Sub(left, right)
		return common.BigIntToNeoBytes(val), nil
	}, func(left, right *big.Int) ([]byte, error) {
		lhs, err := vmtypes.IntValFromBigInt(left)
		if err != nil {
			return nil, err
		}
		rhs, err := vmtypes.IntValFromBigInt(right)
		if err != nil {
			return nil, err
		}
		val, err := lhs.Sub(rhs)
		if err != nil {
			return nil, err
		}

		return val.ToNeoBytes(), nil
	})
}

func TestIntValue_Mul(t *testing.T) {
	compareIntOp(t, func(left, right *big.Int) ([]byte, error) {
		val := big.NewInt(0).Mul(left, right)
		return common.BigIntToNeoBytes(val), nil
	}, func(left, right *big.Int) ([]byte, error) {
		lhs, err := vmtypes.IntValFromBigInt(left)
		if err != nil {
			return nil, err
		}
		rhs, err := vmtypes.IntValFromBigInt(right)
		if err != nil {
			return nil, err
		}
		val, err := lhs.Mul(rhs)
		if err != nil {
			return nil, err
		}

		return val.ToNeoBytes(), nil
	})
}

func TestIntValue_Div(t *testing.T) {
	compareIntOp(t, func(left, right *big.Int) ([]byte, error) {
		val := big.NewInt(0).Quo(left, right)
		return common.BigIntToNeoBytes(val), nil
	}, func(left, right *big.Int) ([]byte, error) {
		lhs, err := vmtypes.IntValFromBigInt(left)
		if err != nil {
			return nil, err
		}
		rhs, err := vmtypes.IntValFromBigInt(right)
		if err != nil {
			return nil, err
		}
		val, err := lhs.Div(rhs)
		if err != nil {
			return nil, err
		}

		return val.ToNeoBytes(), nil
	})
}

func TestIntValue_Mod(t *testing.T) {
	compareIntOp(t, func(left, right *big.Int) ([]byte, error) {
		val := big.NewInt(0).Rem(left, right)
		return common.BigIntToNeoBytes(val), nil
	}, func(left, right *big.Int) ([]byte, error) {
		lhs, err := vmtypes.IntValFromBigInt(left)
		if err != nil {
			return nil, err
		}
		rhs, err := vmtypes.IntValFromBigInt(right)
		if err != nil {
			return nil, err
		}
		val, err := lhs.Mod(rhs)
		if err != nil {
			return nil, err
		}

		return val.ToNeoBytes(), nil
	})
}

func TestIntValue_Rsh(t *testing.T) {
	compareIntOp(t, func(left, right *big.Int) ([]byte, error) {
		b := BigIntZip(left, right, SHR)
		return common.BigIntToNeoBytes(b), nil
	}, func(left, right *big.Int) ([]byte, error) {
		lhs, err := vmtypes.IntValFromBigInt(left)
		if err != nil {
			return nil, err
		}
		rhs, err := vmtypes.IntValFromBigInt(right)
		if err != nil {
			return nil, err
		}
		val, err := lhs.Rsh(rhs)
		if err != nil {
			return nil, err
		}

		return val.ToNeoBytes(), nil
	})
}
