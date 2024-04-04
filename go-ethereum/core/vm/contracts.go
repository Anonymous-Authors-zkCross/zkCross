// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package vm

import "C"
import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net"
	"strings"
	"time"

	"github.com/Nik-U/pbc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/blake2b"
	"github.com/ethereum/go-ethereum/crypto/bls12381"
	"github.com/ethereum/go-ethereum/crypto/bn256"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/zktx"
	"golang.org/x/crypto/ripemd160"
	//lint:ignore SA1019 Needed for precompile
	//"golang.org/x/crypto/ripemd160"
)

// PrecompiledContract is the basic interface for native Go contracts. The implementation
// requires a deterministic gas count based on the input size of the Run method of the
// contract.
type PrecompiledContract interface {
	RequiredGas(input []byte) uint64  // RequiredPrice calculates the contract gas use
	Run(input []byte) ([]byte, error) // Run runs the precompiled contract
}

// gyh : veriGroupsign+proof
// PrecompiledContractsHomestead contains the default set of pre-compiled Ethereum
// contracts used in the Frontier and Homestead releases.
var PrecompiledContractsHomestead = map[common.Address]PrecompiledContract{
	common.BytesToAddress([]byte{1}):  &ecrecover{},
	common.BytesToAddress([]byte{2}):  &sha256hash{},
	common.BytesToAddress([]byte{3}):  &ripemd160hash{},
	common.BytesToAddress([]byte{4}):  &dataCopy{},
	common.BytesToAddress([]byte{19}): &veriGroupsign{}, //gyh :返回值必须是ｂｙｔｅ
	common.BytesToAddress([]byte{20}): &verhfProof{},
}

// PrecompiledContractsByzantium contains the default set of pre-compiled Ethereum
// contracts used in the Byzantium release.
var PrecompiledContractsByzantium = map[common.Address]PrecompiledContract{
	common.BytesToAddress([]byte{1}):  &ecrecover{},
	common.BytesToAddress([]byte{2}):  &sha256hash{},
	common.BytesToAddress([]byte{3}):  &ripemd160hash{},
	common.BytesToAddress([]byte{4}):  &dataCopy{},
	common.BytesToAddress([]byte{5}):  &bigModExp{eip2565: false},
	common.BytesToAddress([]byte{6}):  &bn256AddByzantium{},
	common.BytesToAddress([]byte{7}):  &bn256ScalarMulByzantium{},
	common.BytesToAddress([]byte{8}):  &bn256PairingByzantium{},
	common.BytesToAddress([]byte{19}): &veriGroupsign{},
	common.BytesToAddress([]byte{20}): &verhfProof{},
}

// PrecompiledContractsIstanbul contains the default set of pre-compiled Ethereum
// contracts used in the Istanbul release.
var PrecompiledContractsIstanbul = map[common.Address]PrecompiledContract{
	common.BytesToAddress([]byte{1}):  &ecrecover{},
	common.BytesToAddress([]byte{2}):  &sha256hash{},
	common.BytesToAddress([]byte{3}):  &ripemd160hash{},
	common.BytesToAddress([]byte{4}):  &dataCopy{},
	common.BytesToAddress([]byte{5}):  &bigModExp{eip2565: false},
	common.BytesToAddress([]byte{6}):  &bn256AddIstanbul{},
	common.BytesToAddress([]byte{7}):  &bn256ScalarMulIstanbul{},
	common.BytesToAddress([]byte{8}):  &bn256PairingIstanbul{},
	common.BytesToAddress([]byte{9}):  &blake2F{},
	common.BytesToAddress([]byte{19}): &veriGroupsign{},
	common.BytesToAddress([]byte{20}): &verhfProof{},
}

// PrecompiledContractsYoloV2 contains the default set of pre-compiled Ethereum
// contracts used in the Yolo v2 test release.
var PrecompiledContractsYoloV2 = map[common.Address]PrecompiledContract{
	common.BytesToAddress([]byte{1}):  &ecrecover{},
	common.BytesToAddress([]byte{2}):  &sha256hash{},
	common.BytesToAddress([]byte{3}):  &ripemd160hash{},
	common.BytesToAddress([]byte{4}):  &dataCopy{},
	common.BytesToAddress([]byte{5}):  &bigModExp{eip2565: false},
	common.BytesToAddress([]byte{6}):  &bn256AddIstanbul{},
	common.BytesToAddress([]byte{7}):  &bn256ScalarMulIstanbul{},
	common.BytesToAddress([]byte{8}):  &bn256PairingIstanbul{},
	common.BytesToAddress([]byte{9}):  &blake2F{},
	common.BytesToAddress([]byte{10}): &bls12381G1Add{},
	common.BytesToAddress([]byte{11}): &bls12381G1Mul{},
	common.BytesToAddress([]byte{12}): &bls12381G1MultiExp{},
	common.BytesToAddress([]byte{13}): &bls12381G2Add{},
	common.BytesToAddress([]byte{14}): &bls12381G2Mul{},
	common.BytesToAddress([]byte{15}): &bls12381G2MultiExp{},
	common.BytesToAddress([]byte{16}): &bls12381Pairing{},
	common.BytesToAddress([]byte{17}): &bls12381MapG1{},
	common.BytesToAddress([]byte{18}): &bls12381MapG2{},
	common.BytesToAddress([]byte{19}): &veriGroupsign{},
	common.BytesToAddress([]byte{20}): &verhfProof{},
}

var (
	PrecompiledAddressesYoloV2    []common.Address
	PrecompiledAddressesIstanbul  []common.Address
	PrecompiledAddressesByzantium []common.Address
	PrecompiledAddressesHomestead []common.Address
)

func init() {
	for k := range PrecompiledContractsHomestead {
		PrecompiledAddressesHomestead = append(PrecompiledAddressesHomestead, k)
	}
	for k := range PrecompiledContractsByzantium {
		PrecompiledAddressesHomestead = append(PrecompiledAddressesByzantium, k)
	}
	for k := range PrecompiledContractsIstanbul {
		PrecompiledAddressesIstanbul = append(PrecompiledAddressesIstanbul, k)
	}
	for k := range PrecompiledContractsYoloV2 {
		PrecompiledAddressesYoloV2 = append(PrecompiledAddressesYoloV2, k)
	}
}

// RunPrecompiledContract runs and evaluates the output of a precompiled contract.
// It returns
// - the returned bytes,
// - the _remaining_ gas,
// - any error that occurred
func RunPrecompiledContract(p PrecompiledContract, input []byte, suppliedGas uint64) (ret []byte, remainingGas uint64, err error) {
	gasCost := p.RequiredGas(input)
	if suppliedGas < gasCost {
		return nil, 0, ErrOutOfGas
	}
	suppliedGas -= gasCost
	output, err := p.Run(input)
	return output, suppliedGas, err
}

// ECRECOVER implemented as a native contract.
type ecrecover struct{}

func (c *ecrecover) RequiredGas(input []byte) uint64 {
	return params.EcrecoverGas
}

func (c *ecrecover) Run(input []byte) ([]byte, error) {
	const ecRecoverInputLength = 128

	input = common.RightPadBytes(input, ecRecoverInputLength)
	// "input" is (hash, v, r, s), each 32 bytes
	// but for ecrecover we want (r, s, v)

	r := new(big.Int).SetBytes(input[64:96])
	s := new(big.Int).SetBytes(input[96:128])
	v := input[63] - 27

	// tighter sig s values input homestead only apply to tx sigs
	if !allZero(input[32:63]) || !crypto.ValidateSignatureValues(v, r, s, false) {
		return nil, nil
	}
	// We must make sure not to modify the 'input', so placing the 'v' along with
	// the signature needs to be done on a new allocation
	sig := make([]byte, 65)
	copy(sig, input[64:128])
	sig[64] = v
	// v needs to be at the end for libsecp256k1
	pubKey, err := crypto.Ecrecover(input[:32], sig)
	// make sure the public key is a valid one
	if err != nil {
		return nil, nil
	}

	// the first byte of pubkey is bitcoin heritage
	return common.LeftPadBytes(crypto.Keccak256(pubKey[1:])[12:], 32), nil
}

// SHA256 implemented as a native contract.
type sha256hash struct{}

// RequiredGas returns the gas required to execute the pre-compiled contract.
//
// This method does not require any overflow checking as the input size gas costs
// required for anything significant is so high it's impossible to pay for.
func (c *sha256hash) RequiredGas(input []byte) uint64 {
	return uint64(len(input)+31)/32*params.Sha256PerWordGas + params.Sha256BaseGas
}
func (c *sha256hash) Run(input []byte) ([]byte, error) {
	h := sha256.Sum256(input)
	return h[:], nil
}

// RIPEMD160 implemented as a native contract.
type ripemd160hash struct{}

// RequiredGas returns the gas required to execute the pre-compiled contract.
//
// This method does not require any overflow checking as the input size gas costs
// required for anything significant is so high it's impossible to pay for.
func (c *ripemd160hash) RequiredGas(input []byte) uint64 {
	return uint64(len(input)+31)/32*params.Ripemd160PerWordGas + params.Ripemd160BaseGas
}
func (c *ripemd160hash) Run(input []byte) ([]byte, error) {
	ripemd := ripemd160.New()
	ripemd.Write(input)
	return common.LeftPadBytes(ripemd.Sum(nil), 32), nil
}

// data copy implemented as a native contract.
type dataCopy struct{}

// RequiredGas returns the gas required to execute the pre-compiled contract.
//
// This method does not require any overflow checking as the input size gas costs
// required for anything significant is so high it's impossible to pay for.
func (c *dataCopy) RequiredGas(input []byte) uint64 {
	return uint64(len(input)+31)/32*params.IdentityPerWordGas + params.IdentityBaseGas
}
func (c *dataCopy) Run(in []byte) ([]byte, error) {
	return in, nil
}

// bigModExp implements a native big integer exponential modular operation.
type bigModExp struct {
	eip2565 bool
}

var (
	big0      = big.NewInt(0)
	big1      = big.NewInt(1)
	big3      = big.NewInt(3)
	big4      = big.NewInt(4)
	big7      = big.NewInt(7)
	big8      = big.NewInt(8)
	big16     = big.NewInt(16)
	big20     = big.NewInt(20)
	big32     = big.NewInt(32)
	big64     = big.NewInt(64)
	big96     = big.NewInt(96)
	big480    = big.NewInt(480)
	big1024   = big.NewInt(1024)
	big3072   = big.NewInt(3072)
	big199680 = big.NewInt(199680)
)

// modexpMultComplexity implements bigModexp multComplexity formula, as defined in EIP-198
//
// def mult_complexity(x):
//    if x <= 64: return x ** 2
//    elif x <= 1024: return x ** 2 // 4 + 96 * x - 3072
//    else: return x ** 2 // 16 + 480 * x - 199680
//
// where is x is max(length_of_MODULUS, length_of_BASE)
func modexpMultComplexity(x *big.Int) *big.Int {
	switch {
	case x.Cmp(big64) <= 0:
		x.Mul(x, x) // x ** 2
	case x.Cmp(big1024) <= 0:
		// (x ** 2 // 4 ) + ( 96 * x - 3072)
		x = new(big.Int).Add(
			new(big.Int).Div(new(big.Int).Mul(x, x), big4),
			new(big.Int).Sub(new(big.Int).Mul(big96, x), big3072),
		)
	default:
		// (x ** 2 // 16) + (480 * x - 199680)
		x = new(big.Int).Add(
			new(big.Int).Div(new(big.Int).Mul(x, x), big16),
			new(big.Int).Sub(new(big.Int).Mul(big480, x), big199680),
		)
	}
	return x
}

