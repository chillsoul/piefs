package util

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type Snowflake struct {
	sync.Mutex
	timestamp int64
	workerId  int64
	sequence  int64
}

const (
	epoch          = int64(1648173760000)              // start: 2021-03-25 10:02:40, expire: 2300-12-20 01:13:02
	timestampBits  = uint(43)                          // timestamp bits
	workerIdBits   = uint(7)                           // worker bits
	sequenceBits   = uint(13)                          // sequence bits
	timestampMax   = int64(-1 ^ (-1 << timestampBits)) // max timestamp value
	workerIdMax    = int64(-1 ^ (-1 << workerIdBits))  // max worker num
	sequenceMask   = int64(-1 ^ (-1 << sequenceBits))  // max sequence value
	workerIdShift  = sequenceBits                      // workId lsh bits
	timestampShift = sequenceBits + workerIdBits       // timestamp lsh bits
)

func NewSnowflake(workId int64) (*Snowflake, error) {
	if workId > workerIdMax {
		return nil, errors.New("workId too large")
	}
	return &Snowflake{workerId: workId}, nil
}
func (s *Snowflake) NextVal() uint64 {
	s.Lock()
	defer s.Unlock()
	now := time.Now().UnixMilli() // millisecond
	for s.timestamp > now {
		//clock callback
		time.Sleep(time.Millisecond)
		now = time.Now().UnixMilli()
	}
	if s.timestamp == now {
		// concurrent id generation
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			// now time's sequence is ran out
			time.Sleep(time.Millisecond)
		}
	} else {
		// new timestamp now
		s.sequence = 0
	}
	t := now - epoch
	if t > timestampMax {
		panic(fmt.Sprintf("epoch %d must be between 0 and %d", t, timestampMax))
		return 0
	}
	s.timestamp = now
	r := uint64((t)<<timestampShift | (s.workerId << workerIdShift) | (s.sequence))
	return r
}
