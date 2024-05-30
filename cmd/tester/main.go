package main

// go build -o bin/tester cmd/tester/main.go && ./bin/tester

import (
	"context"
	"log"
	"time"
  "flag"
  "fmt"

	echopb "github.com/Shresth72/go_gRPC_tester/proto/echo"
	initpb "github.com/Shresth72/go_gRPC_tester/proto/init"
	uniqueidpb "github.com/Shresth72/go_gRPC_tester/proto/unique_ids"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RequestType int 

const (
  EchoRequest RequestType = iota
  UniqueIdsRequest
  UnknownRequest
)

func (rt RequestType) String() string {
  switch rt {
  case EchoRequest:
    return "echo"
  case UniqueIdsRequest:
    return "unique_ids"
  default:
    return "unknown"
  }
}

func main() {
  var requestTypeStr string

  flag.StringVar(&requestTypeStr, "request", "", "type of request")
  flag.Parse()

  requestType, err := parseRequestType(requestTypeStr)
  if err != nil {
    log.Fatalf("Invalid request type: %v", err)
  }

  binaryName := requestType.String() 

	conn, err := grpc.NewClient("localhost:5051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer conn.Close()

	initClient := initpb.NewInitServiceClient(conn)
	echoClient := echopb.NewEchoServiceClient(conn)
  uniqueIdsClient := uniqueidpb.NewUniqueIdsServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

  setBinaryNameReq := &initpb.SetBinaryNameRequest{
    BinaryName: binaryName,
  }

  _, err = initClient.SetBinaryName(ctx, setBinaryNameReq)
  if err != nil {
    log.Fatalf("failed to set binary name: %v", err)
  }

	initReq := &initpb.InitRequest {
		Src:  "n1",
		Dest: "n2",
		Body: &initpb.InitRequestBody {
			Type:    "init",
			NodeId:  "n1",
			NodeIds: []string{"n1", "n2"},
		},
	}

	initRes, err := initClient.SendInit(ctx, initReq)
	if err != nil {
		log.Fatalf("Failed to send init request: %v", err)
	}
	log.Printf("Response to init: %s", initRes.Body.Type)

  switch requestType {
  case EchoRequest:
    sendEchoRequest(ctx, echoClient, "hello from grpc")
    sendEchoRequest(ctx, echoClient, "this is not me")
    sendEchoRequest(ctx, echoClient, "dave mustaine")
  case UniqueIdsRequest:
    sendUniqueIdsRequest(ctx, uniqueIdsClient)
    sendUniqueIdsRequest(ctx, uniqueIdsClient)
    sendUniqueIdsRequest(ctx, uniqueIdsClient)
    sendUniqueIdsRequest(ctx, uniqueIdsClient)
    sendUniqueIdsRequest(ctx, uniqueIdsClient)
  default:
    log.Fatalf("unknown request type: %s", requestType)
  }
}

func sendEchoRequest(ctx context.Context, echoClient echopb.EchoServiceClient, echo string) {
	echoReq := &echopb.EchoRequest {
		Src:  "n1",
		Dest: "n2",
		Body: &echopb.EchoRequestBody {
			Type:  "echo",
			MsgId: 1,
			Echo: echo,
		},
	}

	echoRes, err := echoClient.SendEcho(ctx, echoReq)
	if err != nil {
		log.Fatalf("Failed to send echo request: %v", err)
	}
	log.Printf("Response to echo: %s", echoRes.Body.Type)
}

func sendUniqueIdsRequest(ctx context.Context, uniqueIdsClient uniqueidpb.UniqueIdsServiceClient) {
	uniqueIdsReq := &uniqueidpb.UniqueIdsRequest{
		Src:  "n1",
		Dest: "n2",
		Body: &uniqueidpb.UniqueIdsRequestBody{
			Type: "generate",
		},
	}

	uniqueIdsRes, err := uniqueIdsClient.SendUniqueIds(ctx, uniqueIdsReq)
	if err != nil {
		log.Fatalf("Failed to send unique IDs request: %v", err)
	}
	log.Printf("Response to unique IDs: %s", uniqueIdsRes.Body.Type)
}

func parseRequestType(requestTypeStr string) (RequestType, error) {
	switch requestTypeStr {
	case "echo":
		return EchoRequest, nil
	case "unique_ids":
		return UniqueIdsRequest, nil
	default:
		return UnknownRequest, fmt.Errorf("unknown request type: %s", requestTypeStr)
	}
}
