package control

import (
	"io"
	"os"
	"oss/internal/utils"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// UploadObject upload object to peer
func (c *ctrl) UploadObject(data []*os.File, name string, size int64, bucketID int64, objType ObjectType) (*Object, error) {
	obj := &Object{
		ID:        c.objectIDGenerator.GenerateID(),
		Name:      name,
		Size:      size,
		CreatedAt: time.Now().UnixMilli(),
		UpdatedAt: time.Now().UnixMilli(),
		Type:      objType,
		BucketID:  bucketID,
		FilesNum:  0,
		Files:     data,
		Replicas:  2,
	}

	meta := &ObjectMeta{
		ID:        obj.ID,
		Name:      obj.Name,
		Size:      obj.Size,
		CreatedAt: obj.CreatedAt,
		UpdatedAt: obj.UpdatedAt,
		Type:      obj.Type,
		BucketID:  obj.BucketID,
		FilesNum:  obj.FilesNum,
		Replicas:  obj.Replicas,

		BucketLocation: Location{
			NID:      c.peer.GetNID(),
			Location: c.peer.GetAddr(),
		},
		DataShardsMeta:   make(map[int]BlockMeta),
		ParityShardsMeta: make(map[int]BlockMeta),
	}

	// generate object template files
	dataShards, parityShards, err := c.generateObjectTemplateFiles(obj)
	if err != nil {
		return nil, err
	}

	// pick a control to upload dataShards
	dataShardsPeer, parityShardsPeer := make([][]Operator, len(dataShards)), make([][]Operator, len(parityShards))
	for i := range dataShards {
		peer, err := c.peer.PickByBlock(bucketID, obj.ID, dataShards[i])
		if err != nil {
			return nil, err
		}
		dataShardsPeer[i] = peer
	}

	for i := range parityShards {
		peer, err := c.peer.PickByBlock(bucketID, obj.ID, parityShards[i])
		if err != nil {
			return nil, err
		}
		parityShardsPeer[i] = peer
	}
	// upload dataShards and parityShards
	for i := range dataShards {
		blockMeta, err := c.generateBlockMeta(obj, dataShards[i], dataShardsPeer[i])
		if err != nil {
			return nil, err
		}
		for j := range dataShardsPeer[i] {
			err = dataShardsPeer[i][j].UploadBlock(blockMeta, dataShards[i])
			if err != nil {
				return nil, err
			}
		}
		meta.DataShardsMeta[i] = *blockMeta
	}
	for i := range parityShards {
		blockMeta, err := c.generateBlockMeta(obj, dataShards[i], dataShardsPeer[i])
		if err != nil {
			return nil, err
		}
		for j := range parityShardsPeer[i] {
			err = parityShardsPeer[i][j].UploadBlock(blockMeta, parityShards[i])
			if err != nil {
				return nil, err
			}
		}
		meta.ParityShardsMeta[i] = *blockMeta
	}

	// store object object
	if err = c.objMeta.StoreMeta(meta); err != nil {
		return nil, err
	}
	return obj, nil
}

// DownloadObject download object from peer
func (c *ctrl) DownloadObject(bucketID int64, objectID int64) (*Object, error) {
	meta, err := c.objMeta.GetMeta(bucketID, objectID)
	if err != nil {
		return nil, err
	}
	dataShardNum, parityShardNum := len(meta.DataShardsMeta), len(meta.ParityShardsMeta)
	n := dataShardNum + parityShardNum

	object := &Object{
		ID:        meta.ID,
		Name:      meta.Name,
		Size:      meta.Size,
		CreatedAt: meta.CreatedAt,
		UpdatedAt: meta.UpdatedAt,
		Type:      meta.Type,
		BucketID:  meta.BucketID,
		FilesNum:  meta.FilesNum,
		Replicas:  meta.Replicas,
	}
	dir, err := c.getObjectTempDir(object)
	if err != nil {
		return nil, err
	}

	// create source files
	sourceFiles := make([]*os.File, n)
	for i := range sourceFiles {
		sourceFiles[i], err = os.CreateTemp(dir, strconv.Itoa(i)+sourceSuffix)
		if err != nil {
			return nil, err
		}
	}

	fill := make([]*os.File, n)
	dataShards := make([]*os.File, dataShardNum)
	parityShards := make([]*os.File, parityShardNum)

	// pick a control to download dataShards and parityShards
	dataShardPeer, parityShardPeer, err := c.peer.PickByMeta(meta)
	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}
	wg.Add(n)
	// download dataShards
	for i := range dataShardPeer {
		go func(i int) {
			defer wg.Done()
			for _, peers := range dataShardPeer[i] {
				reader, err := peers.DownloadBlock(meta.DataShardsMeta[i])
				if err != nil {
					continue
				}
				_, err = io.Copy(sourceFiles[i], reader)
				if err != nil {
					continue
				}
				fill[i] = nil
				dataShards[i] = sourceFiles[i]
				return
			}
			fill[i] = sourceFiles[i]
			dataShards[i] = nil
		}(i)
	}
	// download parityShards
	for i := range parityShardPeer {
		go func(i int) {
			defer wg.Done()
			for _, peers := range parityShardPeer[i] {
				reader, err := peers.DownloadBlock(meta.ParityShardsMeta[i])
				if err != nil {
					continue
				}
				_, err = io.Copy(sourceFiles[dataShardNum+i], reader)
				if err != nil {
					continue
				}
				fill[dataShardNum+i] = nil
				parityShards[i] = sourceFiles[dataShardNum+i]
				return
			}
			fill[dataShardNum+i] = sourceFiles[dataShardNum+i]
			parityShards[i] = nil
		}(i)
	}

	wg.Wait()

	// verify and reconstruct
	ok, err := c.divider.Verify(dataShards, parityShards)
	if err != nil {
		return nil, err
	}
	if !ok {
		err := c.divider.Reconstruct(dataShards, parityShards, fill)
		if err != nil {
			return nil, err
		}
	}

	object.Files = sourceFiles[:dataShardNum]
	return object, nil
}