// RequiredGas returns the gas required to execute the pre-compiled contract.
func (c *bigModExp) RequiredGas(input []byte) uint64 {
	var (
		baseLen = new(big.Int).SetBytes(getData(input, 0, 32))
		expLen  = new(big.Int).SetBytes(getData(input, 32, 32))
		modLen  = new(big.Int).SetBytes(getData(input, 64, 32))
	)
	if len(input) > 96 {
		input = input[96:]
	} else {
		input = input[:0]
	}
	// Retrieve the head 32 bytes of exp for the adjusted exponent length
	var expHead *big.Int
	if big.NewInt(int64(len(input))).Cmp(baseLen) <= 0 {
		expHead = new(big.Int)
	} else {
		if expLen.Cmp(big32) > 0 {
			expHead = new(big.Int).SetBytes(getData(input, baseLen.Uint64(), 32))
		} else {
			expHead = new(big.Int).SetBytes(getData(input, baseLen.Uint64(), expLen.Uint64()))
		}
	}
	// Calculate the adjusted exponent length
	var msb int
	if bitlen := expHead.BitLen(); bitlen > 0 {
		msb = bitlen - 1
	}
	adjExpLen := new(big.Int)
	if expLen.Cmp(big32) > 0 {
		adjExpLen.Sub(expLen, big32)
		adjExpLen.Mul(big8, adjExpLen)
	}
	adjExpLen.Add(adjExpLen, big.NewInt(int64(msb)))
	// Calculate the gas cost of the operation
	gas := new(big.Int).Set(math.BigMax(modLen, baseLen))
	if c.eip2565 {
		// EIP-2565 has three changes
		// 1. Different multComplexity (inlined here)
		// in EIP-2565 (https://eips.ethereum.org/EIPS/eip-2565):
		//
		// def mult_complexity(x):
		//    ceiling(x/8)^2
		//
		//where is x is max(length_of_MODULUS, length_of_BASE)
		gas = gas.Add(gas, big7)
		gas = gas.Div(gas, big8)
		gas.Mul(gas, gas)

		gas.Mul(gas, math.BigMax(adjExpLen, big1))
		// 2. Different divisor (`GQUADDIVISOR`) (3)
		gas.Div(gas, big3)
		if gas.BitLen() > 64 {
			return math.MaxUint64
		}
		// 3. Minimum price of 200 gas
		if gas.Uint64() < 200 {
			return 200
		}
		return gas.Uint64()
	}
	gas = modexpMultComplexity(gas)
	gas.Mul(gas, math.BigMax(adjExpLen, big1))
	gas.Div(gas, big20)

	if gas.BitLen() > 64 {
		return math.MaxUint64
	}
	return gas.Uint64()
}

func (c *bigModExp) Run(input []byte) ([]byte, error) {
	var (
		baseLen = new(big.Int).SetBytes(getData(input, 0, 32)).Uint64()
		expLen  = new(big.Int).SetBytes(getData(input, 32, 32)).Uint64()
		modLen  = new(big.Int).SetBytes(getData(input, 64, 32)).Uint64()
	)
	if len(input) > 96 {
		input = input[96:]
	} else {
		input = input[:0]
	}
	// Handle a special case when both the base and mod length is zero
	if baseLen == 0 && modLen == 0 {
		return []byte{}, nil
	}
	// Retrieve the operands and execute the exponentiation
	var (
		base = new(big.Int).SetBytes(getData(input, 0, baseLen))
		exp  = new(big.Int).SetBytes(getData(input, baseLen, expLen))
		mod  = new(big.Int).SetBytes(getData(input, baseLen+expLen, modLen))
	)
	if mod.BitLen() == 0 {
		// Modulo 0 is undefined, return zero
		return common.LeftPadBytes([]byte{}, int(modLen)), nil
	}
	return common.LeftPadBytes(base.Exp(base, exp, mod).Bytes(), int(modLen)), nil
}

// newCurvePoint unmarshals a binary blob into a bn256 elliptic curve point,
// returning it, or an error if the point is invalid.
func newCurvePoint(blob []byte) (*bn256.G1, error) {
	p := new(bn256.G1)
	if _, err := p.Unmarshal(blob); err != nil {
		return nil, err
	}
	return p, nil
}

// newTwistPoint unmarshals a binary blob into a bn256 elliptic curve point,
// returning it, or an error if the point is invalid.
func newTwistPoint(blob []byte) (*bn256.G2, error) {
	p := new(bn256.G2)
	if _, err := p.Unmarshal(blob); err != nil {
		return nil, err
	}
	return p, nil
}

// runBn256Add implements the Bn256Add precompile, referenced by both
// Byzantium and Istanbul operations.
func runBn256Add(input []byte) ([]byte, error) {
	x, err := newCurvePoint(getData(input, 0, 64))
	if err != nil {
		return nil, err
	}
	y, err := newCurvePoint(getData(input, 64, 64))
	if err != nil {
		return nil, err
	}
	res := new(bn256.G1)
	res.Add(x, y)
	return res.Marshal(), nil
}

// bn256Add implements a native elliptic curve point addition conforming to
// Istanbul consensus rules.
type bn256AddIstanbul struct{}

// RequiredGas returns the gas required to execute the pre-compiled contract.
func (c *bn256AddIstanbul) RequiredGas(input []byte) uint64 {
	return params.Bn256AddGasIstanbul
}

func (c *bn256AddIstanbul) Run(input []byte) ([]byte, error) {
	return runBn256Add(input)
}

// bn256AddByzantium implements a native elliptic curve point addition
// conforming to Byzantium consensus rules.
type bn256AddByzantium struct{}

// RequiredGas returns the gas required to execute the pre-compiled contract.
func (c *bn256AddByzantium) RequiredGas(input []byte) uint64 {
	return params.Bn256AddGasByzantium
}

func (c *bn256AddByzantium) Run(input []byte) ([]byte, error) {
	return runBn256Add(input)
}

// runBn256ScalarMul implements the Bn256ScalarMul precompile, referenced by
// both Byzantium and Istanbul operations.
func runBn256ScalarMul(input []byte) ([]byte, error) {
	p, err := newCurvePoint(getData(input, 0, 64))
	if err != nil {
		return nil, err
	}
	res := new(bn256.G1)
	res.ScalarMult(p, new(big.Int).SetBytes(getData(input, 64, 32)))
	return res.Marshal(), nil
}

// bn256ScalarMulIstanbul implements a native elliptic curve scalar
// multiplication conforming to Istanbul consensus rules.
type bn256ScalarMulIstanbul struct{}

// RequiredGas returns the gas required to execute the pre-compiled contract.
func (c *bn256ScalarMulIstanbul) RequiredGas(input []byte) uint64 {
	return params.Bn256ScalarMulGasIstanbul
}

func (c *bn256ScalarMulIstanbul) Run(input []byte) ([]byte, error) {
	return runBn256ScalarMul(input)
}

// bn256ScalarMulByzantium implements a native elliptic curve scalar
// multiplication conforming to Byzantium consensus rules.
type bn256ScalarMulByzantium struct{}

// RequiredGas returns the gas required to execute the pre-compiled contract.
func (c *bn256ScalarMulByzantium) RequiredGas(input []byte) uint64 {
	return params.Bn256ScalarMulGasByzantium
}

func (c *bn256ScalarMulByzantium) Run(input []byte) ([]byte, error) {
	return runBn256ScalarMul(input)
}

var (
	// true32Byte is returned if the bn256 pairing check succeeds.
	true32Byte = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}

	// false32Byte is returned if the bn256 pairing check fails.
	false32Byte = make([]byte, 32)

	// errBadPairingInput is returned if the bn256 pairing input is invalid.
	errBadPairingInput = errors.New("bad elliptic curve pairing size")
)

// runBn256Pairing implements the Bn256Pairing precompile, referenced by both
// Byzantium and Istanbul operations.
func runBn256Pairing(input []byte) ([]byte, error) {
	// Handle some corner cases cheaply
	if len(input)%192 > 0 {
		return nil, errBadPairingInput
	}
	// Convert the input into a set of coordinates
	var (
		cs []*bn256.G1
		ts []*bn256.G2
	)
	for i := 0; i < len(input); i += 192 {
		c, err := newCurvePoint(input[i : i+64])
		if err != nil {
			return nil, err
		}
		t, err := newTwistPoint(input[i+64 : i+192])
		if err != nil {
			return nil, err
		}
		cs = append(cs, c)
		ts = append(ts, t)
	}
	// Execute the pairing checks and return the results
	if bn256.PairingCheck(cs, ts) {
		return true32Byte, nil
	}
	return false32Byte, nil
}

// bn256PairingIstanbul implements a pairing pre-compile for the bn256 curve
// conforming to Istanbul consensus rules.
type bn256PairingIstanbul struct{}

// RequiredGas returns the gas required to execute the pre-compiled contract.
func (c *bn256PairingIstanbul) RequiredGas(input []byte) uint64 {
	return params.Bn256PairingBaseGasIstanbul + uint64(len(input)/192)*params.Bn256PairingPerPointGasIstanbul
}

func (c *bn256PairingIstanbul) Run(input []byte) ([]byte, error) {
	return runBn256Pairing(input)
}

// bn256PairingByzantium implements a pairing pre-compile for the bn256 curve
// conforming to Byzantium consensus rules.
type bn256PairingByzantium struct{}

// RequiredGas returns the gas required to execute the pre-compiled contract.
func (c *bn256PairingByzantium) RequiredGas(input []byte) uint64 {
	return params.Bn256PairingBaseGasByzantium + uint64(len(input)/192)*params.Bn256PairingPerPointGasByzantium
}

func (c *bn256PairingByzantium) Run(input []byte) ([]byte, error) {
	return runBn256Pairing(input)
}

type blake2F struct{}

func (c *blake2F) RequiredGas(input []byte) uint64 {
	// If the input is malformed, we can't calculate the gas, return 0 and let the
	// actual call choke and fault.
	if len(input) != blake2FInputLength {
		return 0
	}
	return uint64(binary.BigEndian.Uint32(input[0:4]))
}

const (
	blake2FInputLength        = 213
	blake2FFinalBlockBytes    = byte(1)
	blake2FNonFinalBlockBytes = byte(0)
)

var (
	errBlake2FInvalidInputLength = errors.New("invalid input length")
	errBlake2FInvalidFinalFlag   = errors.New("invalid final flag")
)

func (c *blake2F) Run(input []byte) ([]byte, error) {
	// Make sure the input is valid (correct length and final flag)
	if len(input) != blake2FInputLength {
		return nil, errBlake2FInvalidInputLength
	}
	if input[212] != blake2FNonFinalBlockBytes && input[212] != blake2FFinalBlockBytes {
		return nil, errBlake2FInvalidFinalFlag
	}
	// Parse the input into the Blake2b call parameters
	var (
		rounds = binary.BigEndian.Uint32(input[0:4])
		final  = (input[212] == blake2FFinalBlockBytes)

		h [8]uint64
		m [16]uint64
		t [2]uint64
	)
	for i := 0; i < 8; i++ {
		offset := 4 + i*8
		h[i] = binary.LittleEndian.Uint64(input[offset : offset+8])
	}
	for i := 0; i < 16; i++ {
		offset := 68 + i*8
		m[i] = binary.LittleEndian.Uint64(input[offset : offset+8])
	}
	t[0] = binary.LittleEndian.Uint64(input[196:204])
	t[1] = binary.LittleEndian.Uint64(input[204:212])

	// Execute the compression function, extract and return the result
	blake2b.F(&h, m, t, final, rounds)

	output := make([]byte, 64)
	for i := 0; i < 8; i++ {
		offset := i * 8
		binary.LittleEndian.PutUint64(output[offset:offset+8], h[i])
	}
	return output, nil
}

var (
	errBLS12381InvalidInputLength          = errors.New("invalid input length")
	errBLS12381InvalidFieldElementTopBytes = errors.New("invalid field element top bytes")
	errBLS12381G1PointSubgroup             = errors.New("g1 point is not on correct subgroup")
	errBLS12381G2PointSubgroup             = errors.New("g2 point is not on correct subgroup")
)

// bls12381G1Add implements EIP-2537 G1Add precompile.
type bls12381G1Add struct{}

// RequiredGas returns the gas required to execute the pre-compiled contract.
func (c *bls12381G1Add) RequiredGas(input []byte) uint64 {
	return params.Bls12381G1AddGas
}

