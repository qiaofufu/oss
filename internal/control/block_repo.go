package control

import "io"

type BlockRepo interface {
	StoreBlock(meta BlockMeta, data io.Reader) error
	GetBlock(bucketID int64, objectID int64, blockID int64) (io.Reader, error)
	DeleteBlock(bucketID int64, objectID int64, blockID int64) error
	GetBlockMeta(bucketID int64, objectID int64, blockID int64) (*BlockMeta, error)
}

type BlockMeta struct {
	ID        int64      `json:"id,omitempty"`
	BucketID  int64      `json:"bucket_id,omitempty"`
	ObjectID  int64      `json:"object_id,omitempty"`
	Size      int64      `json:"size,omitempty"`
	Checksum  uint32     `json:"checksum,omitempty"`
	CreatedAt int64      `json:"created_at,omitempty"`
	UpdatedAt int64      `json:"updated_at,omitempty"`
	Path      string     `json:"path,omitempty"`
	Locations []Location `json:"locations,omitempty"`
}

type Location struct {
	NID      int64
	Location string
}
