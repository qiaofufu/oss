package control

import (
	"io"
	"os"
)

type Divider interface {
	CalcShardsNum(size int64) (dataShard int, parityShard int)
	Encode(data []io.Reader, size int64, dataShards []*os.File, parityShards []*os.File) error
	Verify(dataShards []*os.File, parityShards []*os.File) (bool, error)
	Reconstruct(dataShards []*os.File, parityShards []*os.File, fill []*os.File) error
}