func (c *bls12381G1Add) Run(input []byte) ([]byte, error) {
	// Implements EIP-2537 G1Add precompile.
	// > G1 addition call expects `256` bytes as an input that is interpreted as byte concatenation of two G1 points (`128` bytes each).
	// > Output is an encoding of addition operation result - single G1 point (`128` bytes).
	if len(input) != 256 {
		return nil, errBLS12381InvalidInputLength
	}
	var err error
	var p0, p1 *bls12381.PointG1

	// Initialize G1
	g := bls12381.NewG1()

	// Decode G1 point p_0
	if p0, err = g.DecodePoint(input[:128]); err != nil {
		return nil, err
	}
	// Decode G1 point p_1
	if p1, err = g.DecodePoint(input[128:]); err != nil {
		return nil, err
	}

	// Compute r = p_0 + p_1
	r := g.New()
	g.Add(r, p0, p1)

	// Encode the G1 point result into 128 bytes
	return g.EncodePoint(r), nil
}

// bls12381G1Mul implements EIP-2537 G1Mul precompile.
type bls12381G1Mul struct{}

// RequiredGas returns the gas required to execute the pre-compiled contract.
func (c *bls12381G1Mul) RequiredGas(input []byte) uint64 {
	return params.Bls12381G1MulGas
}

func (c *bls12381G1Mul) Run(input []byte) ([]byte, error) {
	// Implements EIP-2537 G1Mul precompile.
	// > G1 multiplication call expects `160` bytes as an input that is interpreted as byte concatenation of encoding of G1 point (`128` bytes) and encoding of a scalar value (`32` bytes).
	// > Output is an encoding of multiplication operation result - single G1 point (`128` bytes).
	if len(input) != 160 {
		return nil, errBLS12381InvalidInputLength
	}
	var err error
	var p0 *bls12381.PointG1

	// Initialize G1
	g := bls12381.NewG1()

	// Decode G1 point
	if p0, err = g.DecodePoint(input[:128]); err != nil {
		return nil, err
	}
	// Decode scalar value
	e := new(big.Int).SetBytes(input[128:])

	// Compute r = e * p_0
	r := g.New()
	g.MulScalar(r, p0, e)

	// Encode the G1 point into 128 bytes
	return g.EncodePoint(r), nil
}

// bls12381G1MultiExp implements EIP-2537 G1MultiExp precompile.
type bls12381G1MultiExp struct{}

// RequiredGas returns the gas required to execute the pre-compiled contract.
func (c *bls12381G1MultiExp) RequiredGas(input []byte) uint64 {
	// Calculate G1 point, scalar value pair length
	k := len(input) / 160
	if k == 0 {
		// Return 0 gas for small input length
		return 0
	}
	// Lookup discount value for G1 point, scalar value pair length
	var discount uint64
	if dLen := len(params.Bls12381MultiExpDiscountTable); k < dLen {
		discount = params.Bls12381MultiExpDiscountTable[k-1]
	} else {
		discount = params.Bls12381MultiExpDiscountTable[dLen-1]
	}
	// Calculate gas and return the result
	return (uint64(k) * params.Bls12381G1MulGas * discount) / 1000
}

func (c *bls12381G1MultiExp) Run(input []byte) ([]byte, error) {
	// Implements EIP-2537 G1MultiExp precompile.
	// G1 multiplication call expects `160*k` bytes as an input that is interpreted as byte concatenation of `k` slices each of them being a byte concatenation of encoding of G1 point (`128` bytes) and encoding of a scalar value (`32` bytes).
	// Output is an encoding of multiexponentiation operation result - single G1 point (`128` bytes).
	k := len(input) / 160
	if len(input) == 0 || len(input)%160 != 0 {
		return nil, errBLS12381InvalidInputLength
	}
	var err error
	points := make([]*bls12381.PointG1, k)
	scalars := make([]*big.Int, k)

	// Initialize G1
	g := bls12381.NewG1()

	// Decode point scalar pairs
	for i := 0; i < k; i++ {
		off := 160 * i
		t0, t1, t2 := off, off+128, off+160
		// Decode G1 point
		if points[i], err = g.DecodePoint(input[t0:t1]); err != nil {
			return nil, err
		}
		// Decode scalar value
		scalars[i] = new(big.Int).SetBytes(input[t1:t2])
	}

	// Compute r = e_0 * p_0 + e_1 * p_1 + ... + e_(k-1) * p_(k-1)
	r := g.New()
	g.MultiExp(r, points, scalars)

	// Encode the G1 point to 128 bytes
	return g.EncodePoint(r), nil
}

// bls12381G2Add implements EIP-2537 G2Add precompile.
type bls12381G2Add struct{}

// RequiredGas returns the gas required to execute the pre-compiled contract.
func (c *bls12381G2Add) RequiredGas(input []byte) uint64 {
	return params.Bls12381G2AddGas
}

func (c *bls12381G2Add) Run(input []byte) ([]byte, error) {
	// Implements EIP-2537 G2Add precompile.
	// > G2 addition call expects `512` bytes as an input that is interpreted as byte concatenation of two G2 points (`256` bytes each).
	// > Output is an encoding of addition operation result - single G2 point (`256` bytes).
	if len(input) != 512 {
		return nil, errBLS12381InvalidInputLength
	}
	var err error
	var p0, p1 *bls12381.PointG2

	// Initialize G2
	g := bls12381.NewG2()
	r := g.New()

	// Decode G2 point p_0
	if p0, err = g.DecodePoint(input[:256]); err != nil {
		return nil, err
	}
	// Decode G2 point p_1
	if p1, err = g.DecodePoint(input[256:]); err != nil {
		return nil, err
	}

	// Compute r = p_0 + p_1
	g.Add(r, p0, p1)

	// Encode the G2 point into 256 bytes
	return g.EncodePoint(r), nil
}

// bls12381G2Mul implements EIP-2537 G2Mul precompile.
type bls12381G2Mul struct{}

// RequiredGas returns the gas required to execute the pre-compiled contract.
func (c *bls12381G2Mul) RequiredGas(input []byte) uint64 {
	return params.Bls12381G2MulGas
}

func (c *bls12381G2Mul) Run(input []byte) ([]byte, error) {
	// Implements EIP-2537 G2MUL precompile logic.
	// > G2 multiplication call expects `288` bytes as an input that is interpreted as byte concatenation of encoding of G2 point (`256` bytes) and encoding of a scalar value (`32` bytes).
	// > Output is an encoding of multiplication operation result - single G2 point (`256` bytes).
	if len(input) != 288 {
		return nil, errBLS12381InvalidInputLength
	}
	var err error
	var p0 *bls12381.PointG2

	// Initialize G2
	g := bls12381.NewG2()

	// Decode G2 point
	if p0, err = g.DecodePoint(input[:256]); err != nil {
		return nil, err
	}
	// Decode scalar value
	e := new(big.Int).SetBytes(input[256:])

	// Compute r = e * p_0
	r := g.New()
	g.MulScalar(r, p0, e)

	// Encode the G2 point into 256 bytes
	return g.EncodePoint(r), nil
}

// bls12381G2MultiExp implements EIP-2537 G2MultiExp precompile.
type bls12381G2MultiExp struct{}

// RequiredGas returns the gas required to execute the pre-compiled contract.
func (c *bls12381G2MultiExp) RequiredGas(input []byte) uint64 {
	// Calculate G2 point, scalar value pair length
	k := len(input) / 288
	if k == 0 {
		// Return 0 gas for small input length
		return 0
	}
	// Lookup discount value for G2 point, scalar value pair length
	var discount uint64
	if dLen := len(params.Bls12381MultiExpDiscountTable); k < dLen {
		discount = params.Bls12381MultiExpDiscountTable[k-1]
	} else {
		discount = params.Bls12381MultiExpDiscountTable[dLen-1]
	}
	// Calculate gas and return the result
	return (uint64(k) * params.Bls12381G2MulGas * discount) / 1000
}

func (c *bls12381G2MultiExp) Run(input []byte) ([]byte, error) {
	// Implements EIP-2537 G2MultiExp precompile logic
	// > G2 multiplication call expects `288*k` bytes as an input that is interpreted as byte concatenation of `k` slices each of them being a byte concatenation of encoding of G2 point (`256` bytes) and encoding of a scalar value (`32` bytes).
	// > Output is an encoding of multiexponentiation operation result - single G2 point (`256` bytes).
	k := len(input) / 288
	if len(input) == 0 || len(input)%288 != 0 {
		return nil, errBLS12381InvalidInputLength
	}
	var err error
	points := make([]*bls12381.PointG2, k)
	scalars := make([]*big.Int, k)

	// Initialize G2
	g := bls12381.NewG2()

	// Decode point scalar pairs
	for i := 0; i < k; i++ {
		off := 288 * i
		t0, t1, t2 := off, off+256, off+288
		// Decode G1 point
		if points[i], err = g.DecodePoint(input[t0:t1]); err != nil {
			return nil, err
		}
		// Decode scalar value
		scalars[i] = new(big.Int).SetBytes(input[t1:t2])
	}

	// Compute r = e_0 * p_0 + e_1 * p_1 + ... + e_(k-1) * p_(k-1)
	r := g.New()
	g.MultiExp(r, points, scalars)

	// Encode the G2 point to 256 bytes.
	return g.EncodePoint(r), nil
}

// bls12381Pairing implements EIP-2537 Pairing precompile.
type bls12381Pairing struct{}

// RequiredGas returns the gas required to execute the pre-compiled contract.
func (c *bls12381Pairing) RequiredGas(input []byte) uint64 {
	return params.Bls12381PairingBaseGas + uint64(len(input)/384)*params.Bls12381PairingPerPairGas
}

func (c *bls12381Pairing) Run(input []byte) ([]byte, error) {
	// Implements EIP-2537 Pairing precompile logic.
	// > Pairing call expects `384*k` bytes as an inputs that is interpreted as byte concatenation of `k` slices. Each slice has the following structure:
	// > - `128` bytes of G1 point encoding
	// > - `256` bytes of G2 point encoding
	// > Output is a `32` bytes where last single byte is `0x01` if pairing result is equal to multiplicative identity in a pairing target field and `0x00` otherwise
	// > (which is equivalent of Big Endian encoding of Solidity values `uint256(1)` and `uin256(0)` respectively).
	k := len(input) / 384
	if len(input) == 0 || len(input)%384 != 0 {
		return nil, errBLS12381InvalidInputLength
	}

	// Initialize BLS12-381 pairing engine
	e := bls12381.NewPairingEngine()
	g1, g2 := e.G1, e.G2

	// Decode pairs
	for i := 0; i < k; i++ {
		off := 384 * i
		t0, t1, t2 := off, off+128, off+384

		// Decode G1 point
		p1, err := g1.DecodePoint(input[t0:t1])
		if err != nil {
			return nil, err
		}
		// Decode G2 point
		p2, err := g2.DecodePoint(input[t1:t2])
		if err != nil {
			return nil, err
		}

		// 'point is on curve' check already done,
		// Here we need to apply subgroup checks.
		if !g1.InCorrectSubgroup(p1) {
			return nil, errBLS12381G1PointSubgroup
		}
		if !g2.InCorrectSubgroup(p2) {
			return nil, errBLS12381G2PointSubgroup
		}

		// Update pairing engine with G1 and G2 ponits
		e.AddPair(p1, p2)
	}
	// Prepare 32 byte output
	out := make([]byte, 32)

	// Compute pairing and set the result
	if e.Check() {
		out[31] = 1
	}
	return out, nil
}

// decodeBLS12381FieldElement decodes BLS12-381 elliptic curve field element.
// Removes top 16 bytes of 64 byte input.
func decodeBLS12381FieldElement(in []byte) ([]byte, error) {
	if len(in) != 64 {
		return nil, errors.New("invalid field element length")
	}
	// check top bytes
	for i := 0; i < 16; i++ {
		if in[i] != byte(0x00) {
			return nil, errBLS12381InvalidFieldElementTopBytes
		}
	}
	out := make([]byte, 48)
	copy(out[:], in[16:])
	return out, nil
}

// bls12381MapG1 implements EIP-2537 MapG1 precompile.
type bls12381MapG1 struct{}

