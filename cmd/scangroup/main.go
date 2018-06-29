package main

import (
	"log"
	"net"

	"gopkg.linkai.io/v1/repos/am/services/scangroup/protoc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.linkai.io/v1/repos/am/services/scangroup"
	"gopkg.linkai.io/v1/repos/am/services/scangroup/store/pg"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	store := pg.New()

	service := scangroup.New(store)
	s := grpc.NewServer()
	protoc.RegisterScanGroupServer(s, service)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
