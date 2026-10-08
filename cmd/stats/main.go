package main

import (
	"context"
	"log"
	"net"
	statsv1 "webhook-relay/cmd/gen/stats/v1"

	"google.golang.org/grpc"
)

type server struct {
	statsv1.UnimplementedStatsServiceServer
}

func (s *server) RecordDelivery(ctx context.Context, req *statsv1.RecordDeliveryRequest) (*statsv1.RecordDeliveryResponse, error) {
	log.Println(">>> RecordDelivery called")
	log.Printf("delivery: %s success=%v latency=%dms", req.WebhookId, req.Success, req.LatencyMs)
	// TODO: записать в Postgres/Redis
	return &statsv1.RecordDeliveryResponse{}, nil
}

func (s *server) GetStats(ctx context.Context, req *statsv1.GetStatsRequest) (*statsv1.GetStatsResponse, error) {
	// TODO: прочитать из хранилища
	return &statsv1.GetStatsResponse{Total: 0}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}
	s := grpc.NewServer()
	statsv1.RegisterStatsServiceServer(s, &server{})
	log.Println("stats gRPC listening on :50051")
	log.Fatal(s.Serve(lis))
}
