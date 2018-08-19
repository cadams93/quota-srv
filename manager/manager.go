package manager

import (
	"errors"
	"sync"
	"time"

	"github.com/micro/go-micro"
	"github.com/pborman/uuid"

	proto "github.com/microhq/quota-srv/proto"
	"golang.org/x/net/context"
)

type manager struct {
	sync.Mutex
	// unique id
	id string
	// key resource:bucket
	counters map[string]*counter
	// the global config
	config *proto.Config

	// alloc chan
	calloc chan *proto.Allocation

	// local allocs
	allocs map[string]*proto.Allocation
}

var (
	// publish updates on tick
	updateTick = time.Second * 5

	// idle ttl tick
	idleTick = time.Minute * 10

	// global quota with defaults
	d = newManager()

	ErrTooManyRequests = errors.New("too many requests")
	ErrServerError     = errors.New("internal server error")
)

func newManager() *manager {
	return &manager{
		id:       uuid.NewUUID().String(),
		counters: make(map[string]*counter),
		// sane default
		config: &proto.Config{
			WindowSize: int64(time.Hour.Seconds()),
			RateLimit:  10,
			IdleTtl:    int64((time.Hour * 3).Seconds()),
		},
		calloc: make(chan *proto.Allocation, 100),
		allocs: make(map[string]*proto.Allocation),
	}
}

func key(resource, bucket string) string {
	return resource + ":" + bucket
}

func (m *manager) run(c *proto.Config, p micro.Publisher) {
	// update ticker
	t := time.NewTicker(updateTick)
	defer t.Stop()

	// idle ticker
	t2 := time.NewTicker(idleTick)
	defer t2.Stop()

	for {
		select {
		// store local allocs for publishing
		case a := <-m.calloc:
			k := key(a.Resource, a.Bucket)
			al, ok := m.allocs[k]
			if !ok {
				m.allocs[k] = a
				continue
			}
			al.Total += a.Total
		// on update tick publish allocations
		case <-t.C:
			// gen update
			u := &proto.Update{
				Id:        m.id,
				Timestamp: time.Now().Unix(),
			}

			// read allocs
			for k, a := range m.allocs {
				u.Allocations = append(u.Allocations, a)
				// flush cache
				delete(m.allocs, k)
			}

			// publish updates if exist
			if len(u.Allocations) > 0 {
				p.Publish(context.Background(), u)
			}
		// on idle tick clear idle counters
		case <-t2.C:
			m.Lock()
			for k, v := range m.counters {
				// idle time in seconds
				d := int64(time.Since(v.Last()).Seconds())
				// delete idle counters
				if d > c.IdleTtl {
					delete(m.counters, k)
					v.Stop()
				}
			}
			m.Unlock()
		}
	}
}

// Allocate is doing local allocation based on a request
func (m *manager) Allocate(resource, bucket string, numAlloc int64) (int64, error) {
	m.Lock()

	// get counter
	k := key(resource, bucket)
	c, ok := m.counters[k]
	if !ok {
		c = newCounter(m.config.WindowSize)
		m.counters[k] = c
	}

	// get counts
	i := c.Get()
	t := c.Total()

	// rate limit exceeded
	if i >= m.config.RateLimit {
		m.Unlock()
		return 0, ErrTooManyRequests
	}

	// total limit exceeded
	if m.config.TotalLimit > 0 && t >= m.config.TotalLimit {
		m.Unlock()
		return 0, ErrTooManyRequests
	}

	// get bucket delta
	d := m.config.RateLimit - i

	// if requested allocation is greater than
	// the rate per second, cap the allocation
	if numAlloc > d {
		numAlloc = d
	}

	// set the counter
	c.Incr(numAlloc)

	m.Unlock()

	// update local allocs
	m.calloc <- &proto.Allocation{
		Resource:  resource,
		Bucket:    bucket,
		Total:     numAlloc,
		Timestamp: time.Now().Unix(),
	}

	return numAlloc, nil
}

// Update is processing updates from all quota services
func (m *manager) Update(u *proto.Update) {
	// skip own updates
	if u.Id == m.id {
		return
	}

	// process other updates
	m.Lock()
	for _, a := range u.Allocations {
		k := key(a.Resource, a.Bucket)

		// get counter
		c, ok := m.counters[k]
		if !ok {
			// create new counte if it does not exist
			c = newCounter(m.config.WindowSize)
			m.counters[k] = c
		}

		// increment the local counter
		c.Incrd(a.Total, a.Timestamp)
	}
	m.Unlock()

	return
}

// Start runs the quota service manager
func (m *manager) Start(c *proto.Config, p micro.Publisher) error {
	// set config
	m.Lock()
	m.config = c
	m.Unlock()

	// run manager
	go m.run(c, p)

	// success
	return nil
}

// Allocate using default manager
func Allocate(resource, bucket string, numAlloc int64) (int64, error) {
	return d.Allocate(resource, bucket, numAlloc)
}

// Update using default manager
func Update(u *proto.Update) {
	d.Update(u)
}

// Start runs the default manager
func Start(c *proto.Config, p micro.Publisher) error {
	return d.Start(c, p)
}
