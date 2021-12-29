package client

import (
	"autodiscovery/certs"
	config2 "autodiscovery/config"
	service "autodiscovery/grpc"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"time"
)

type GRPCClient struct {
	ctx    context.Context
	conn   *grpc.ClientConn
	cfg    config2.Config
	client service.SBClient
}

func NewgRPCClient(ctx context.Context, address string) (*GRPCClient, error) {

	tlsConfig := certs.GetTLSConfig()
	cred := credentials.NewTLS(tlsConfig)
	conn, err := grpc.Dial(fmt.Sprintf("%s:1818", address), grpc.WithTransportCredentials(cred))
	if err != nil {
		return nil, err
	}

	sBClient := service.NewSBClient(conn)

	return &GRPCClient{
		ctx:    ctx,
		conn:   conn,
		cfg:    config2.GetConfig(),
		client: sBClient,
	}, nil

}

func (c *GRPCClient) HeartBeat() {

	chanClient, err := c.client.LatencyChan(c.ctx, []grpc.CallOption{}...)
	if err != nil {
		panic(err)
	}

	log.Println("Starting Heartbeat...")

	for {

		if c.ctx.Err() != nil {
			return
		}

		timeStart := time.Now()

		if err := chanClient.Send(&service.Timestamp{
			SentTime:       time.Now().UnixNano(),
			SourceLocation: c.cfg.Location,
		}); err != nil {
			log.Println(err)
		}

		var ts service.Timestamp
		if err := chanClient.RecvMsg(&ts); err != nil {
			log.Printf("[Client]: Recv Heartbeat: %v\n", ts)
		}

		unixNow := time.Now().UnixNano()
		msSince := (unixNow - ts.GetSentTime()) / int64(time.Millisecond)
		log.Printf("[Client] Latency from %s->%s is %dms\n", ts.SourceLocation, c.cfg.Location, msSince)

		time.Sleep(time.Minute - time.Since(timeStart))
	}

}
