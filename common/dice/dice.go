// Package dice contains common functions to generate random number.
// It also initialize math/rand with the time in seconds at launch time.
package dice

import (
	crand "crypto/rand"
	"io"
	"math/big"
	"math/rand"
	"time"

	"github.com/v2fly/v2ray-core/v5/common"
)

// Roll returns a non-negative number between 0 (inclusive) and n (exclusive).
func Roll(n int) int {
	if n == 1 {
		return 0
	}
	return rand.Intn(n)
}

// RollWith returns a non-negative number between 0 (inclusive) and n (exclusive).
// Use random as the random source, if read fails, it panics.
func RollWith(n int, random io.Reader) int {
	if n == 1 {
		return 0
	}
	mrand, err := crand.Int(random, big.NewInt(int64(n)))
	common.Must(err)
	return int(mrand.Int64())
}

// Roll returns a non-negative number between 0 (inclusive) and n (exclusive).
func RollDeterministic(n int, seed int64) int {
	if n == 1 {
		return 0
	}
	return rand.New(rand.NewSource(seed)).Intn(n)
}

// RollUint16 returns a random uint16 value.
func RollUint16() uint16 {
	return uint16(rand.Int63() >> 47)
}

func RollUint64() uint64 {
	return rand.Uint64()
}

func NewDeterministicDice(seed int64) *DeterministicDice {
	return &DeterministicDice{rand.New(rand.NewSource(seed))}
}

type DeterministicDice struct {
	*rand.Rand
}

func (dd *DeterministicDice) Roll(n int) int {
	if n == 1 {
		return 0
	}
	return dd.Intn(n)
}

func init() {
	rand.Seed(time.Now().Unix())
}
