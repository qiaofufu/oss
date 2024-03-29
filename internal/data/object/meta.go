package object

import (
	"encoding/json"
	"errors"
	"os"
	"oss/internal/control"
	"oss/internal/utils"
	"path/filepath"
	"strconv"
)

var (
	ErrMetaAlreadyExists = errors.New("object already exists")
	ErrMetaNotFound      = errors.New("object not found")
)

func NewObjectMetaStore(baseDir string) control.MetaStore {
	err := utils.CreateDirIfNotExists(baseDir)
	if err != nil {
		panic(err)
	}

	return &store{baseDir: baseDir}
}

type store struct {
	baseDir string
}

func (o *store) GetMetaList(bucketID int64) ([]*control.ObjectMeta, error) {
	var (
		list   = make([]*control.ObjectMeta, 0)
		walkFn = func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || filepath.Ext(path) != ".json" {
				return nil
			}
			f, err := os.Open(info.Name())
			if err != nil {
				return err
			}
			meta := &control.ObjectMeta{}
			err = json.NewDecoder(f).Decode(meta)
			list = append(list, meta)
			return nil
		}
	)
	bucketDir := filepath.Join(o.baseDir, strconv.FormatInt(bucketID, 10))
	if err := filepath.Walk(bucketDir, walkFn); err != nil {
		return nil, err
	}
	return list, nil
}

func (o *store) StoreMeta(meta *control.ObjectMeta) error {
	if _, err := o.GetMeta(meta.BucketID, meta.ID); err == nil {
		return ErrMetaAlreadyExists
	}
	dir := filepath.Join(o.baseDir, strconv.FormatInt(meta.BucketID, 10))
	if err := utils.CreateDirIfNotExists(dir); err != nil {
		return err
	}
	filename := filepath.Join(dir, strconv.FormatInt(meta.ID, 10)+".json")
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	err = json.NewEncoder(f).Encode(meta)
	if err != nil {
		return err
	}
	return nil
}

func (o *store) GetMeta(bucketID int64, objectID int64) (*control.ObjectMeta, error) {
	dir := filepath.Join(o.baseDir, strconv.FormatInt(bucketID, 10))
	if err := utils.CreateDirIfNotExists(dir); err != nil {
		return nil, err
	}
	filename := filepath.Join(dir, strconv.FormatInt(objectID, 10)+".json")
	f, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrMetaNotFound
		}
		return nil, err
	}
	meta := &control.ObjectMeta{}
	err = json.NewDecoder(f).Decode(meta)
	if err != nil {
		return nil, err
	}
	return meta, nil
}
