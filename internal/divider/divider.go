package divider

import (
	"errors"
)

var (
	ErrInvalidShardNumber    = errors.New("invalid shard number")
	ErrResetFileOffsetFailed = errors.New("reset file offset failed")
)

const (
	MaxSize = 1024 * 1024 * 1
)

type Option struct {
	Strategy func(size int64) (int, int)
}

func WithStrategy(strategy func(size int64) (int, int)) Option {
	return Option{Strategy: strategy}
}

func (o Option) apply(opt *Option) {
	if o.Strategy != nil {
		opt.Strategy = o.Strategy
	}
}

func defaultStrategy(size int64) (int, int) {
	dataShard := int(size / MaxSize)
	if size%MaxSize != 0 {
		dataShard++
	}
	return dataShard, dataShard / 2
}
