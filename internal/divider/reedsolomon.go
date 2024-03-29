package divider

import (
	"github.com/klauspost/reedsolomon"
	"io"
	"os"
	"oss/internal/control"
)

type reedSolomon struct {
	opt Option
}

func NewReedSolomon(opt ...Option) control.Divider {
	c := &reedSolomon{
		opt: Option{
			Strategy: defaultStrategy,
		},
	}
	for _, o := range opt {
		o.apply(&c.opt)
	}
	return c
}

// CalcShardsNum Calculate the number of dataShards and parityShards
func (c *reedSolomon) CalcShardsNum(size int64) (dataShard int, parityShard int) {
	return c.opt.Strategy(size)
}

// Encode the data to dataShards and parityShards
func (c *reedSolomon) Encode(data []io.Reader, size int64, dataShards []*os.File, parityShards []*os.File) error {
	dataShard, parityShard := c.opt.Strategy(size)
	if len(dataShards) != dataShard || len(parityShards) != parityShard {
		return ErrInvalidShardNumber
	}
	enc, err := reedsolomon.NewStream(dataShard, parityShard)
	if err != nil {
		return err
	}
	// Create a buffer to store the data
	buf := io.MultiReader(data...)

	// Split the data into dataShard
	tmpWriter := make([]io.Writer, dataShard)
	for i := range tmpWriter {
		tmpWriter[i] = dataShards[i]
	}
	if err = enc.Split(buf, tmpWriter, size); err != nil {
		return err
	}
	// Reset the file offset to 0
	for i := range dataShards {
		_, err := dataShards[i].Seek(0, 0)
		if err != nil {
			return err
		}
	}
	// Encode parity
	tmpReader := convertFilesToReaders(dataShards)
	tmpWriter = convertFilesToWriters(parityShards)
	if err = enc.Encode(tmpReader, tmpWriter); err != nil {
		return err
	}
	// Reset the file offset to 0
	if err = filesSeek(parityShards, 0, 0); err != nil {
		return ErrResetFileOffsetFailed
	}
	if err = filesSeek(dataShards, 0, 0); err != nil {
		return ErrResetFileOffsetFailed
	}

	return nil
}

// Verify the dataShards and parityShards
func (c *reedSolomon) Verify(dataShards []*os.File, parityShards []*os.File) (bool, error) {
	dataShard, parityShard := c.opt.Strategy(0)
	if len(dataShards) != dataShard || len(parityShards) != parityShard {
		return false, ErrInvalidShardNumber
	}
	enc, err := reedsolomon.NewStream(dataShard, parityShard)
	if err != nil {
		return false, err
	}
	tmpReader := convertFilesToReaders(dataShards, parityShards)
	res, err := enc.Verify(tmpReader)
	if err != nil {
		return false, err
	}
	// Reset the file offset to 0
	if err = filesSeek(dataShards, 0, 0); err != nil {
		return false, ErrResetFileOffsetFailed
	}
	if err = filesSeek(parityShards, 0, 0); err != nil {
		return false, ErrResetFileOffsetFailed
	}
	return res, nil
}

// Reconstruct the data from dataShards and parityShards to fill
func (c *reedSolomon) Reconstruct(dataShards []*os.File, parityShards []*os.File, fill []*os.File) error {
	dataShard, parityShard := c.opt.Strategy(0)
	if len(dataShards) != dataShard || len(parityShards) != parityShard {
		return ErrInvalidShardNumber
	}
	enc, err := reedsolomon.NewStream(dataShard, parityShard)
	if err != nil {
		return err
	}

	tmpReader := convertFilesToReaders(dataShards, parityShards)
	tmpWriter := convertFilesToWriters(fill)

	err = enc.Reconstruct(tmpReader, tmpWriter)
	if err != nil {
		return err
	}
	// Reset the file offset to 0
	if err = filesSeek(fill, 0, 0); err != nil {
		return ErrResetFileOffsetFailed
	}
	if err = filesSeek(dataShards, 0, 0); err != nil {
		return ErrResetFileOffsetFailed
	}
	if err = filesSeek(parityShards, 0, 0); err != nil {
		return ErrResetFileOffsetFailed
	}

	return nil
}

func filesSeek(files []*os.File, offset int64, whence int) error {
	for _, f := range files {
		if f == nil {
			continue
		}
		_, err := f.Seek(offset, whence)
		if err != nil {
			return err
		}
	}
	return nil
}

func convertFilesToReaders(files ...[]*os.File) []io.Reader {
	n := 0
	for i := range files {
		n += len(files[i])
	}
	readers := make([]io.Reader, 0, n)
	for i := range files {
		for j := range files[i] {
			if files[i][j] == nil {
				readers = append(readers, nil)
				continue
			}
			readers = append(readers, files[i][j])
		}
	}
	return readers
}

func convertFilesToWriters(files ...[]*os.File) []io.Writer {
	n := 0
	for i := range files {
		n += len(files[i])
	}
	writers := make([]io.Writer, 0, n)
	for i := range files {
		for j := range files[i] {
			if files[i][j] == nil {
				writers = append(writers, nil)
				continue
			}
			writers = append(writers, files[i][j])
		}
	}
	return writers
}
