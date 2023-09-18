package sabercache_server

import (
	"context"
	"fmt"
	"log"
	"net"
	pb "sabercache_server/sabercachepb"
	"sabercache_server/util"
	"strings"
	"sync"

	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedSaberCacheServer
	addr       string
	status     bool
	mu         sync.Mutex
	stopSignal chan error
}

func NewServer() (*Server, error) {
	util.InitConst()
	addr := util.RPCAddr
	if !util.ValidPeerAddr(addr) {
		return nil, fmt.Errorf("invalid addr %s , it should be x.x.x.x:port", addr)
	}
	return &Server{addr: addr}, nil
}

func (s *Server) Start() error {
	s.mu.Lock()
	if s.status == true {
		s.mu.Unlock()
		return fmt.Errorf("server already started")
	}
	s.status = true
	s.stopSignal = make(chan error)

	port := strings.Split(s.addr, ":")[1]
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterSaberCacheServer(grpcServer, s)

	go func() {
		err := Register("sabercache", s.addr, s.stopSignal)
		if err != nil {
			log.Fatalf(err.Error())
		}
		close(s.stopSignal)
		err = lis.Close()
		if err != nil {
			log.Fatalf(err.Error())
		}
		log.Printf("[%s] Revoke service and close tcp socket ok.", s.addr)
	}()
	s.mu.Unlock()
	if err := grpcServer.Serve(lis); s.status && err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
	return nil
}

func (s *Server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	key := in.GetKey()
	resp := &pb.GetResponse{}

	log.Printf("[sabercache_svr %s] Recv RPC Request - (%s)", s.addr, key)
	if key == "" {
		return resp, fmt.Errorf("key required")
	}
	view, err := sabercache.Get(key)
	if err != nil {
		return resp, err
	}
	resp.Value = view.ByteSlice()
	return resp, nil
}

func (s *Server) GetAll(ctx context.Context, in *pb.GetAllRequest) (*pb.GetAllResponse, error) {
	resp := &pb.GetAllResponse{}

	log.Printf("[sabercache_svr %s] Recv RPC Request", s.addr)
	kv := sabercache.GetAll()
	for _, v := range kv {
		resp.Kv = append(resp.Kv, &pb.KeyValue{Key: v.Key, Value: v.Value.(ByteView).ByteSlice()})
	}
	return resp, nil
}

func (s *Server) Set(ctx context.Context, in *pb.SetRequest) (*pb.SetResponse, error) {
	key, value, ttl := in.GetKey(), in.GetValue(), in.GetTtl()
	resp := &pb.SetResponse{}
	log.Printf("[sabercache_svr %s] Recv RPC Request - (%s)", s.addr, key)
	if key == "" {
		return resp, fmt.Errorf("key required")
	}
	resp.Ok = sabercache.Set(key, ByteView{value}, ttl)
	return resp, nil
}

func (s *Server) TTL(ctx context.Context, in *pb.TTLRequest) (*pb.TTLResponse, error) {
	key := in.GetKey()
	resp := &pb.TTLResponse{}
	log.Printf("[sabercache_svr %s] Recv RPC Request - (%s)", s.addr, key)
	if key == "" {
		return resp, fmt.Errorf("key required")
	}
	resp.Ttl = sabercache.TTL(key)
	return resp, nil
}
func (s *Server) Save(ctx context.Context, in *pb.SaveRequest) (*pb.SaveResponse, error) {
	resp := &pb.SaveResponse{}
	log.Printf("[sabercache_svr %s] Recv RPC Request save", s.addr)
	resp.Ok = sabercache.Save()
	return resp, nil
}
