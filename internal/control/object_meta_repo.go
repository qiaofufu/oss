package control

import "os"

const (
	ObjectTypeTextPlain       ObjectType = "text/plain"
	ObjectTypeTextHtml        ObjectType = "text/html"
	ObjectTypeImageJpeg       ObjectType = "image/jpeg"
	ObjectTypeImagePng        ObjectType = "image/png"
	ObjectTypeImageGif        ObjectType = "image/gif"
	ObjectTypeImageBmp        ObjectType = "image/bmp"
	ObjectTypeImageWebp       ObjectType = "image/webp"
	ObjectTypeAudioMpeg       ObjectType = "audio/mpeg"
	ObjectTypeAudioWav        ObjectType = "audio/wav"
	ObjectTypeAudioOgg        ObjectType = "audio/ogg"
	ObjectTypeVideoMpeg       ObjectType = "video/mpeg"
	ObjectTypeVideoMp4        ObjectType = "video/mp4"
	ObjectTypeVideoWebm       ObjectType = "video/webm"
	ObjectTypeVideoOgg        ObjectType = "video/ogg"
	objectTypeApplicationJson ObjectType = "application/json"
	ObjectTypeApplicationXml  ObjectType = "application/xml"
	ObjectTypeStreamOctet     ObjectType = "stream/octet"
)

// ObjectType is the type of object.
type ObjectType string

type ObjectMeta struct {
	ID               int64             `json:"id"`
	Name             string            `json:"name"`
	Size             int64             `json:"size"`
	CreatedAt        int64             `json:"created_at"`
	UpdatedAt        int64             `json:"updated_at"`
	Type             ObjectType        `json:"type"`
	BucketID         int64             `json:"bucket_id"`
	FilesNum         int               `json:"files_num"`
	BucketLocation   Location          `json:"bucket_location"`
	DataShardsMeta   map[int]BlockMeta `json:"data_shards_location"`
	ParityShardsMeta map[int]BlockMeta `json:"parity_shards_location"`
	Replicas         int               `json:"replicas"`
}

type ObjectMetaRepo interface {
	StoreMeta(meta *ObjectMeta) error
	GetMeta(bucketID int64, objectID int64) (*ObjectMeta, error)
	GetMetaList(bucketID int64) ([]*ObjectMeta, error)
}

const (
	sourceSuffix = "source"
	dataSuffix   = "data"
	paritySuffix = "parity"
)

type Object struct {
	ID        int64
	Name      string
	Size      int64
	CreatedAt int64
	UpdatedAt int64
	Type      ObjectType
	BucketID  int64
	FilesNum  int
	Files     []*os.File
	Replicas  int
}
