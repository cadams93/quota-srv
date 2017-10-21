package manager

import (
	"sync"
	"time"
)

type counter struct {
	// counter window in seconds
	window int64

	sync.RWMutex
	buckets []*bucket

	// exit chan
	exit chan bool

	// time of last update
	last time.Time
}

type bucket struct {
	c     int64
	first time.Time
	last  time.Time
}

func newBucket() *bucket {
	return &bucket{
		c:     0,
		first: time.Now(),
		last:  time.Now(),
	}
}

func newCounter(window int64) *counter {
	c := &counter{
		buckets: []*bucket{newBucket()},
		window:  window,
		exit:    make(chan bool),
		last:    time.Now(),
	}
	go c.run()
	return c
}

func (c *counter) run() {
	t := time.NewTicker(updateTick)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			c.slide()
		case <-c.exit:
			return
		}
	}
}

func (c *counter) slide() {
	c.Lock()

	// now
	n := time.Now()

	// new bucket
	c.buckets = append(c.buckets, newBucket())

	for _, b := range c.buckets {
		// expire buckets older than the window
		if int64(n.Sub(b.last).Seconds()) > c.window {
			c.buckets = c.buckets[1:]
		}
	}

	c.Unlock()
}

func (c *counter) Get() int64 {
	c.RLock()
	defer c.RUnlock()

	// current bucket
	b := c.buckets[len(c.buckets)-1]

	// the bucket is older than a second
	// return 0
	if time.Since(b.first) > time.Second {
		return 0
	}

	// return the count
	return b.c
}

func (c *counter) Total() int64 {
	c.RLock()
	defer c.RUnlock()

	var t int64

	// now
	n := time.Now()

	// get total
	for _, b := range c.buckets {
		// skip anything outside the window
		if int64(n.Sub(b.last).Seconds()) > c.window {
			continue
		}
		// add bucket to total
		t += b.c
	}

	// return total
	return t
}

func (c *counter) Incr(d int64) {
	c.Lock()

	b := c.buckets[len(c.buckets)-1]

	// if the bucket is older than a second create a new one
	if time.Since(b.first) > time.Second {
		b = newBucket()
		c.buckets = append(c.buckets, b)
	}
	// increment bucket
	b.c += d

	// set update time
	b.last = time.Now()
	c.last = time.Now()

	c.Unlock()
}

// Increment at timestamp
func (c *counter) Incrd(d, t int64) {
	c.Lock()
	// reverse loop
	for idx := len(c.buckets) - 1; idx >= 0; idx-- {
		b := c.buckets[idx]
		if t >= b.first.Unix() {
			b.c += d
			break
		}
	}
	c.Unlock()
}

func (c *counter) Last() time.Time {
	c.RLock()
	defer c.RUnlock()
	return c.last
}

func (c *counter) Stop() {
	select {
	case <-c.exit:
		return
	default:
		close(c.exit)
	}
}
