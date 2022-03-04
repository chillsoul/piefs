package store

import (
	"os"
	"sync"
)

type Volume struct {
	ID            uint64   //volume id
	File          *os.File //volume file
	Size          uint64
	Path          string
	CurrentOffset uint64
	lock          sync.Mutex
}