// RequiredGas returns the gas required to execute the pre-compiled contract.
func (c *bls12381MapG1) RequiredGas(input []byte) uint64 {
	return params.Bls12381MapG1Gas
}

func (c *bls12381MapG1) Run(input []byte) ([]byte, error) {
	// Implements EIP-2537 Map_To_G1 precompile.
	// > Field-to-curve call expects `64` bytes an an input that is interpreted as a an element of the base field.
	// > Output of this call is `128` bytes and is G1 point following respective encoding rules.
	if len(input) != 64 {
		return nil, errBLS12381InvalidInputLength
	}

	// Decode input field element
	fe, err := decodeBLS12381FieldElement(input)
	if err != nil {
		return nil, err
	}

	// Initialize G1
	g := bls12381.NewG1()

	// Compute mapping
	r, err := g.MapToCurve(fe)
	if err != nil {
		return nil, err
	}

	// Encode the G1 point to 128 bytes
	return g.EncodePoint(r), nil
}

// bls12381MapG2 implements EIP-2537 MapG2 precompile.
type bls12381MapG2 struct{}

// RequiredGas returns the gas required to execute the pre-compiled contract.
func (c *bls12381MapG2) RequiredGas(input []byte) uint64 {
	return params.Bls12381MapG2Gas
}

func (c *bls12381MapG2) Run(input []byte) ([]byte, error) {
	// Implements EIP-2537 Map_FP2_TO_G2 precompile logic.
	// > Field-to-curve call expects `128` bytes an an input that is interpreted as a an element of the quadratic extension field.
	// > Output of this call is `256` bytes and is G2 point following respective encoding rules.
	if len(input) != 128 {
		return nil, errBLS12381InvalidInputLength
	}

	// Decode input field element
	fe := make([]byte, 96)
	c0, err := decodeBLS12381FieldElement(input[:64])
	if err != nil {
		return nil, err
	}
	copy(fe[48:], c0)
	c1, err := decodeBLS12381FieldElement(input[64:])
	if err != nil {
		return nil, err
	}
	copy(fe[:48], c1)

	// Initialize G2
	g := bls12381.NewG2()

	// Compute mapping
	r, err := g.MapToCurve(fe)
	if err != nil {
		return nil, err
	}

	// Encode the G2 point to 256 bytes
	return g.EncodePoint(r), nil
}

//gyh group sign

const str = `type a
	q 8780710799663312522437781984754049815806883199414208211028653399266475630880222957078625179422662221423155858769582317459277713367317481324925129998224791
	h 12016012264891146079388821366740534204802954401251311822919615131047207289359704531102844802183906537786776
	r 730750818665451621361119245571504901405976559617
	exp2 159
	exp1 107
	sign1 1
	sign0 1`

var pairing, _ = pbc.NewPairingFromString(str)

type TIBGSMasterPublicKey struct {
	g, g2, h1, u0, u1, u2, u3, u4, n *pbc.Element //g，h1是G1上的点 g2,u0,u1,u2,u3,u4是G2上的点  n是GT上的点
}

type TIBGSMasterSecretKeyi struct {
	h2i *pbc.Element //G2上的一个点
}

type Sharealphar struct {
	alphai, ri *pbc.Element
}

type TIBGSGroupSecretKeyi struct {
	a0i, a2i, a3i, a4i, a5i *pbc.Element //a0i, a2i, a3i, a4i是G2上的点  a5i是G1上的点
}

type TIBGSGroupVerifyKeyi struct {
	gai *pbc.Element //G1上的点
}

type TIBGSUserSecretKey struct {
	b0, b3, b4, b5 *pbc.Element //b0, b3, b4是G2上的点 b5是G1上的点
}

type TIBGSOK struct {
	ok1, ok2 *pbc.Element //GT上的点
}

type TIBGSPOK struct {
	c, s1, s2, s3 *pbc.Element //Zr上的数
}

type TIBGSSIG struct {
	c0, c5, c6, e1, e2, e3 *pbc.Element //c5, e1是G1上的点  c0, c6, e2是G2上的点  e3是GT上的点
	pok                    TIBGSPOK
}

var gstatic = []byte{86, 128, 202, 65, 129, 43, 121, 169, 37, 55, 68, 20, 250, 195, 252, 225, 59, 245, 228, 150, 17, 53, 79, 113, 124, 91, 189, 249, 157, 250, 124, 10, 195, 133, 15, 124, 165, 45, 119, 11, 104, 244, 204, 204, 106, 189, 172, 17, 84, 198, 17, 53, 172, 21, 89, 170, 56, 164, 225, 154, 195, 127, 29, 73, 142, 57, 28, 134, 143, 128, 214, 189, 149, 72, 108, 147, 24, 194, 17, 242, 231, 21, 233, 253, 119, 7, 247, 91, 166, 162, 172, 232, 29, 242, 138, 187, 86, 227, 209, 146, 88, 36, 204, 155, 247, 34, 156, 167, 98, 171, 98, 208, 52, 25, 48, 160, 119, 57, 0, 3, 164, 67, 89, 130, 85, 246, 255, 244}
var g2static = []byte{60, 153, 28, 98, 67, 154, 113, 16, 117, 120, 181, 29, 234, 86, 207, 225, 171, 23, 134, 192, 22, 236, 107, 212, 207, 93, 43, 26, 192, 208, 27, 169, 114, 115, 10, 179, 118, 25, 66, 55, 169, 194, 205, 38, 172, 140, 141, 39, 71, 152, 26, 200, 97, 11, 17, 73, 106, 27, 24, 138, 199, 11, 13, 157, 24, 78, 211, 63, 217, 221, 169, 202, 207, 134, 184, 36, 121, 94, 75, 145, 204, 104, 72, 229, 68, 158, 10, 151, 94, 29, 159, 76, 140, 7, 35, 254, 148, 5, 251, 240, 18, 141, 113, 0, 39, 129, 152, 167, 129, 51, 140, 213, 204, 239, 79, 171, 165, 193, 71, 210, 80, 166, 165, 94, 228, 154, 1, 133}

// var h1static = []byte{146,57,123,198,171,171,126,151,125,8,13,215,238,187,71,141,249,7,64,116,83,56,91,173,170,130,165,218,172,39,209,219,250,166,42,140,38,207,89,212,48,211,13,245,124,58,172,235,91,156,173,68,96,113,115,180,235,196,254,17,138,16,113,113,66,72,203,192,12,28,214,217,129,114,38,255,84,212,1,187,85,13,188,67,4,239,136,123,40,20,201,41,113,11,157,90,29,205,123,217,166,51,172,186,37,13,150,206,138,175,40,107,61,224,116,43,146,232,81,120,254,147,7,169,115,166,164,198}
var u0static = []byte{33, 198, 191, 118, 110, 0, 171, 144, 76, 112, 162, 224, 84, 35, 127, 76, 146, 190, 225, 228, 173, 50, 207, 50, 120, 113, 235, 235, 192, 248, 154, 89, 196, 228, 107, 53, 205, 219, 52, 202, 35, 232, 53, 74, 89, 82, 17, 12, 108, 107, 147, 62, 1, 17, 201, 175, 199, 165, 119, 121, 169, 76, 212, 205, 84, 111, 199, 136, 65, 224, 130, 60, 220, 165, 77, 130, 73, 80, 191, 109, 220, 233, 7, 250, 16, 27, 228, 63, 123, 153, 66, 119, 43, 58, 103, 38, 157, 4, 90, 134, 49, 179, 239, 169, 78, 203, 153, 167, 120, 194, 212, 208, 236, 159, 160, 57, 69, 1, 247, 87, 204, 40, 3, 191, 195, 98, 197, 127}
var u1static = []byte{134, 101, 209, 13, 93, 59, 143, 183, 33, 192, 71, 144, 93, 157, 215, 69, 197, 98, 22, 43, 32, 228, 37, 169, 69, 117, 239, 123, 38, 116, 239, 255, 232, 129, 96, 84, 109, 131, 221, 34, 135, 186, 250, 68, 199, 225, 56, 181, 32, 238, 251, 89, 64, 58, 66, 184, 152, 163, 240, 35, 185, 78, 29, 255, 32, 90, 21, 17, 195, 116, 184, 70, 26, 196, 160, 218, 50, 119, 19, 110, 156, 104, 214, 222, 37, 31, 238, 43, 124, 162, 210, 176, 189, 53, 236, 222, 163, 189, 3, 37, 85, 18, 42, 26, 134, 118, 85, 236, 155, 55, 154, 139, 242, 159, 156, 245, 46, 68, 210, 6, 35, 204, 88, 128, 127, 97, 217, 26}
var u2static = []byte{132, 123, 240, 68, 4, 50, 183, 140, 79, 162, 119, 200, 178, 44, 170, 12, 214, 31, 15, 207, 162, 144, 159, 147, 111, 192, 215, 127, 42, 1, 218, 125, 56, 201, 81, 114, 243, 53, 109, 70, 36, 161, 56, 160, 203, 21, 135, 252, 245, 108, 244, 95, 94, 79, 76, 106, 100, 149, 1, 237, 58, 71, 253, 70, 164, 190, 33, 1, 243, 227, 229, 198, 55, 19, 81, 66, 246, 66, 53, 40, 192, 199, 192, 4, 29, 219, 149, 90, 5, 19, 239, 165, 109, 166, 171, 88, 37, 220, 9, 160, 236, 149, 35, 118, 80, 220, 153, 113, 207, 155, 62, 134, 148, 34, 202, 143, 79, 208, 23, 173, 4, 142, 92, 187, 175, 105, 132, 253}
var u3static = []byte{92, 248, 171, 53, 226, 107, 27, 134, 195, 216, 221, 30, 133, 227, 150, 106, 38, 155, 71, 183, 235, 65, 132, 59, 180, 39, 181, 35, 159, 197, 135, 197, 228, 162, 31, 49, 169, 135, 121, 51, 234, 187, 99, 197, 88, 111, 222, 242, 120, 152, 56, 56, 230, 125, 126, 15, 102, 152, 29, 214, 65, 5, 138, 105, 67, 217, 43, 156, 241, 192, 12, 245, 103, 190, 53, 29, 103, 11, 92, 167, 69, 137, 122, 186, 110, 217, 190, 2, 235, 167, 31, 117, 37, 131, 222, 104, 236, 102, 170, 190, 187, 226, 107, 63, 109, 45, 115, 202, 2, 140, 19, 173, 8, 96, 42, 46, 46, 68, 133, 69, 171, 164, 181, 224, 9, 67, 54, 173}
var u4static = []byte{37, 130, 102, 158, 66, 112, 163, 71, 3, 23, 88, 31, 182, 102, 95, 23, 156, 134, 56, 37, 25, 13, 131, 215, 149, 156, 90, 101, 63, 221, 149, 171, 92, 189, 132, 122, 193, 229, 186, 179, 4, 180, 202, 90, 229, 43, 223, 179, 138, 138, 238, 222, 162, 24, 231, 172, 185, 3, 93, 11, 96, 207, 8, 209, 83, 57, 29, 157, 19, 111, 215, 133, 130, 195, 212, 144, 101, 215, 24, 115, 131, 208, 109, 93, 37, 194, 198, 111, 11, 207, 14, 2, 158, 162, 88, 19, 212, 175, 20, 118, 243, 25, 235, 177, 151, 213, 147, 242, 32, 70, 202, 163, 112, 237, 175, 219, 77, 235, 147, 220, 38, 39, 121, 113, 219, 2, 203, 26}
var nstatic = []byte{93, 122, 146, 183, 222, 144, 80, 111, 128, 201, 250, 100, 43, 213, 102, 147, 47, 168, 91, 130, 155, 17, 166, 2, 189, 212, 173, 9, 121, 59, 157, 110, 10, 44, 123, 170, 130, 53, 80, 159, 221, 208, 65, 185, 167, 197, 154, 135, 40, 99, 194, 165, 163, 207, 157, 124, 11, 37, 232, 186, 75, 135, 111, 194, 153, 80, 154, 234, 8, 167, 95, 162, 109, 225, 40, 97, 30, 113, 253, 88, 84, 5, 44, 107, 184, 130, 145, 10, 111, 246, 147, 115, 230, 167, 0, 99, 102, 157, 58, 50, 183, 98, 51, 202, 101, 219, 66, 17, 204, 98, 80, 225, 177, 146, 118, 198, 160, 98, 135, 232, 167, 249, 145, 162, 41, 108, 224, 93}

