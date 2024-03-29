package control

type UniqueIDGenerator interface {
	GenerateID() int64
}
