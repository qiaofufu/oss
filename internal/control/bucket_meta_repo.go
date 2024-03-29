package control

type BucketMetaRepo interface {
	CreateBucket(name string, ownerID int64) (*BucketMeta, error)
	GetBucketByID(id int64) (*BucketMeta, error)
	GetBucketByOwnerID(ownerID int64) ([]*BucketMeta, error)

	DeleteBucket(id int64) error
}

type BucketMeta struct {
	ID        int64  `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
	UpdatedAt int64  `json:"updated_at,omitempty"`

	OwnerID int64 `json:"owner_id,omitempty"`
}
