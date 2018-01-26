package main

import (
	"testing"

	"github.com/x3art/x3t/xt"
)

func BenchmarkUniverse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = xt.GetUniverse("data/x3_universe.xml")
	}
}
