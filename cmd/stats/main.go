package main

import (
	"context"
	"log"
	"net"
	statsv1 "webhook-relay/cmd/gen/stats/v1"
	"webhook-relay/internal/config"
	"webhook-relay/internal/redisx"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type server struct {
	statsv1.UnimplementedStatsServiceServer
	rdb *redisx.RedisX
}

type statsHash struct {
	Total        int64 `redis:"total"`
	Succeeded    int64 `redis:"succeeded"`
	Failed       int64 `redis:"failed"`
	LatencySumMs int64 `redis:"latency_sum_ms"`
}

func (h statsHash) avgLatencyMs() float64 {
	if h.Total == 0 {
		return 0
	}
	return float64(h.LatencySumMs) / float64(h.Total)
}

func (h statsHash) toResponse() *statsv1.GetStatsResponse {
	return &statsv1.GetStatsResponse{
		Total: h.Total, Succeeded: h.Succeeded, Failed: h.Failed, AvgLatencyMs: h.avgLatencyMs(),
	}
}
func (s *server) loadStats(ctx context.Context, webhookID string) (statsHash, error) {
	var h statsHash
	key := "stats:" + webhookID
	if err := s.rdb.HGetAll(ctx, key).Scan(&h); err != nil {
		return h, status.Errorf(codes.Internal, "read stats: %v", err)
	}
	return h, nil
}

func (s *server) StreamStats(req *statsv1.StreamStatsRequest, stream statsv1.StatsService_StreamStatsServer) error {
	ctx := stream.Context()
	channel := "deliveries" + req.WebhookId

	var ps *redis.PubSub
	if req.WebhookId != "" {
		ps = s.rdb.Subscribe(ctx, channel)
	} else {
		ps = s.rdb.PSubscribe(ctx, "deliveries:*")
	}
	defer ps.Close()

	if _, err := ps.Receive(ctx); err != nil {
		return status.Errorf(codes.Unavailable, "subscribe: %v", err)
	}

	ch := ps.Channel()

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case msg, ok := <-ch:
			if !ok {
				log.Printf("ch is close: %v", ctx)
				return ctx.Err()
			}
			var d statsv1.RecordDeliveryRequest
			if err := proto.Unmarshal([]byte(msg.Payload), &d); err != nil {
				log.Printf("bad payload: %v", err)
				continue
			}
			h, err := s.loadStats(ctx, req.WebhookId)
			if err != nil {
				log.Printf("ch is close: %v", ctx)
				continue
			}
			if err := stream.Send(&statsv1.StatsUpdate{
				Total:      h.Total,
				Succeeded:  h.Succeeded,
				Failed:     h.Failed,
				AvgLatenct: h.avgLatencyMs(),
			}); err != nil {
				log.Printf("stream send: %v", err)
				return err
			}
		}
	}
}

func (s *server) RecordDelivery(ctx context.Context, req *statsv1.RecordDeliveryRequest) (*statsv1.RecordDeliveryResponse, error) {
	log.Printf("delivery: %s success=%v latency=%dms", req.WebhookId, req.Success, req.LatencyMs)
	key := "stats:" + req.WebhookId
	msg, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	pipe := s.rdb.Pipeline()
	pipe.HIncrBy(ctx, key, "total", 1)
	if req.Success {
		pipe.HIncrBy(ctx, key, "succeeded", 1)
	} else {
		pipe.HIncrBy(ctx, key, "failed", 1)
	}
	pipe.HIncrBy(ctx, key, "latency_sum_ms", req.LatencyMs)
	pipe.Publish(ctx, "deliveries"+req.WebhookId, msg)
	if _, err := pipe.Exec(ctx); err != nil {
		log.Printf("record stats: %v", err)
	}
	return &statsv1.RecordDeliveryResponse{}, nil
}

func (s *server) GetStats(ctx context.Context, req *statsv1.GetStatsRequest) (*statsv1.GetStatsResponse, error) {
	h, err := s.loadStats(ctx, req.WebhookId)
	if err != nil {
		return nil, err
	}
	return &statsv1.GetStatsResponse{
		Total:        h.Total,
		Succeeded:    h.Succeeded,
		Failed:       h.Failed,
		AvgLatencyMs: h.avgLatencyMs(),
	}, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg, err := config.LoadConfig("../../internal/config")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	rdb, err := redisx.NewRedis(ctx, &cfg.RedisCfg)
	if err != nil {
		log.Fatalf("init redis: %v", err)
	}
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}
	s := grpc.NewServer()
	server := &server{
		rdb: rdb,
	}
	statsv1.RegisterStatsServiceServer(s, server)
	log.Println("stats gRPC listening on :50051")
	log.Fatal(s.Serve(lis))
}
