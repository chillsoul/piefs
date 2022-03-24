package benchmark

import (
	"testing"
)

func TestBenchmark(t *testing.T) {
	Benchmark("localhost", 8080, 12, 10000, 1024*50)
}
