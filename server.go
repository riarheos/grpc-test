package main

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	log "github.com/sirupsen/logrus"
	"github.com/tmc/grpc-websocket-proxy/wsproxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"io"
	"net"
	"net/http"
	pb "grpc-test/proto"
	"time"
)

type server struct {
	pb.UnimplementedGreetServer
}

func (s *server) GreetMe(ctx context.Context, req *pb.GreetRequest) (*pb.GreetResponse, error) {
	log.Printf("Received: %v", req.Name)
	return &pb.GreetResponse{
		Message: "Hello, " + req.Name,
	}, nil
}

func (s *server) GreetTimer(req *pb.GreetRequest, server pb.Greet_GreetTimerServer) error {
	log.Printf("Received stream: %v", req.Name)

	for i := 0; i < 10; i++ {
		err := server.Send(&pb.GreetResponse{
			Message: fmt.Sprintf("Hello, %v #%v", req.Name, i),
		})
		if err != nil {
			return err
		}

		time.Sleep(time.Second)
	}

	return nil
}

func (s *server) GreetUltra(server pb.Greet_GreetUltraServer) error {
	err := server.Send(&pb.GreetResponse{
		Message: "Initial server message you didn't get before",
	})
	if err != nil {
		return err
	}
	for {
		in, err := server.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		log.Printf("Received ultra: %v", in.Name)

		err = server.Send(&pb.GreetResponse{
			Message: fmt.Sprintf("Hello, %v", in.Name),
		})
		if err != nil {
			return err
		}
	}
}

func makePost (incoming *http.Request, outgoing *http.Request) *http.Request {
	outgoing.Method = "POST"
	return outgoing
}

func runProxy() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := pb.RegisterGreetHandlerFromEndpoint(ctx, mux, "localhost:11111", opts)
	if err != nil {
		return err
	}

	logger := log.New()
	logger.SetLevel(log.DebugLevel)
        log.Printf("http proxy listening at [::]:11112")
	return http.ListenAndServe("[::]:11112", mux)
}


func runWsProxy() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := pb.RegisterGreetHandlerFromEndpoint(ctx, mux, "localhost:11111", opts)
	if err != nil {
		return err
	}

	logger := log.New()
	logger.SetLevel(log.DebugLevel)
        log.Printf("ws proxy listening at [::]:11113")
	return http.ListenAndServe("[::]:11113",
		wsproxy.WebsocketProxy(
			mux,
			wsproxy.WithLogger(logger),
			wsproxy.WithRequestMutator(makePost)))
}


func main() {
	lis, err := net.Listen("tcp", "[::]:11111")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreetServer(s, &server{})
	reflection.Register(s)
	log.Printf("grpc listening at %v", lis.Addr())

	go runProxy()
	go runWsProxy()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
