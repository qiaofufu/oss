package utils

import (
	"hash/crc32"
	"io"
	"os"
)

func ChecksumByFile(f *os.File) (uint32, error) {
	hash := crc32.NewIEEE()
	stat, err := f.Stat()
	if err != nil {
		return 0, err
	}
	n, err := io.Copy(hash, f)
	if err != nil {
		return 0, err
	}
	if n != stat.Size() {
		return 0, io.ErrShortWrite
	}
	sum := hash.Sum32()
	_, err = f.Seek(0, 0)
	if err != nil {
		return 0, err
	}
	return sum, nil
}
