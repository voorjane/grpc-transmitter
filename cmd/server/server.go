package main

import (
	gRPC "grpc-transmitter/proto"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	gRPC.TransmitterServer
}

func (s *Server) Transmit(_ *gRPC.Request, t gRPC.Transmitter_TransmitServer) error {
	id := uuid.New().String()
	random := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	mean := -10.0 + random.Float64()*20
	sd := 0.3 + rand.Float64()*1.2
	log.Println(id, mean, sd)
	for {
		resp := &gRPC.Response{
			SessionId: id,
			Frequency: random.NormFloat64()*sd + mean,
			Timestamp: timestamppb.New(time.Now().UTC()),
		}
		err := t.Send(resp)
		if err != nil {
			return err
		}
	}
}

func main() {
	lis, err := net.Listen("tcp", "localhost:3333")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	gRPC.RegisterTransmitterServer(s, &Server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
