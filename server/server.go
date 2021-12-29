package server

import (
	"autodiscovery/certs"
	config2 "autodiscovery/config"
	service "autodiscovery/grpc"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"log"
	"net"
	"time"
)

type gRPCServer struct {
	ctx context.Context
	cfg config2.Config
	service.UnimplementedSBServer
}

func NewActivegRPCServer(ctx context.Context) {

	actualGRPCServer := &gRPCServer{
		cfg: config2.GetConfig(),
	}

	lis, err := net.Listen("tcp", "localhost:1818")
	if err != nil {
		log.Printf("failed to listen: %v\n", err)
		panic(err)
	}

	tlsConfig := certs.GetTLSConfig()

	grpcServer := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))
	service.RegisterSBServer(grpcServer, actualGRPCServer)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			panic(err)
		}
	}()

	select {
	case <-ctx.Done():
		grpcServer.Stop()
	}

}

func (g *gRPCServer) BuyChan(stream service.SB_BCServer) error {
	return nil
}

func (g *gRPCServer) LatencyChan(stream service.SB_LatencyChanServer) error {

	callingService := ""
	log.Println("[Server] Latency Chan Call....")

	for {
		var ts service.Timestamp
		if err := stream.RecvMsg(&ts); err != nil {
			if err == io.EOF {
				log.Printf("Caught EOF from %s - ending\n", callingService)
				return nil
			}
			return err
		}

		callingService = ts.SourceLocation
		unixNow := time.Now().UnixNano()
		msSince := (unixNow - ts.GetSentTime()) / int64(time.Millisecond)
		log.Printf("[Server] Latency from %s->%s is %dms\n", callingService, g.cfg.Location, msSince)

		if err := stream.Send(&service.Timestamp{
			SentTime:       time.Now().UnixNano(),
			SourceLocation: g.cfg.Location,
		}); err != nil {
			log.Printf("[Server] Error sending to %s: %v\n", callingService, err)
		}
	}

}