func convert(array interface{}) string {
	return strings.Replace(strings.Trim(fmt.Sprint(array), "[]"), " ", ",", -1)
}

func G(ID string) *pbc.Element {
	I := pairing.NewZr().SetFromStringHash(ID, sha256.New())
	return I
}

func CN(Cs []*pbc.Element, N int) []*pbc.Element {
	NN := make([]*pbc.Element, N)
	// fmt.Println("NmanagerEXLEL", NN)

	for i := 0; i < N; i++ {

		NN[i] = CNi(Cs, N, i)
	}
	// fmt.Println("Nmanager", NN)
	return NN
}

func CNi(Cs []*pbc.Element, N, i int) *pbc.Element {
	sum := pairing.NewZr().Set0()
	pre := int32(1)
	for _, coef := range Cs {
		temp1 := pairing.NewZr().MulInt32(coef, pre)
		sum.Add(sum, temp1)
		pre = pre * int32((i + 1))
	}
	// fmt.Println("sum",sum, i)
	return sum
}

// GenCoef generates the coefficients used by func C
func GenCoef(szero *pbc.Element, t int) []*pbc.Element {
	Cs := make([]*pbc.Element, t)
	Cs[0] = szero
	for i := 1; i < t; i++ { //生成K-1个系数
		Cs[i] = pairing.NewZr().Rand()
	}
	// fmt.Println("Cs",Cs)
	return Cs
}
func SharesGen(thresholdOfLevel0 int, numOfLevel0 int, s *pbc.Element) ([]*pbc.Element, []*pbc.Element) {
	//t1:=time.Now()
	//每个manager生成 s,t
	// s := pairing.NewZr().Rand()
	// t := pairing.NewZr().Rand()
	//产生多项式系数  k-1次多项式 t个系数，第一个是要分享的
	sk := GenCoef(s, thresholdOfLevel0)
	// tk := GenCoef(t, thresholdOfLevel0)
	//产生N个管理员的分享

	SCN := CN(sk, numOfLevel0)
	// TCN := CN(tk, numOfLevel0)
	//t2:=time.Now()
	//f, _ := os.OpenFile("SharesGen.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	//fmt.Fprintln(f,t2.Sub(t1))
	return SCN, nil
}

//通过n个分享生成 alphai 和 ri
func ReconManager(ss []*pbc.Element, tt []*pbc.Element, N int) (*pbc.Element, *pbc.Element) {
	//t1:=time.Now()
	if len(ss) == N && len(tt) == N {
		alphai := pairing.NewZr().Set0()
		ri := pairing.NewZr().Set0()
		for i := 0; i < N; i++ {
			alphai.Add(alphai, ss[i])
			ri.Add(ri, tt[i])
		}
		//t2:=time.Now()
		//fmt.Println("time:ReconManager",t2.Sub(t1))
		//f, _ := os.OpenFile("ReconManager.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
		//fmt.Fprintln(f,t2.Sub(t1))
		return alphai, ri
	}
	return nil, nil
}

// L generates the Lagrange coefficient of an index   L(i)  i=indexOfArray+1<=k
func L(K, indexOfArray int) *pbc.Element { // indexOfArray represents index of SelectedNodes
	if indexOfArray > K-1 {
		return nil
	}
	L := pairing.NewZr().Set1()
	I := pairing.NewZr().SetInt32(int32(indexOfArray + 1)) //I <= K
	for j := 0; j < K; j++ {
		J := pairing.NewZr().SetInt32(int32(j + 1))
		if J.Equals(I) {
			continue
		} else {
			temp1 := pairing.NewZr().Sub(J, I)
			temp2 := pairing.NewZr().Div(J, temp1)
			L.Mul(L, temp2)
			// fmt.Println("L", L)
		}
	}
	return L
}

//k个g^alphai 还原个g^alpha
func Galpha(SelectedNodes []*pbc.Element, K int) *pbc.Element {
	if len(SelectedNodes) < K {
		fmt.Println("galphai not enough")
		return nil
	}
	h1 := pairing.NewG1().Set1()
	for i := 0; i < K; i++ {
		temp := pairing.NewG1().PowZn(SelectedNodes[i], L(K, i))
		h1.Mul(h1, temp)
	}
	return h1
}

func Setup(numOfLevel0, thresholdOfLevel0 int, alphai *pbc.Element, kgalphai []*pbc.Element) (*TIBGSMasterPublicKey, *TIBGSMasterSecretKeyi) {
	//t1:=time.Now()
	var mpk TIBGSMasterPublicKey //g，h1是G1上的点 g2,u0,u1,u2,u3,u4是G2上的点  n是GT上的点
	mpk.g = pairing.NewG1().SetBytes(gstatic)
	mpk.g2 = pairing.NewG2().SetBytes(g2static)
	// mpk.h1 = pairing.NewG1().SetXBytes(h1static)
	mpk.u0 = pairing.NewG2().SetBytes(u0static)
	mpk.u1 = pairing.NewG2().SetBytes(u1static)
	mpk.u2 = pairing.NewG2().SetBytes(u2static)
	mpk.u3 = pairing.NewG2().SetBytes(u3static)
	mpk.u4 = pairing.NewG2().SetBytes(u4static)
	mpk.n = pairing.NewGT().SetBytes(nstatic)
	//拉格朗日插值计算h1
	mpk.h1 = Galpha(kgalphai, thresholdOfLevel0) //通过k个g^alphai求出g^alpha
	//门限 计算mski
	var mski TIBGSMasterSecretKeyi //G2上的一个点
	//通过门限求出alphai n
	mski.h2i = pairing.NewG2().PowZn(mpk.g2, alphai) //g2^alphai
	//t2:=time.Now()
	// fmt.Println("time:Setup",t2.Sub(t1))
	//f, _ := os.OpenFile("Setup.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	//fmt.Fprintln(f,t2.Sub(t1))
	return &mpk, &mski
}

//实现方便，单节点生成
func NewSetup(numOfLevel0, thresholdOfLevel0 int, grpID string) (*TIBGSMasterPublicKey, *TIBGSMasterSecretKeyi, *TIBGSGroupSecretKeyi, *TIBGSGroupVerifyKeyi, []*Sharealphar, error) {
	var mpk TIBGSMasterPublicKey //g，h1是G1上的点 g2,u0,u1,u2,u3,u4是G2上的点  n是GT上的点
	mpk.g = pairing.NewG1().SetBytes(gstatic)
	mpk.g2 = pairing.NewG2().SetBytes(g2static)
	// mpk.h1 = pairing.NewG1().SetXBytes(h1static)
	mpk.u0 = pairing.NewG2().SetBytes(u0static)
	mpk.u1 = pairing.NewG2().SetBytes(u1static)
	mpk.u2 = pairing.NewG2().SetBytes(u2static)
	mpk.u3 = pairing.NewG2().SetBytes(u3static)
	mpk.u4 = pairing.NewG2().SetBytes(u4static)
	mpk.n = pairing.NewGT().SetBytes(nstatic)
	//随机生成alpha和r
	alpha := pairing.NewZr().Rand()
	r := pairing.NewZr().Rand()

	mpk.h1 = pairing.NewG1().PowZn(mpk.g, alpha)

	//生成n个节点的alphai和ri
	sk := GenCoef(alpha, thresholdOfLevel0)
	tk := GenCoef(r, thresholdOfLevel0)
	//产生N个管理员的分享
	nalpha := CN(sk, numOfLevel0)
	nr := CN(tk, numOfLevel0)

	nss := make([]*Sharealphar, numOfLevel0)
	for i := 0; i < numOfLevel0; i++ {
		var tmpSs Sharealphar
		tmpSs.alphai = nalpha[i] //g2^s[i]
		tmpSs.ri = nr[i]         //1
		nss[i] = &tmpSs
	}

	falphai := nalpha[0]
	fri := nr[0]
	//门限 计算mski
	var mski TIBGSMasterSecretKeyi //G2上的一个点
	//通过门限求出alphai n
	mski.h2i = pairing.NewG2().PowZn(mpk.g2, falphai) //g2^alphai

	I := G(grpID)
	// fmt.Println("G(grpID)",G(grpID))
	var gski TIBGSGroupSecretKeyi //a0i, a2i, a3i, a4i是G2上的点  a5i是G1上的点
	//秘密分享ri n
	a0 := pairing.NewG2().PowZn(mpk.u1, I)       //u1^G(grpID)
	a0.Mul(mpk.u0, a0)                           //u0 * u1^G(grpID)
	a0.PowZn(a0, fri)                            //(u0 * u1^G(grpID))^ri
	gski.a0i = pairing.NewG2().Mul(mski.h2i, a0) // h2i*(u0 * u1^G(grpID))^
	// fmt.Println("gski.a0i",gski.a0i)
	gski.a2i = pairing.NewG2().PowZn(mpk.u2, fri)
	gski.a3i = pairing.NewG2().PowZn(mpk.u3, fri)
	gski.a4i = pairing.NewG2().PowZn(mpk.u4, fri)
	gski.a5i = pairing.NewG1().PowZn(mpk.g, fri)

	var gvki TIBGSGroupVerifyKeyi //G1上的点
	gvki.gai = pairing.NewG1().PowZn(mpk.g, falphai)
	return &mpk, &mski, &gski, &gvki, nss, nil
}

//为其他n-1节点生成mski,gski,gvki
func Gen3key(mpk *TIBGSMasterPublicKey, ar *Sharealphar, grpID string) (*TIBGSMasterSecretKeyi, *TIBGSGroupSecretKeyi, *TIBGSGroupVerifyKeyi) {
	falphai := pairing.NewZr().Set(ar.alphai)
	fri := pairing.NewZr().Set(ar.ri)

	//门限 计算mski
	var mski TIBGSMasterSecretKeyi //G2上的一个点
	//通过门限求出alphai n
	mski.h2i = pairing.NewG2().PowZn(mpk.g2, falphai) //g2^alphai

	I := G(grpID)
	// fmt.Println("G(grpID)",G(grpID))
	var gski TIBGSGroupSecretKeyi //a0i, a2i, a3i, a4i是G2上的点  a5i是G1上的点
	//秘密分享ri n
	a0 := pairing.NewG2().PowZn(mpk.u1, I)       //u1^G(grpID)
	a0.Mul(mpk.u0, a0)                           //u0 * u1^G(grpID)
	a0.PowZn(a0, fri)                            //(u0 * u1^G(grpID))^ri
	gski.a0i = pairing.NewG2().Mul(mski.h2i, a0) // h2i*(u0 * u1^G(grpID))^
	// fmt.Println("gski.a0i",gski.a0i)
	gski.a2i = pairing.NewG2().PowZn(mpk.u2, fri)
	gski.a3i = pairing.NewG2().PowZn(mpk.u3, fri)
	gski.a4i = pairing.NewG2().PowZn(mpk.u4, fri)
	gski.a5i = pairing.NewG1().PowZn(mpk.g, fri)

	var gvki TIBGSGroupVerifyKeyi //G1上的点
	gvki.gai = pairing.NewG1().PowZn(mpk.g, falphai)
	return &mski, &gski, &gvki
}

