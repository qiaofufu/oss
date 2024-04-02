package control

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"io"
	"net"
	"oss/internal/proto"
)

type PeerServer struct {
	proto.UnimplementedOperatorServer
	coreCtrl *ctrl
	addr     string
	nid      int64
}

func (p *PeerServer) DownloadBlock(request *proto.DownloadBlockRequest, server proto.Operator_DownloadBlockServer) error {
	data, err := p.coreCtrl.DownloadBlock(request.BucketID, request.ObjectID, request.BlockID)
	if err != nil {
		log.Debugf("download block failed: %v", err)
		server.Send(&proto.DownloadBlockResponse{Success: false})
		return nil
	}
	block, err := io.ReadAll(data)
	if err != nil {
		log.Debugf("read block data failed: %v", err)
		server.Send(&proto.DownloadBlockResponse{Success: false})
		return nil
	}

	server.Send(&proto.DownloadBlockResponse{Success: true, Block: block})
	return nil
}

func (p *PeerServer) UploadBlock(server proto.Operator_UploadBlockServer) error {
	request, err := server.Recv()
	if err != nil {
		log.Debugf("recv upload block request failed: %v", err)
		return nil
	}
	meta := &BlockMeta{
		BucketID:  request.BlockMeta.BlockID,
		ObjectID:  request.BlockMeta.ObjectID,
		ID:        request.BlockMeta.BlockID,
		Size:      request.BlockMeta.Size,
		Checksum:  request.BlockMeta.Checksum,
		CreatedAt: request.BlockMeta.CreatedAt,
		UpdatedAt: request.BlockMeta.UpdatedAt,
		Path:      request.BlockMeta.Path,
		Locations: make([]Location, 0),
	}
	for _, location := range request.BlockMeta.Locations {
		meta.Locations = append(meta.Locations, Location{
			NID:  location.NID,
			Addr: location.Addr,
		})

	}
	reader := bytes.NewReader(request.Block)
	if err := p.coreCtrl.UploadBlock(*meta, reader); err != nil {
		log.Debugf("upload block failed: %v", err)
		server.SendAndClose(&proto.UploadBlockResponse{Success: false, Message: err.Error()})
		return nil
	}
	server.SendAndClose(&proto.UploadBlockResponse{Success: true})
	return nil
}

func NewPeerServer(addr string, nid int64, coreCtrl *ctrl) *PeerServer {
	return &PeerServer{
		addr:     addr,
		nid:      nid,
		coreCtrl: coreCtrl,
	}
}

func RunServer(server *PeerServer) {
	ln, err := net.Listen("tcp", server.addr)
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	proto.RegisterOperatorServer(s, server)
	log.Debugf("server listen on %s", server.addr)
	if err := s.Serve(ln); err != nil {
		panic(err)
	}
}
