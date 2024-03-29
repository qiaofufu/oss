package block

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"oss/internal/control"
	"path/filepath"
	"strconv"
)

type store struct {
	baseDir string
}

func (s *store) StoreBlock(meta control.BlockMeta, data io.Reader) error {
	dir, err := s.getBlockDir(meta.BucketID, meta.ObjectID, meta.ID)
	if err != nil {
		return err
	}
	// Store data
	filename := filepath.Join(dir, "data")
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	n, err := io.Copy(file, data)
	if err != nil {
		return err
	}
	if n != meta.Size {
		return fmt.Errorf("write size not match, expect %d, got %d", meta.Size, n)
	}
	// Store meta
	filename = filepath.Join(dir, "object.json")
	metaFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer metaFile.Close()

	return json.NewEncoder(file).Encode(meta)
}

func (s *store) GetBlock(bucketID int64, objectID int64, blockID int64) (io.Reader, error) {
	dir, err := s.getBlockDir(bucketID, objectID, blockID)
	if err != nil {
		return nil, err
	}
	filename := filepath.Join(dir, "data")
	return os.Open(filename)
}

func (s *store) DeleteBlock(bucketID int64, objectID int64, blockID int64) error {
	dir, err := s.getBlockDir(bucketID, objectID, blockID)
	if err != nil {
		return err
	}
	return os.RemoveAll(dir)
}

func (s *store) GetBlockMeta(bucketID int64, objectID int64, blockID int64) (*control.BlockMeta, error) {
	dir, err := s.getBlockDir(bucketID, objectID, blockID)
	if err != nil {
		return nil, err
	}
	filename := filepath.Join(dir, "object.json")
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	meta := &control.BlockMeta{}
	err = json.NewDecoder(file).Decode(meta)
	return meta, err
}

func (s *store) getBlockDir(bucketID int64, objectID int64, blockID int64) (string, error) {
	dir := filepath.Join(
		s.baseDir,
		strconv.FormatInt(bucketID, 10),
		strconv.FormatInt(objectID, 10),
		strconv.FormatInt(blockID, 10),
	)
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}
	return dir, nil
}
