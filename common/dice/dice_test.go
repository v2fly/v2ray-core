package dice_test

import (
	"math/rand"
	"testing"

	. "github.com/ghxhy/v2ray-core/v5/common/dice"
)

func BenchmarkRoll1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Roll(1)
	}
}

func BenchmarkRoll20(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Roll(20)
	}
}

func BenchmarkIntn1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rand.Intn(1) //nolint:staticcheck
	}
}

func BenchmarkIntn20(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rand.Intn(20)
	}
}

func BenchmarkInt63(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = uint16(rand.Int63() >> 47)
	}
}

func BenchmarkInt31(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = uint16(rand.Int31() >> 15)
	}
}

func BenchmarkIntn(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = uint16(rand.Intn(65536))
	}
}
