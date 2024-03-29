package control

import (
	"io"
	"os"
)

type Peer interface {
	Picker
	Discover
	Register
}

type Picker interface {
	PickByBucket(param any) (peers []Operator, err error)
	PickByBlock(bucketID, objectID int64, block *os.File) (peers []Operator, err error)
	PickByObject(param any) (peers []Operator, err error)

	PickByMeta(meta *ObjectMeta) (dataShardPeer [][]Operator, parityShardPeer [][]Operator, err error)
}

type Operator interface {
	UploadBlock(meta *BlockMeta, data io.Reader) error
	DownloadBlock(meta BlockMeta) (data io.Reader, err error)

	// Peer Info Getter

	Addr() string
	NID() int64
}

type Discover interface {
	Discover() (peers []Operator, err error)
	AddPeer(peer Operator) error
	RemovePeer(peer Operator) error
	UpdatePeer(peer Operator) error
}

type Register interface {
	Register(selfIp string) error
	GetNID() int64
	GetAddr() string
}

type ctrl struct {
	tmpBaseDir        string
	bucketIDGenerator UniqueIDGenerator
	objectIDGenerator UniqueIDGenerator

	blockRepo BlockRepo
	objMeta   ObjectMetaRepo

	divider Divider
	peer    Peer
}
