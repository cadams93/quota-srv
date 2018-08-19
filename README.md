# Quota SRV

The quota service maintains a quota for any resource in a global and distributed fashion.

This is useful for building rate limiting or circuit breaking features across a platform.

## How it works

The service provides sane defaults with dynamic creation of resource buckets. The quota service 
has zero external dependencies, maintaining an in memory structure and publishing updates 
between instances using the go-micro broker to distribute its state.

The quota service uses "resource" and "bucket" as a unique key pair. Anything matching this will be stored in the same bucket.

Examples

```
resource: service.Hello.World
bucket: from.my.service
```

```
resource: http://localhost:8080/foo/bar
bucket: ip:10.0.10.1
```

## How it's used

Before a client makes a request to a resource:

1. A request is made to the quota service for an allocation
2. If quota is available it is allocated. If the rate or total limited is exceeded an error is returned
3. The client carries out the request if it has allocation or backs off it an error occurred

This can be used to throttle RPC requests and control flow for a service.

The quota service should be used in conjuction with go-micro wrappers.

## Usage

Prerequisite: Run service discovery

Get service 

```
go get github.com/microhq/quota-srv
```

Run service

```
quota-srv
```

Call Quota.Allocate to request an allocation

```
micro query go.micro.srv.quota Quota.Allocate '{"resource": "foo", "bucket": "bar", "allocation": 1}'
```

Returns an allocation

```
{
	"allocation": 1
}
```

### Flags

Specify flags to set the window size, rate limit, total limit and idle ttl.

```
--window_size # The window length quota is managed for. 10ms|10ns|10s|10m|10h
```

```
--rate_limit # Per second rate limit
```

```
--total_limit # Total quota limit over the window length
```

```
--idle_ttl # Time after which to expire idle buckets. 10ms|10ns|10s|10m|10h
```

### Wrappers

A client wrapper will call the quota service before every request to make an allocation. This is a naive implementation. Ideally it should be performed 
in the background.

```
import (
	"github.com/micro/go-micro"
	"github.com/micro/go-plugins/wrapper/ratelimiter/quota"
)

srv := micro.NewService(
	micro.WrapClient(
		quota.NewClientWrapper("my.service"),
	),
)
```

### API Plugin

TODO

## TODO

- [ ] Per bucket config loading
- [ ] Pluggable shared data structure
