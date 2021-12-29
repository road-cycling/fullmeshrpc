package main

// protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative grpc/service.proto

import (
	"autodiscovery/discovery"
	"autodiscovery/server"
	"context"
	"fmt"
	"time"
)

func main() {
	fmt.Println("hello world")

	rootContext, rootCtxCancel := context.WithCancel(context.Background())
	defer rootCtxCancel()

	go server.NewActivegRPCServer(rootContext)

	d := discovery.NewDiscovery(context.Background())
	go d.DiscoveryListener()
	d.PublishInfoTicker()

	time.Sleep(time.Hour)

}
