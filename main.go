package main

import (
	"time"

	"github.com/micro/cli"
	"github.com/micro/go-log"
	"github.com/micro/go-micro"

	"github.com/micro/quota-srv/handler"
	"github.com/micro/quota-srv/manager"
	proto "github.com/micro/quota-srv/proto"
	"github.com/micro/quota-srv/subscriber"
)

var (
	updateTopic = "go.micro.srv.quota.update"

	// total window over which we manage quota
	windowSize = time.Hour

	// per second limit
	rateLimit = 10

	// per window limit. no limit by default
	totalLimit = 0

	// expire buckets after idle time
	idleTTL = windowSize * 2
)

func main() {
	// TODO: allow per bucket config loading
	service := micro.NewService(
		micro.Name("go.micro.srv.quota"),
		micro.Flags(
			cli.DurationFlag{
				Name:        "window_size",
				Usage:       "The window length quota is managed for. 10ms|10ns|10s|10m|10h",
				Value:       windowSize,
				Destination: &windowSize,
			},
			cli.IntFlag{
				Name:        "rate_limit",
				Usage:       "Per second rate limit",
				Value:       rateLimit,
				Destination: &rateLimit,
			},
			cli.IntFlag{
				Name:        "total_limit",
				Usage:       "Total quota limit over the window length",
				Value:       totalLimit,
				Destination: &totalLimit,
			},
			cli.DurationFlag{
				Name:        "idle_ttl",
				Usage:       "Time after which to expire idle buckets. 10ms|10ns|10s|10m|10h",
				Value:       idleTTL,
				Destination: &idleTTL,
			},
		),
	)

	service.Init(
		micro.AfterStart(func() error {
			config := &proto.Config{
				WindowSize: int64(windowSize.Seconds()),
				RateLimit:  int64(rateLimit),
				TotalLimit: int64(totalLimit),
				IdleTtl:    int64(idleTTL.Seconds()),
			}
			publisher := micro.NewPublisher(updateTopic, service.Client())
			return manager.Start(config, publisher)
		}),
	)

	// handler
	proto.RegisterQuotaHandler(service.Server(), &handler.Quota{"go.micro.srv.quota"})

	// subscriber
	micro.RegisterSubscriber(updateTopic, service.Server(), subscriber.Update)

	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