func GrpSetUp(mpk *TIBGSMasterPublicKey, mski *TIBGSMasterSecretKeyi, grpID string, ri, alphai *pbc.Element) (*TIBGSGroupSecretKeyi, *TIBGSGroupVerifyKeyi) {
	//t1:=time.Now()
	I := G(grpID)
	// fmt.Println("G(grpID)",G(grpID))
	var gski TIBGSGroupSecretKeyi //a0i, a2i, a3i, a4i是G2上的点  a5i是G1上的点
	//秘密分享ri n
	a0 := pairing.NewG2().PowZn(mpk.u1, I)       //u1^G(grpID)
	a0.Mul(mpk.u0, a0)                           //u0 * u1^G(grpID)
	a0.PowZn(a0, ri)                             //(u0 * u1^G(grpID))^ri
	gski.a0i = pairing.NewG2().Mul(mski.h2i, a0) // h2i*(u0 * u1^G(grpID))^
	// fmt.Println("gski.a0i",gski.a0i)
	gski.a2i = pairing.NewG2().PowZn(mpk.u2, ri)
	gski.a3i = pairing.NewG2().PowZn(mpk.u3, ri)
	gski.a4i = pairing.NewG2().PowZn(mpk.u4, ri)
	gski.a5i = pairing.NewG1().PowZn(mpk.g, ri)

	var gvki TIBGSGroupVerifyKeyi //G1上的点
	gvki.gai = pairing.NewG1().PowZn(mpk.g, alphai)
	//t2:=time.Now()
	// fmt.Println("time:GrpSetUp",t2.Sub(t1))
	//f, _ := os.OpenFile("GrpSetUp.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	//fmt.Fprintln(f,t2.Sub(t1))
	return &gski, &gvki
}

func ExtShare(gski *TIBGSGroupSecretKeyi, userID string) *TIBGSUserSecretKey {
	//t1:=time.Now()
	I := G(userID)
	var uski TIBGSUserSecretKey                  //b0, b3, b4是G2上的点 b5是G1上的点
	mid := pairing.NewG2().PowZn(gski.a2i, I)    //a2i^G(userID)
	uski.b0 = pairing.NewG2().Mul(gski.a0i, mid) //a0i * a2i^G(userID)
	uski.b3 = gski.a3i
	uski.b4 = gski.a4i
	uski.b5 = gski.a5i
	//t2:=time.Now()
	// fmt.Println("time:ExtShare",t2.Sub(t1))
	//f, _ := os.OpenFile("ExtShare.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	//fmt.Fprintln(f,t2.Sub(t1))
	return &uski
}

func VerifyShare(uski *TIBGSUserSecretKey, gvki *TIBGSGroupVerifyKeyi, mpk *TIBGSMasterPublicKey, grpID string, userID string) bool {
	//t1:=time.Now()
	IG := G(grpID)
	IU := G(userID)

	left1 := pairing.NewGT().Pair(mpk.g, uski.b0)      //e(g , b0)
	right11 := pairing.NewGT().Pair(gvki.gai, mpk.g2)  //e(gvk , g2)
	right12 := pairing.NewGT().Pair(uski.b5, mpk.u0)   //e(b5, u0)
	u1grpID := pairing.NewG2().PowZn(mpk.u1, IG)       //u1^G(grpID)
	right13 := pairing.NewGT().Pair(uski.b5, u1grpID)  //e(b5,u1^G(grpID))
	u2userID := pairing.NewG2().PowZn(mpk.u2, IU)      //u2^G(userID)
	right14 := pairing.NewGT().Pair(uski.b5, u2userID) //e(b5, u2^G(userID))
	right1 := pairing.NewGT().Mul(right11, right12)    // e(gvk , g2) e(b5, u0)
	right1.Mul(right1, right13)                        //e(gvk , g2) e(b5, u0) e(b5,u1^G(grpID))
	right1.Mul(right1, right14)                        //e(gvk , g2) e(b5, u0) e(b5,u1^G(grpID)) e(b5, u2^G(userID))
	left2 := pairing.NewGT().Pair(mpk.g, uski.b3)      //e(g, b3)
	right2 := pairing.NewGT().Pair(uski.b5, mpk.u3)    //e(b5, u3)
	left3 := pairing.NewGT().Pair(mpk.g, uski.b4)      //e(g, b4)
	right3 := pairing.NewGT().Pair(uski.b5, mpk.u4)    //e(b5, u4)
	tf := left1.Equals(right1) && left2.Equals(right2) && left3.Equals(right3)
	//t2:=time.Now()
	// fmt.Println("time:VerifyShare",t2.Sub(t1))
	//f, _ := os.OpenFile("VerifyShare.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	//fmt.Fprintln(f,t2.Sub(t1))
	return tf
}

func ReconstKey(uskis []*TIBGSUserSecretKey, K int, mpk *TIBGSMasterPublicKey, grpID string, userID string) *TIBGSUserSecretKey {
	//t1:=time.Now()
	IG := G(grpID)
	IU := G(userID)
	var usk TIBGSUserSecretKey //b0, b3, b4是G2上的点 b5是G1上的点
	// KL := make([]*pbc.Element, K)
	var b01, b31, b41, b51 *pbc.Element
	mb0, mb3, mb4, mb5 := pairing.NewG2().Set1(), pairing.NewG2().Set1(), pairing.NewG2().Set1(), pairing.NewG1().Set1()
	for i := 0; i < K; i++ {
		KL := L(K, i) //i从0开始  在L里会+1
		b01 = pairing.NewG2().PowZn(uskis[i].b0, KL)
		mb0.Mul(mb0, b01)
		b31 = pairing.NewG2().PowZn(uskis[i].b3, KL)
		mb3.Mul(mb3, b31)
		b41 = pairing.NewG2().PowZn(uskis[i].b4, KL)
		mb4.Mul(mb4, b41)
		b51 = pairing.NewG1().PowZn(uskis[i].b5, KL)
		mb5.Mul(mb5, b51)
	}
	// fmt.Println("mb0,mb3,mb4,mb5",mb0,mb3,mb4,mb5)
	b02 := pairing.NewG2().PowZn(mpk.u1, IG) //u1^G(grpID)
	b02.Mul(mpk.u0, b02)                     //u0 * u1^G(grpID)
	b03 := pairing.NewG2().PowZn(mpk.u2, IU) //u2^G(UserID)
	b02.Mul(b02, b03)                        //u0 * u1^G(grpID)*u2^G(UserID)
	r2 := pairing.NewZr().Rand()             //随机
	b02.PowZn(b02, r2)                       //(u0 * u1^G(grpID)*u2^G(UserID))^r2
	usk.b0 = pairing.NewG2().Mul(mb0, b02)
	usk.b3 = pairing.NewG2().PowZn(mpk.u3, r2) //u3^r2
	usk.b3.Mul(mb3, usk.b3)
	usk.b4 = pairing.NewG2().PowZn(mpk.u4, r2)
	usk.b4.Mul(mb4, usk.b4)
	usk.b5 = pairing.NewG1().PowZn(mpk.g, r2)
	usk.b5.Mul(mb5, usk.b5)
	//t2:=time.Now()
	// fmt.Println("time:ReconstKey",t2.Sub(t1))
	//f, _ := os.OpenFile("ReconstKey.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	//fmt.Fprintln(f,t2.Sub(t1))
	return &usk
}

func Sign(mpk *TIBGSMasterPublicKey, usk *TIBGSUserSecretKey, message string, grpID string, userID string) *TIBGSSIG {
	//t1:=time.Now()
	msg := G(message)
	IG := G(grpID)
	IU := G(userID)
	var ssig TIBGSSIG //c5, e1是G1上的点  c0, c6, e2是G2上的点  e3是GT上的点
	var POK TIBGSPOK  //c, s1, s2, s3 Zr上的数
	r3, f, rID := pairing.NewZr().Rand(), pairing.NewZr().Rand(), pairing.NewZr().Rand()
	c02 := pairing.NewG2().PowZn(usk.b3, msg)   //b3^m
	c03 := pairing.NewG2().PowZn(usk.b4, rID)   //b4^G(rID)
	c04 := pairing.NewG2().PowZn(mpk.u1, IG)    //u1^G(grpID)
	c04.Mul(mpk.u0, c04)                        //u0 * u1^G(grpID)
	c043 := pairing.NewG2().PowZn(mpk.u2, IU)   //u2^G(userID)
	c04.Mul(c04, c043)                          //u0 * u1^G(grpID)*u2^G(userID)
	c044 := pairing.NewG2().PowZn(mpk.u3, msg)  //u3^m
	c04.Mul(c04, c044)                          //u0 * u1^G(grpID)*u2^G(userID) * u3^m
	c045 := pairing.NewG2().PowZn(mpk.u4, rID)  // u4^G(rID)
	c04.Mul(c04, c045)                          //u0 * u1^G(grpID)*u2^G(userID) * u3^m * u4^G(rID)
	c04.PowZn(c04, r3)                          //(u0 * u1^G(grpID) * u2^G(userID) * u3^m * u4^G(rID))^r3
	ssig.c0 = pairing.NewG2().Mul(usk.b0, c02)  // b0 * b3^m
	ssig.c0.Mul(ssig.c0, c03)                   //b0 * b3^m * b4^G(rID)
	ssig.c0.Mul(ssig.c0, c04)                   // b0 * b3^m * b4^G(rID) * (u0 * u1^G(grpID) * u2^G(userID) * u3^m * u4^G(rID))^r3
	ssig.c5 = pairing.NewG1().PowZn(mpk.g, r3)  //g^r3
	ssig.c5.Mul(usk.b5, ssig.c5)                //b5 * g^r3
	c61 := pairing.NewG2().PowZn(mpk.u2, IU)    //u2^G(userID)
	c62 := pairing.NewG2().PowZn(mpk.u4, rID)   //u4^G(rID)
	ssig.c6 = pairing.NewG2().Mul(c61, c62)     // u2^G(userID) * u4^G(rID)
	ssig.e1 = pairing.NewG1().PowZn(mpk.g, f)   //g^f
	ssig.e2 = pairing.NewG2().PowZn(mpk.u1, IG) //u1^G(grpID)
	ssig.e2.Mul(mpk.u0, ssig.e2)                //u0 * u1^G(grpID)
	ssig.e2.PowZn(ssig.e2, f)                   //(u0 * u1^G(grpID))^f
	e31 := pairing.NewGT().PowZn(mpk.n, IU)     //n^G(userID)
	e32 := pairing.NewGT().Pair(mpk.h1, mpk.g2) //e(h1, g2)
	e32.PowZn(e32, f)                           // e(h1, g2)^f
	ssig.e3 = pairing.NewGT().Mul(e31, e32)     //n^G(userID) * e(h1, g2)^f
	//生成POK
	k1, k2, k3 := pairing.NewZr().Rand(), pairing.NewZr().Rand(), pairing.NewZr().Rand()
	hatf := pairing.NewG2().PowZn(mpk.u1, IG)    //u1^G(grpID)
	hatf.Mul(mpk.u0, hatf)                       //u0*u1^G(grpID)
	bigg := pairing.NewGT().Pair(mpk.h1, mpk.g2) //e(h1,g2)
	pr11 := pairing.NewG2().PowZn(mpk.u2, k1)    //u2^k1
	pr12 := pairing.NewG2().PowZn(mpk.u4, k2)    //u4^k2
	pr1 := pairing.NewG2().Mul(pr11, pr12)       //u2^k1 * u4^k2
	pr2 := pairing.NewG1().PowZn(mpk.g, k3)      //g^k3
	pr3 := pairing.NewG2().PowZn(hatf, k3)       //hatf^k3
	pt41 := pairing.NewGT().PowZn(mpk.n, k1)     //n^k1
	pt42 := pairing.NewGT().PowZn(bigg, k3)      //bigg^k3
	pr4 := pairing.NewGT().Mul(pt41, pt42)       //n^k1* bigg^k3

	POK.c = pairing.NewZr().SetFromStringHash(pr1.String()+pr2.String()+pr3.String()+pr4.String(), sha256.New())
	cx := pairing.NewZr().Mul(POK.c, IU)  //c*IU
	POK.s1 = pairing.NewZr().Add(k1, cx)  //k1 + c*IU
	cy := pairing.NewZr().Mul(POK.c, rID) //c*rID
	POK.s2 = pairing.NewZr().Add(k2, cy)  //k2+c*rID
	cz := pairing.NewZr().Mul(POK.c, f)
	POK.s3 = pairing.NewZr().Add(k3, cz)
	ssig.pok = POK
	//t2:=time.Now()
	// fmt.Println("time:Sign",t2.Sub(t1))
	//file, _ := os.OpenFile("Sign.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	//fmt.Fprintln(file,t2.Sub(t1))
	return &ssig
}