// UploadBlock upload block to local disk
func (c *ctrl) UploadBlock(meta BlockMeta, data io.Reader) error {
	err := c.blockRepo.StoreBlock(meta, data)
	if err != nil {
		return err
	}
	return nil
}

// DownloadBlock download block from local disk
func (c *ctrl) DownloadBlock(meta BlockMeta) (io.Reader, error) {
	data, err := c.blockRepo.GetBlock(meta.BucketID, meta.ObjectID, meta.ID)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *ctrl) generateObjectTemplateFiles(object *Object) (dataShards []*os.File, parityShards []*os.File, err error) {
	// create temp dir
	dir, err := c.getObjectTempDir(object)
	if err != nil {
		return nil, nil, err
	}
	// fill dataShards and parityShards
	dataShard, parityShard := c.divider.CalcShardsNum(object.Size)
	dataShards = make([]*os.File, dataShard)
	parityShards = make([]*os.File, parityShard)
	for i := 0; i < dataShard; i++ {
		dataShards[i], err = os.CreateTemp(dir, strconv.Itoa(i)+dataSuffix)
		if err != nil {
			return nil, nil, err
		}
	}
	for i := 0; i < parityShard; i++ {
		parityShards[i], err = os.CreateTemp(dir, strconv.Itoa(i)+paritySuffix)
		if err != nil {
			return nil, nil, err
		}
	}
	data := make([]io.Reader, len(object.Files))
	for i := range object.Files {
		data[i] = object.Files[i]
	}
	// encode data to dataShards and parityShards
	err = c.divider.Encode(data, object.Size, dataShards, parityShards)
	if err != nil {
		return nil, nil, err
	}
	return dataShards, parityShards, nil
}

func (c *ctrl) getObjectTempDir(object *Object) (string, error) {
	dir := filepath.Join(c.tmpBaseDir, strconv.FormatInt(object.BucketID, 10), strconv.FormatInt(object.ID, 10))
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

func (c *ctrl) generateBlockMeta(object *Object, block *os.File, operators []Operator) (*BlockMeta, error) {
	stat, err := block.Stat()
	if err != nil {
		return nil, err
	}
	checksum, err := utils.ChecksumByFile(block)
	if err != nil {
		return nil, err
	}
	blockMeta := &BlockMeta{
		ID:       c.bucketIDGenerator.GenerateID(),
		BucketID: object.BucketID,
		ObjectID: object.ID,
		Size:     stat.Size(),
		Checksum: checksum,
	}
	for i := range operators {
		blockMeta.Locations = append(blockMeta.Locations, Location{Location: operators[i].Addr(), NID: operators[i].NID()})
	}
	return blockMeta, nil
}
