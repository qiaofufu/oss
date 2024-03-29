package snowflake

import (
	"sync"
	"time"
)

const (
	NodeBit     = 10
	SequenceBit = 12
	MaxSequence = 1 << SequenceBit
	MaxNode     = 1 << NodeBit
	TimeShift   = NodeBit + SequenceBit
	NodeShift   = SequenceBit
)

var onSyncTime = func() int64 {
	return time.Now().UnixNano()
}

func SetSyncTimeFunc(f func() int64) {
	onSyncTime = f
}

type Snowflake struct {
	sync.Mutex
	node      int64
	sequence  int64
	startTime int64
	timestamp int64
}

func NewSnowflake(node int64, startTime int64) *Snowflake {
	if node < 0 || node >= MaxNode {
		panic("node number must be between 0 and 1023")
	}
	return &Snowflake{
		node: node,
	}
}

func (s *Snowflake) GenerateID() int64 {
	s.Lock()
	defer s.Unlock()

	now := onSyncTime()
	if now < s.startTime {
		panic("invalid system time")
	}
	if now == s.timestamp {
		s.sequence = (s.sequence + 1) % MaxSequence
		if s.sequence == 0 {
			for now <= s.timestamp {
				now = onSyncTime()
			}
		}
	} else {
		s.sequence = 0
	}
	s.timestamp = now
	return (now-s.startTime)<<TimeShift | s.node<<NodeShift | s.sequence
}