func Verify(ssig *TIBGSSIG, mpk *TIBGSMasterPublicKey, message string, grpID string) bool {
	//t1:=time.Now()
	msg := G(message)
	IG := G(grpID)

	t := pairing.NewZr().Rand()
	M := pairing.NewGT().Rand()
	// fmt.Println("M",M)
	d1 := pairing.NewG1().PowZn(mpk.g, t)        //g^t
	d22 := pairing.NewG2().PowZn(mpk.u1, IG)     // u1^G(grpID)
	d23 := pairing.NewG2().PowZn(mpk.u3, msg)    //u3^m
	d2 := pairing.NewG2().Mul(mpk.u0, d22)       //u0 * u1^G(grpID)
	d2.Mul(d2, d23)                              //u0 * u1^G(grpID) * u3^m
	d2.Mul(d2, ssig.c6)                          //u0 * u1^G(grpID) * u3^m *c6
	d2.PowZn(d2, t)                              //(u0 * u1^G(grpID) * u3^m *c6)^t
	zeta := pairing.NewGT().Pair(mpk.h1, mpk.g2) //e(h1, g2)
	zeta.PowZn(zeta, t)                          //e(h1, g2)^t
	zeta.Mul(M, zeta)                            //M * e(h1, g2)^t
	temp1 := pairing.NewGT().Pair(ssig.c5, d2)   //e(c5, d2)
	temp2 := pairing.NewGT().Pair(d1, ssig.c0)   //e(d1, c0)
	right := pairing.NewGT().Div(temp1, temp2)   // e(c5, d2) / e(d1, c0)
	right.Mul(zeta, right)                       // zeta * (e(c5, d2) / e(d1, c0))
	//验证POK
	rr11 := pairing.NewG2().PowZn(mpk.u2, ssig.pok.s1) //u2^s1
	rr12 := pairing.NewG2().PowZn(mpk.u4, ssig.pok.s2) //u4^s2
	negc := pairing.NewZr().Neg(ssig.pok.c)            //-c
	rr13 := pairing.NewG2().PowZn(ssig.c6, negc)       //c6^-c
	rr1 := pairing.NewG2().Mul(rr11, rr12)             //u2^s1 * u4^s2
	rr1.Mul(rr1, rr13)                                 //u2^s1 * u4^s2* c6^-c
	rr21 := pairing.NewG2().PowZn(mpk.g, ssig.pok.s3)  //g^s3
	rr22 := pairing.NewG2().PowZn(ssig.e1, negc)       //e1^-c
	rr2 := pairing.NewG2().Mul(rr21, rr22)             //g^s3 * e1^-c
	hatf := pairing.NewG2().PowZn(mpk.u1, IG)          //u1^G(grpID)
	hatf.Mul(mpk.u0, hatf)                             //u0*u1^G(grpID)
	bigg := pairing.NewGT().Pair(mpk.h1, mpk.g2)
	rr31 := pairing.NewG2().PowZn(hatf, ssig.pok.s3)  //hatf^s3
	rr32 := pairing.NewG2().PowZn(ssig.e2, negc)      //e2^-c
	rr3 := pairing.NewG2().Mul(rr31, rr32)            //hatf^s3 * e2^-c
	tt41 := pairing.NewGT().PowZn(mpk.n, ssig.pok.s1) //n^s1
	tt42 := pairing.NewGT().PowZn(bigg, ssig.pok.s3)  //bigg^s3
	tt43 := pairing.NewGT().PowZn(ssig.e3, negc)      //e3^-c
	tt4 := pairing.NewGT().Mul(tt41, tt42)            //n^s1* bigg^s3
	tt4.Mul(tt4, tt43)                                //n^s1* bigg^s3 *e3^-c
	cc := pairing.NewZr().SetFromStringHash(rr1.String()+rr2.String()+rr3.String()+tt4.String(), sha256.New())
	// fmt.Println("cc",cc)
	// fmt.Println("pok.c",ssig.pok.c)
	tf := M.Equals(right) && cc.Equals(ssig.pok.c)
	//t2:=time.Now()
	// fmt.Println("time:Verify",t2.Sub(t1))
	//f, _ := os.OpenFile("Verify.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	//fmt.Fprintln(f,t2.Sub(t1))
	return tf
}

func OpenPart(gski *TIBGSGroupSecretKeyi, ssig *TIBGSSIG) *TIBGSOK {
	//t1:=time.Now()
	var OKi TIBGSOK //ok1, ok2 GT上的点
	OKi.ok1 = pairing.NewGT().Pair(ssig.e1, gski.a0i)
	OKi.ok2 = pairing.NewGT().Pair(gski.a5i, ssig.e2)
	//t2:=time.Now()
	// fmt.Println("time:OpenPart",t2.Sub(t1))
	//f, _ := os.OpenFile("OpenPart.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	//fmt.Fprintln(f,t2.Sub(t1))
	return &OKi
}

func Open(OKK []*TIBGSOK, K int) *pbc.Element {
	//t1:=time.Now()
	//lag
	if len(OKK) < K {
		fmt.Println("ok is not enough")
		return nil
	}
	temp1, temp2 := pairing.NewGT().Set1(), pairing.NewGT().Set1()
	var m1, m2 *pbc.Element
	for i := 0; i < K; i++ {
		KL := L(K, i) //i从0开始  在L里会+1
		m1 = pairing.NewGT().PowZn(OKK[i].ok1, KL)
		temp1.Mul(temp1, m1)
		m2 = pairing.NewGT().PowZn(OKK[i].ok2, KL)
		temp2.Mul(temp2, m2)
	}
	gama := pairing.NewGT().Div(temp1, temp2)
	//t2:=time.Now()
	// fmt.Println("time:Open",t2.Sub(t1))
	//f, _ := os.OpenFile("Open.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	//fmt.Fprintln(f,t2.Sub(t1))
	return gama
}

func FindUser(UIDS []string, gama *pbc.Element, ssig *TIBGSSIG, mpk *TIBGSMasterPublicKey) string {
	//t1:=time.Now()
	for _, ID := range UIDS {
		GU := G(ID)
		right := pairing.NewGT().PowZn(mpk.n, GU)
		right.Mul(right, gama)

		if ssig.e3.Equals(right) {
			//t2:=time.Now()
			// fmt.Println("time:FindUser",t2.Sub(t1))
			//f, _ := os.OpenFile("FindUser.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
			//fmt.Fprintln(f,t2.Sub(t1))
			return ID
		}

	}
	return "no user here"
}

func BytesCombine1(pBytes ...[]byte) []byte {
	length := len(pBytes)
	s := make([][]byte, length)
	for index := 0; index < length; index++ {
		s[index] = pBytes[index]
	}
	sep := []byte("")
	return bytes.Join(s, sep)
}

//TIBGS
type veriGroupsign struct{}

func (c *veriGroupsign) RequiredGas(input []byte) uint64 {
	return params.VeriGroupsign
}

