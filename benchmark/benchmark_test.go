package benchmark

import (
	"testing"
)

func TestBenchmark(t *testing.T) {
	Benchmark("localhost", 8080, 12, 1000, 1024*100)
}
