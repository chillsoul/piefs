package benchmark

import (
	"testing"
)

func TestBenchmark(t *testing.T) {
	Benchmark("localhost", 8080, 12, 100, 1024*100)
}
