package main

import (
	"context"
	"encoding/json"
	"fmt"
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

func (s *server) SendInit(ctx context.Context, in *initpb.InitRequest) (*initpb.InitResponse, error) {
  s.mu.Lock()
  defer s.mu.Unlock()

  message, err := json.Marshal(in)
  if err != nil {
    return nil, fmt.Errorf("failed to marshal message: %w", err)
  }

  _, err = s.stdinPipe.WriteString(string(message) + "\n")
  if err != nil {
    return nil, err
  }

  println("writing echo: %s", string(message))

  return &initpb.InitResponse {
    Src: in.Dest,
    Dest: in.Src,
    Body: &initpb.InitResponseBody {
      Type: "init_ok",
      MsgId: 0,
      InReplyTo: nil,
    },  
  }, nil
}

func (s *server) SendEcho(ctx context.Context, in *echopb.EchoRequest) (*echopb.EchoResponse, error) {
  s.mu.Lock()
  defer s.mu.Unlock()

  message, err := json.Marshal(in)
  if err != nil {
    return nil, fmt.Errorf("failed to marshal message: %w", err)
  }

  _, err = s.stdinPipe.WriteString(string(message) + "\n")
  if err != nil {
    return nil, err
  }

  return &echopb.EchoResponse {
    Src: in.Dest,
    Dest: in.Src,
    Body: &echopb.EchoResponseBody {
      Type: "echo_ok",
      Echo: in.Body.Echo,
      MsgId: in.Body.MsgId,
      InReplyTo: in.Body.MsgId,
    },
  }, nil
}

func main() {
	rustCmd := exec.Command("/home/shrestha/rust/distributed_systems/target/debug/echo")

	stdinPipe, err := rustCmd.StdinPipe()
	if err != nil {
		log.Fatalf("Failed to get stdin pipe: %v", err)
	}

	rustCmd.Stdout = os.Stdout
	rustCmd.Stderr = os.Stderr

	if err := rustCmd.Start(); err != nil {
		log.Fatalf("Failed to start Rust binary: %v", err)
	}

	s := &server {
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