func (c *veriGroupsign) Run(input []byte) ([]byte, error) {
	//func Verify(ssig *TIBGSSIG, mpk *TIBGSMasterPublicKey, message string, grpID string) bool{
	// TODO: 输入从byte转化(h1,ssig与message)
	fmt.Println("run veriGroupsign test ....")
	//fmt.Println("input:", input)
	//fmt.Println("input_string:", string(input))
	//fmt.Println("input_1", input[1])
	input_string := string(input)
	//input_string = strings.Replace(input_string, "[", "", -1)
	//input_string = strings.Replace(input_string, "]", "", -1)
	fmt.Println("input_string:", input_string)
	strarry := strings.Split(input_string, "#")
	// fmt.Println(tkvkstr)
	//rkeysstr := []byte(strarry[0])
	//fmt.Println("rkeysstr_byte=", rkeysstr)

	//pkstr :=[]byte(strarry[1])
	//fmt.Println("rkeysstr_byte=", strarry[0])
	g_get,_ := hex.DecodeString(strarry[0])
	g2_get,_ := hex.DecodeString(strarry[1])
	u0_get,_ := hex.DecodeString(strarry[2])
	u1_get,_ := hex.DecodeString(strarry[3])
	u2_get,_ := hex.DecodeString(strarry[4])
	u3_get,_ := hex.DecodeString(strarry[5])
	u4_get,_ := hex.DecodeString(strarry[6])
	n_get,_ := hex.DecodeString(strarry[7])
	h1_get,_ := hex.DecodeString(strarry[8])
	c0_get,_ := hex.DecodeString(strarry[9])
	c5_get,_ := hex.DecodeString(strarry[10])
	c6_get,_ := hex.DecodeString(strarry[11])
	e1_get,_ := hex.DecodeString(strarry[12])
	e2_get,_ := hex.DecodeString(strarry[13])
	e3_get,_ := hex.DecodeString(strarry[14])
	c_get,_ := hex.DecodeString(strarry[15])
	s1_get,_ := hex.DecodeString(strarry[16])
	s2_get,_ := hex.DecodeString(strarry[17])
	s3_get,_ := hex.DecodeString(strarry[18])
	mess_get_temp,_ := hex.DecodeString(strarry[19])
	mess_get := string(mess_get_temp)
	fmt.Println("mess_get",mess_get)

	//fmt.Println("temp_byte=", temp_byte)
	//temp_arr := make([]byte, 0, 32)
	//temp_arr := BytesCombine1(rkeysstr,pkstr)
	//fmt.Println("temp_arr=",temp_arr)

	//tkstr := strarry[2]
	//vkstr := strarry[3]
	//fmt.Println("first=", rkeysstr)
	//fmt.Println("second=", pkstr)
	//return nil,nil

	//ssig := "abc"
	//mpk := "abc"e
	//message := "123"
	//grpID := "1"
	//gyh: mpk不变，所以此处直接赋值，除了里边的h1，经过测试，里边的ｈ1是变的

	bytes32Ans := make([]byte, 0, 32) //判断是否成功验证
	for i := 0; i < 31; i++ {
		bytes32Ans = append(bytes32Ans, 0)
	}

	var mpk TIBGSMasterPublicKey  // g,h1 是G1  g2~u4是G2上的点　n是GT上的点　
	var ssig TIBGSSIG //c5, e1是G1上的点  c0, c6, e2是G2上的点  e3是GT上的点
	var POK TIBGSPOK  //c, s1, s2, s3 Zr上的数
	//mpk初始化
	mpk.g = pairing.NewG1()
	mpk.h1 = pairing.NewG1()
	mpk.g2 = pairing.NewG2()
	mpk.u0 = pairing.NewG2()
	mpk.u1 = pairing.NewG2()
	mpk.u2 = pairing.NewG2()
	mpk.u3 = pairing.NewG2()
	mpk.u4 = pairing.NewG2()
	mpk.n = pairing.NewGT()
	ssig.c0 = pairing.NewG2()
	ssig.c6 = pairing.NewG2()
	ssig.e2 = pairing.NewG2()
	ssig.c5 = pairing.NewG1()
	ssig.e1 = pairing.NewG1()
	ssig.e3 = pairing.NewGT()
	POK.c = pairing.NewZr()
	POK.s1 = pairing.NewZr()
	POK.s2 = pairing.NewZr()
	POK.s3 = pairing.NewZr()
	/*
	//接收的参数
	var res_len = len(input) - 2000 //整个输入的长度减去固定的参数，即message的长度。
	fmt.Println("len:", res_len)
	var mpk_get_g [128]byte
	var mpk_get_g2 [128]byte
	var mpk_get_u0 [128]byte
	var mpk_get_u1 [128]byte
	var mpk_get_u2 [128]byte
	var mpk_get_u3 [128]byte
	var mpk_get_u4 [128]byte
	var mpk_get_n [128]byte
	var mpk_get_h1 [128]byte

	var ssig_get_c0 [128]byte
	var ssig_get_c5 [128]byte
	var ssig_get_c6 [128]byte
	var ssig_get_e1 [128]byte
	var ssig_get_e2 [128]byte
	var ssig_get_e3 [128]byte
	var pok_get_c [20]byte
	var pok_get_s1 [20]byte
	var pok_get_s2 [20]byte
	var pok_get_s3 [20]byte
	var mess_get [30000]byte

	for i := 0; i <= 127; i++ {
		//fmt.Println(res_byte[i])
		mpk_get_g[i] = input[i]
		mpk_get_g2[i] = input[i+128]
		mpk_get_u0[i] = input[i+128*2]
		mpk_get_u1[i] = input[i+128*3]
		mpk_get_u2[i] = input[i+128*4]
		mpk_get_u3[i] = input[i+128*5]
		mpk_get_u4[i] = input[i+128*6]
		mpk_get_n[i] = input[i+128*7]
		mpk_get_h1[i] = input[i+128*8]

		ssig_get_c0[i] = input[i+128*9]
		ssig_get_c5[i] = input[i+128*10]
		ssig_get_c6[i] = input[i+128*11]
		ssig_get_e1[i] = input[i+128*12]
		ssig_get_e2[i] = input[i+128*13]
		ssig_get_e3[i] = input[i+128*14]
	}
	fmt.Println("mpk_g:", mpk_get_g)
	fmt.Println("mpk_g2:", mpk_get_g2)
	fmt.Println("mpk_u0:", mpk_get_u0)
	fmt.Println("mpk_u1:", mpk_get_u2)
	fmt.Println("mpk_u2:", mpk_get_u2)
	fmt.Println("mpk_u3:", mpk_get_u3)
	fmt.Println("mpk_u4:", mpk_get_u4)
	fmt.Println("mpk_n:", mpk_get_n)
	fmt.Println("mpk_h1:", mpk_get_h1)
	fmt.Println("ssig_get_c0:", ssig_get_c0)
	fmt.Println("ssig_get_c5:", ssig_get_c5)
	fmt.Println("ssig_get_c6:", ssig_get_c6)
	fmt.Println("ssig_get_e1:", ssig_get_e1)
	fmt.Println("ssig_get_e2:", ssig_get_e2)
	fmt.Println("ssig_get_e3:", ssig_get_e3)

	for i := 0; i <= 19; i++ {
		pok_get_c[i] = input[i+128*15]
		//fmt.Println("c=",pok_get_c)
		pok_get_s1[i] = input[i+128*15+20]
		//fmt.Println("s1=",pok_get_s1)
		pok_get_s2[i] = input[i+128*15+20*2]
		//fmt.Println("s2=",pok_get_s2)
		pok_get_s3[i] = input[i+128*15+20*3]
		//fmt.Println("s3=",pok_get_s3)
	}
	fmt.Println("pok_get_c:", pok_get_c)
	fmt.Println("pok_get_s1:", pok_get_s1)
	fmt.Println("pok_get_s2:", pok_get_s2)
	fmt.Println("pok_get_s3:", pok_get_s3)

	for i := 0; i <= res_len-1; i++ {

		mess_get[i] = input[i+128*15+20*4]

	}

	fmt.Println("mess_get:", mess_get[:res_len])

	mess_get_str := string(mess_get[:res_len])

	fmt.Println("mess_get_str:", mess_get_str)
	*/

	//gyh :　将传入的参数分开
	//mpk的h1

	//mpk.h1.SetBytes()
	//println("ddddddddddddd",mpk.g.Bytes())

	mpk.g.SetBytes(g_get[:])
	mpk.g2.SetBytes(g2_get[:])
	mpk.u0.SetBytes(u0_get[:])
	mpk.u1.SetBytes(u1_get[:])
	mpk.u2.SetBytes(u2_get[:])
	mpk.u3.SetBytes(u3_get[:])
	mpk.u4.SetBytes(u4_get[:])
	mpk.n.SetBytes(n_get[:])
	mpk.h1.SetBytes(h1_get[:])
	ssig.c0.SetBytes(c0_get)
	ssig.c5.SetBytes(c5_get)
	ssig.c6.SetBytes(c6_get)
	ssig.e1.SetBytes(e1_get)
	ssig.e2.SetBytes(e2_get)
	ssig.e3.SetBytes(e3_get)
	ssig.pok = POK
	//pok里边的c,s1,s2,s3
	ssig.pok.c.SetBytes(c_get)
	ssig.pok.s1.SetBytes(s1_get)
	ssig.pok.s2.SetBytes(s2_get)
	ssig.pok.s3.SetBytes(s3_get)

	time_start := time.Now()
	grpID := "computer" //默认就是这个
	msg := G(mess_get)
	IG := G(grpID)

	t := pairing.NewZr().Rand()
	M := pairing.NewGT().Rand()
	// fmt.Println("M",M)
	d1 := pairing.NewG1().PowZn(mpk.g, t)        //g^t
	d22 := pairing.NewG2().PowZn(mpk.u1, IG)     // u1^G(grpID)
	d23 := pairing.NewG2().PowZn(mpk.u3, msg)    //u3^m
	d2 := pairing.NewG2().Mul(mpk.u0, d22)       //u0 * u1^G(grpID)
	d2.Mul(d2, d23)                              //u0 * u1^G(grpID) * u3^m
	d2.Mul(d2, ssig.c6)                          //u0 * u1^G(grpID) * u3^m *c6      // ssig需要从byte转化进来
	d2.PowZn(d2, t)                              //(u0 * u1^G(grpID) * u3^m *c6)^t
	zeta := pairing.NewGT().Pair(mpk.h1, mpk.g2) //e(h1, g2)

	zeta.PowZn(zeta, t)                        //e(h1, g2)^t
	zeta.Mul(M, zeta)                          //M * e(h1, g2)^t
	temp1 := pairing.NewGT().Pair(ssig.c5, d2) //e(c5, d2)
	temp2 := pairing.NewGT().Pair(d1, ssig.c0) //e(d1, c0)
	right := pairing.NewGT().Div(temp1, temp2) // e(c5, d2) / e(d1, c0)
	right.Mul(zeta, right)                     // zeta * (e(c5, d2) / e(d1, c0))
	//验证POK
	rr11 := pairing.NewG2().PowZn(mpk.u2, ssig.pok.s1) //u2^s1
	rr12 := pairing.NewG2().PowZn(mpk.u4, ssig.pok.s2) //u4^s2
	negc := pairing.NewZr().Neg(ssig.pok.c)            //-c
	rr13 := pairing.NewG2().PowZn(ssig.c6, negc)       //c6^-c
	rr1 := pairing.NewG2().Mul(rr11, rr12)             //u2^s1 * u4^s2
	rr1.Mul(rr1, rr13)                                 //u2^s1 * u4^s2* c6^-c
	rr21 := pairing.NewG2().PowZn(mpk.g, ssig.pok.s3)  //g^s3
	rr22 := pairing.NewG2().PowZn(ssig.e1, negc)       //e1^-c
	rr2 := pairing.NewG2().Mul(rr21, rr22)             //g^s3 * e1^-c
	hatf := pairing.NewG2().PowZn(mpk.u1, IG)          //u1^G(grpID)
	hatf.Mul(mpk.u0, hatf)                             //u0*u1^G(grpID)
	bigg := pairing.NewGT().Pair(mpk.h1, mpk.g2)
	rr31 := pairing.NewG2().PowZn(hatf, ssig.pok.s3)  //hatf^s3
	rr32 := pairing.NewG2().PowZn(ssig.e2, negc)      //e2^-c
	rr3 := pairing.NewG2().Mul(rr31, rr32)            //hatf^s3 * e2^-c
	tt41 := pairing.NewGT().PowZn(mpk.n, ssig.pok.s1) //n^s1
	tt42 := pairing.NewGT().PowZn(bigg, ssig.pok.s3)  //bigg^s3
	tt43 := pairing.NewGT().PowZn(ssig.e3, negc)      //e3^-c
	tt4 := pairing.NewGT().Mul(tt41, tt42)            //n^s1* bigg^s3
	tt4.Mul(tt4, tt43)                                //n^s1* bigg^s3 *e3^-c
	cc := pairing.NewZr().SetFromStringHash(rr1.String()+rr2.String()+rr3.String()+tt4.String(), sha256.New())
	// fmt.Println("cc",cc)
	// fmt.Println("pok.c",ssig.pok.c)
	tf := M.Equals(right) && cc.Equals(ssig.pok.c)
	println(" ")
	fmt.Println("时间：",time.Since(time_start))
	println("tf=",tf)
	//t2:=time.Now()
	// fmt.Println("time:Verify",t2.Sub(t1))
	//f, _ := os.OpenFile("Verify.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	//fmt.Fprintln(f,t2.Sub(t1))
	//flag1 := []byte("false")
	if tf {
		//flag1 = []byte("true")
		bytes32Ans = append(bytes32Ans, 1)
		fmt.Println("群签名验证成功！！！！！！！！！！！！！")
		return bytes32Ans, nil
	}
	fmt.Println("群签名验证不成功？？？？？？？？")
	bytes32Ans = append(bytes32Ans, 0)
	return bytes32Ans, nil

	//fmt.Print("新建的智能合约调用成功！！！！！！！！！！！！！")

	//return input ,nil

}

type verhfProof struct{}

func (c *verhfProof) RequiredGas(input []byte) uint64 {
	return params.VerProofGas
}

func (c *verhfProof) Run(input []byte) ([]byte, error) {

	fmt.Println("running verProof test ....")
	// start time --ZHOU
	t2 := time.Now()
	length := len(input)
	fmt.Printf("inputData size: ")
	fmt.Println(length)

	// verify proof
	var buffer bytes.Buffer   // Buffer can be write and read with byte
	messageID := []byte{0, 1} // 01 represents verify proof data

	buffer.Write(messageID)
	buffer.Write(input)
	inputData := buffer.Bytes()
	fmt.Printf("inputData: ")
	fmt.Println(inputData)

	// result := C.verifyHighFeeproof(inputData)
	time_st := time.Now()
	result := zktx.VerifyHighFeeProof() // gyh: 暂时没有参数
	//result := 1
	fmt.Println("proof运行时间:",time.Since(time_st))
	bytes32Ans := make([]byte, 0, 32)

	for i := 0; i < 31; i++ {
		bytes32Ans = append(bytes32Ans, 0)
	}

	if result == 1 {
		bytes32Ans = append(bytes32Ans, 1) // verify successfully
	} else if result == 2 {
		bytes32Ans = append(bytes32Ans, 2) // verify unsuccessfully
	} else if result == 3 {
		bytes32Ans = append(bytes32Ans, 3) // exist error
	} else {
		bytes32Ans = append(bytes32Ans, 0) // receive null data
	}

	// end time --ZHOU
	verify_time := time.Since(t2)
	fmt.Println("Appverify_time: ", verify_time)
	fmt.Println(bytes32Ans)
	return bytes32Ans, nil

}

/////////////////////////////////////
//  connect libsnark to verify proof　暂时没用//
/////////////////////////////////////
func checkVerifyConnection(conn net.Conn, err error) bool {
	if err != nil {
		log.Warn(err.Error())
		fmt.Printf("error %v connecting, please check hdsnark\n", conn)
		return false
	}
	fmt.Printf("connected with %v\n", conn)
	return true
}

func localVerifyConnection(inputData []byte) uint32 {
	conn, err := net.Dial("tcp", "127.0.0.1:8032")

	if !checkVerifyConnection(conn, err) {
		return 3
	}

	conn.Write(inputData) // send original data

	receiveData := make([]byte, 2)

	indexEnd, err := conn.Read(receiveData)
	fmt.Printf("receive data: ")
	fmt.Println(receiveData)
	fmt.Printf("receive data index end: ")
	fmt.Println(indexEnd)

	if err != nil {
		log.Warn(err.Error())
		return 3
	}

	// var result uint32
	// resIndex := (int)(unsafe.Sizeof(result))
	// result = uint32(binary.LittleEndian.Uint32(receiveData[0:resIndex]))

	result := (uint32)(receiveData[0] - 48)

	fmt.Printf("receive result: ")
	fmt.Println(result)

	//////////////
	// result = 1, 2, 3 represents true, false, error when verifying proof
	defer conn.Close()
	return result
}
