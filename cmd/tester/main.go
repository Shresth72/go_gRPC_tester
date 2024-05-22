package main

import (
	"context"
  "log"
	"time"

	echopb "github.com/Shresth72/go_gRPC_tester/proto/echo"
	initpb "github.com/Shresth72/go_gRPC_tester/proto/init"
	"google.golang.org/grpc"
)

func main() {
  conn, err := grpc.Dial("localhost:5051", grpc.WithInsecure(), grpc.WithBlock())
  if err := nil {
    log.Fatalf("Failed to listen: %v", err)
  }
  defer conn.Close()

  initClient := initpb.NewInitServiceClient(conn)
  echoClient := echopb.NewEchoServiceClient(conn)

  ctx, cancel := context.WithTimeout(context.Background(), time.Second)
  defer cancel()

  initReq := &initpb.InitRequest {
    Src: "n1",
    Dest: "n2",
    Body: &initpb.InitRequestBody {
      Type: "init",
      NodeId: "n1",
      NodeIds: []string{"n1", "n2"},
    },  
  }

  initRes, err := initClient.SendInit(ctx, initReq)
  if err != nil {
    log.Fatalf("Failed to send init request: %v", err)
  }
  log.Printf("Response to init: %s", initRes.Body.Type)

  echoReq := &echopb.EchoRequest {
    Src: "n1",
    Dest: "n2",
    Body: &echopb.EchoRequestBody {
      Type: "echo",
      MsgId: "1",
      Echo: "Hello from gRPC",
    },
  }

  echoRes, err := echoClient.SendInit(ctx, echoReq)
  if err != nil {
    log.Fatalf("Failed to send init request: %v", err)
  }
  log.Printf("Response to init: %s", echoRes.Body.Type)
}
