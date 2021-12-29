package discovery

import (
	client "autodiscovery/client"
	config2 "autodiscovery/config"
	"autodiscovery/types"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"os"
	"sync"
	"time"
)

type Discovery struct {
	ctx   context.Context
	cfg   config2.Config
	redis *redis.Client

	meshClients map[string]*client.GRPCClient

	sync.Mutex
}

func NewDiscovery(ctx context.Context) *Discovery {

	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("redis_addr"),
		Password: os.Getenv("redis_password"),
		DB:       0,
	})

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	discovery := &Discovery{
		ctx:         ctx,
		cfg:         config2.GetConfig(),
		redis:       redisClient,
		meshClients: make(map[string]*client.GRPCClient),
	}

	return discovery

}

func (d *Discovery) DiscoveryListener() {
	pubSub := d.redis.Subscribe(d.ctx, "discovery")
	pubSubChan := pubSub.Channel()

	for {
		select {
		case msg := <-pubSubChan:
			d.DiscoveryMessageHandler([]byte(msg.Payload))
		case <-d.ctx.Done():
			pubSub.Close()
			return
		}
	}
}

func (d *Discovery) DiscoveryMessageHandler(message []byte) {

	var publishedMessage types.Publisher
	if err := json.Unmarshal(message, &publishedMessage); err != nil {
		return
	}

	msLatency := (time.Now().UnixNano() - publishedMessage.TimeNow) / int64(time.Millisecond)
	fmt.Printf("Discovered %s:%s - latency .%dms\n", publishedMessage.AssignedName, publishedMessage.IPAddress, msLatency)

	//if publishedMessage.AssignedName == d.cfg.Location {
	//	return
	//}

	clientName := publishedMessage.AssignedName
	_, exists := d.meshClients[clientName]
	if !exists {
		fmt.Printf("[Discovery] Client %s doesn't have entry... creating\n", clientName)
		newClient, err := client.NewgRPCClient(d.ctx, publishedMessage.IPAddress)
		if err != nil {
			fmt.Printf("[Discovery] Error creating client %s: %v\n", clientName, err)
			return
		}

		d.meshClients[clientName] = newClient
		go d.meshClients[clientName].HeartBeat()
	}

}

func (d *Discovery) PublishInfoTicker() {

	ticker := time.NewTicker(time.Minute)
	d.PublishInfo()

	for {
		select {
		case <-ticker.C:
			d.PublishInfo()
		case <-d.ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (d *Discovery) PublishInfo() {

	discoveryInterface := types.Publisher{
		TimeNow:      time.Now().UnixNano(),
		AssignedName: d.cfg.Location,
		IPAddress:    d.cfg.IPAddress,
	}

	discoveryInterfaceBytes, _ := json.Marshal(discoveryInterface)

	d.redis.Publish(d.ctx, "discovery", discoveryInterfaceBytes)
}
