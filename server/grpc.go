package server

import (
	"context"
	"log"
	"net"

	"github.com/gevulotnetwork/cloud-orchestrator/cloud"
	"github.com/gevulotnetwork/cloud-orchestrator/server/api/v1/pb"
	"google.golang.org/grpc"
)

type Config struct {
	ListenAddr string

	Orchestrator *cloud.Orchestrator
}

type server struct {
	pb.UnimplementedOrchestratorServer

	orchestrator *cloud.Orchestrator
}

func (srv *server) PrepareImage(ctx context.Context, req *pb.PrepareImageRequest) (*pb.PrepareImageResponse, error) {
	log.Printf("preparing program image \"%s:%s\"", req.Program, req.Path)

	err := srv.orchestrator.PrepareProgramImage(req.Program, req.Path)
	if err != nil {
		log.Printf("error: failed to prepare program %q image: %v", req.Program, err)
	} else {
		log.Printf("prepared program image for %q", req.Program)
	}

	// TODO: Can Gevulot main node benefit from knowing about potential error?
	// If the image preparation fails, the instance creation fails as well and
	// no data will be lost.
	return &pb.PrepareImageResponse{}, nil
}

func (srv *server) CreateInstance(ctx context.Context, req *pb.CreateInstanceRequest) (*pb.CreateInstanceResponse, error) {
	instanceId, err := srv.orchestrator.CreateInstance(req.Program)
	if err != nil {
		log.Printf("error: failed to create instance for program %q: %v", req.Program, err)
	} else {
		log.Printf("created instance for program %q: %q", req.Program, instanceId)
	}

	// TODO: Can Gevulot main node benefit from knowing about potential error?
	// If instance creation fails, nothing pulls tasks from Gevulot node's
	// scheduling queue and no data loss will occur.
	return &pb.CreateInstanceResponse{InstanceId: instanceId}, nil
}

func (srv *server) DeleteInstance(ctx context.Context, req *pb.DeleteInstanceRequest) (*pb.DeleteInstanceResponse, error) {
	err := srv.orchestrator.DeleteInstance(req.InstanceId)
	if err != nil {
		log.Printf("error: failed to delete instance %q: %v", req.InstanceId, err)
	} else {
		log.Printf("deleted instance %q", req.InstanceId)
	}

	// TODO: Can Gevulot main node benefit from knowing about potential error?
	// If instance creation fails, nothing pulls tasks from Gevulot node's
	// scheduling queue and no data loss will occur.
	return &pb.DeleteInstanceResponse{}, nil
}

func Start(cfg Config) {
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = "localhost:0"
	}

	if cfg.Orchestrator == nil {
		log.Fatal("Orchestrator not configured")
	}

	listener, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		panic(err)
	}

	log.Printf("listening at %q", listener.Addr())

	s := grpc.NewServer()
	pb.RegisterOrchestratorServer(s, &server{
		orchestrator: cfg.Orchestrator,
	})

	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
