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
	epoch          = int64(1648173760000)              // 设置起始时间(时间戳/毫秒)：2021-03-25 10:02:40，有效期557年
	timestampBits  = uint(44)                          // 时间戳占用位数 	// 数据中心id所占位数
	workerIdBits   = uint(7)                           // 机器id所占位数
	sequenceBits   = uint(12)                          // 序列所占的位数
	timestampMax   = int64(-1 ^ (-1 << timestampBits)) // 时间戳最大值 	// 支持的最大数据中心id数量
	workerIdMax    = int64(-1 ^ (-1 << workerIdBits))  // 支持的最大机器id数量
	sequenceMask   = int64(-1 ^ (-1 << sequenceBits))  // 支持的最大序列id数量
	workerIdShift  = sequenceBits                      // 机器id左移位数
	timestampShift = sequenceBits + workerIdBits       // 时间戳左移位数
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
	now := time.Now().UnixMilli() // 转毫秒
	if s.timestamp == now {
		// 当同一时间戳（精度：毫秒）下多次生成id会增加序列号
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			// 如果当前序列超出12bit长度，则需要等待下一毫秒
			// 下一毫秒将使用sequence:0
			time.Sleep(time.Microsecond)
		}
	} else {
		// 不同时间戳（精度：毫秒）下直接使用序列号：0
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
