package util

import (
	"fmt"
	"testing"
)

func TestSnowflake_NextVal(t *testing.T) {
	s, _ := NewSnowflake(1)
	for i := 0; i < 100; i++ {
		fmt.Println(s.NextVal())
	}
}
