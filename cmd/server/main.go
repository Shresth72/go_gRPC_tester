package main

import (
	"log"
	"net"
	"os"
	"os/exec"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	echopb "github.com/Shresth72/go_gRPC_tester/proto/echo"
	initpb "github.com/Shresth72/go_gRPC_tester/proto/init"
)

type server struct {
	initpb.UnimplementedInitServiceServer
	echopb.UnimplementedEchoServiceServer
	rustProcess *exec.Cmd
	stdinPipe   *os.File
	mu          sync.Mutex
}

func main() {
	rustCmd := exec.Command("path/to/rust_binary")

	stdinPipe, err := rustCmd.StdinPipe()
	if err != nil {
		log.Fatalf("Failed to get stdin pipe: %v", err)
	}

	rustCmd.Stdout = os.Stdout
	rustCmd.Stderr = os.Stderr

	if err := rustCmd.Start(); err != nil {
		log.Fatalf("Failed to start Rust binary: %v", err)
	}

	s := &server{
		rustProcess: rustCmd,
		stdinPipe:   stdinPipe.(*os.File),
	}

	grpcServer := grpc.NewServer()
	initpb.RegisterInitServiceServer(grpcServer, s)
	echopb.RegisterEchoServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":5051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Server is listening on :5051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

	if err := rustCmd.Wait(); err != nil {
		log.Fatalf("Rust process exited with error: %v", err)
	}
}
