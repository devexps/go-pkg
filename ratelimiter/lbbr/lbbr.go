package lbbr

import (
	"fmt"
	"math"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/devexps/go-pkg/v2/ratelimiter"
	"github.com/devexps/go-pkg/v2/window"
)

type cpuGetter func() int64

// LBBR implements Bottleneck Bandwidth and Round-trip propagation time limiter.
type LBBR struct {
	cpu             cpuGetter
	passStat        window.RollingCounter
	rtStat          window.RollingCounter
	inFlight        int64
	bucketPerSecond int64
	bucketDuration  time.Duration

	// prevDropTime defines previous start drop since initTime
	prevDropTime atomic.Value
	maxPASSCache atomic.Value
	minRtCache   atomic.Value

	opts options
}

// NewLimiter returns a L-BBR limiter
func NewLimiter(opts ...Option) *LBBR {
	opt := options{
		Window:       time.Second * 10,
		Bucket:       100,
		CPUThreshold: 800,
	}
	for _, o := range opts {
		o(&opt)
	}

	bucketDuration := opt.Window / time.Duration(opt.Bucket)
	passStat := window.NewRollingCounter(window.RollingCounterOpts{Size: opt.Bucket, BucketDuration: bucketDuration})
	rtStat := window.NewRollingCounter(window.RollingCounterOpts{Size: opt.Bucket, BucketDuration: bucketDuration})

	limiter := &LBBR{
		opts:            opt,
		passStat:        passStat,
		rtStat:          rtStat,
		bucketDuration:  bucketDuration,
		bucketPerSecond: int64(time.Second / bucketDuration),
		cpu:             func() int64 { return atomic.LoadInt64(&gCPU) },
	}

	if opt.CPUQuota != 0 {
		// if cpuQuota is set, use new cpuGetter,Calculate the real CPU value based on the number of CPUs and Quota.
		limiter.cpu = func() int64 {
			return int64(float64(atomic.LoadInt64(&gCPU)) * float64(runtime.NumCPU()) / opt.CPUQuota)
		}
	}

	return limiter
}

func (l *LBBR) maxPASS() int64 {
	passCache := l.maxPASSCache.Load()
	if passCache != nil {
		ps := passCache.(*counterCache)
		if l.timespan(ps.time) < 1 {
			return ps.val
		}
	}
	rawMaxPass := int64(l.passStat.Reduce(func(iterator window.Iterator) float64 {
		var result = 1.0
		for i := 1; iterator.Next() && i < l.opts.Bucket; i++ {
			bucket := iterator.Bucket()
			count := 0.0
			for _, p := range bucket.Points {
				count += p
			}
			result = math.Max(result, count)
		}
		return result
	}))
	l.maxPASSCache.Store(&counterCache{
		val:  rawMaxPass,
		time: time.Now(),
	})
	return rawMaxPass
}

// timespan returns the passed bucket count
// since lastTime, if it is one bucket duration earlier than
// the last recorded time, it will return the BucketNum.
func (l *LBBR) timespan(lastTime time.Time) int {
	v := int(time.Since(lastTime) / l.bucketDuration)
	if v > -1 {
		return v
	}
	return l.opts.Bucket
}

func (l *LBBR) minRT() int64 {
	rtCache := l.minRtCache.Load()
	if rtCache != nil {
		rc := rtCache.(*counterCache)
		if l.timespan(rc.time) < 1 {
			return rc.val
		}
	}
	rawMinRT := int64(math.Ceil(l.rtStat.Reduce(func(iterator window.Iterator) float64 {
		var result = math.MaxFloat64
		for i := 1; iterator.Next() && i < l.opts.Bucket; i++ {
			bucket := iterator.Bucket()
			if len(bucket.Points) == 0 {
				continue
			}
			total := 0.0
			for _, p := range bucket.Points {
				total += p
			}
			avg := total / float64(bucket.Count)
			result = math.Min(result, avg)
		}
		return result
	})))
	if rawMinRT <= 0 {
		rawMinRT = 1
	}
	l.minRtCache.Store(&counterCache{
		val:  rawMinRT,
		time: time.Now(),
	})
	return rawMinRT
}

func (l *LBBR) maxInFlight() int64 {
	return int64(math.Floor(float64(l.maxPASS()*l.minRT()*l.bucketPerSecond)/1000.0) + 0.5)
}
func (l *LBBR) maxInFlight2() int64 {
	fmt.Println(l.maxPASS(), l.minRT(), l.bucketPerSecond)
	return int64(math.Floor(float64(l.maxPASS()*l.minRT()*l.bucketPerSecond)/1000.0) + 0.5)
}

func (l *LBBR) shouldDrop() bool {
	now := time.Duration(time.Now().UnixNano())
	if l.cpu() < l.opts.CPUThreshold {
		// current cpu payload below the threshold
		prevDropTime, _ := l.prevDropTime.Load().(time.Duration)
		if prevDropTime == 0 {
			// haven't start drop,
			// accept current request
			return false
		}
		if time.Duration(now-prevDropTime) <= time.Second {
			// just start drop one second ago,
			// check current inflight count
			inFlight := atomic.LoadInt64(&l.inFlight)
			return inFlight > 1 && inFlight > l.maxInFlight()
		}
		l.prevDropTime.Store(time.Duration(0))
		return false
	}
	// current cpu payload exceeds the threshold
	inFlight := atomic.LoadInt64(&l.inFlight)
	drop := inFlight > 1 && inFlight > l.maxInFlight()
	if drop {
		prevDrop, _ := l.prevDropTime.Load().(time.Duration)
		if prevDrop != 0 {
			// already started drop, return directly
			return drop
		}
		// store start drop time
		l.prevDropTime.Store(now)
	}
	return drop
}

// Stat tasks a snapshot of the L-BBR limiter.
func (l *LBBR) Stat() Stat {
	return Stat{
		CPU:         l.cpu(),
		MinRt:       l.minRT(),
		MaxPass:     l.maxPASS(),
		MaxInFlight: l.maxInFlight(),
		InFlight:    atomic.LoadInt64(&l.inFlight),
	}
}

// Allow checks all inbound traffic.
// Once overload is detected, it raises limit.ErrLimitExceed error.
func (l *LBBR) Allow() (ratelimiter.DoneFunc, error) {
	if l.shouldDrop() {
		return nil, ratelimiter.ErrLimitExceed
	}
	atomic.AddInt64(&l.inFlight, 1)
	start := time.Now().UnixNano()
	ms := float64(time.Millisecond)
	return func(ratelimiter.DoneInfo) {
		rt := int64(math.Ceil(float64(time.Now().UnixNano()-start)) / ms)
		l.rtStat.Add(rt)
		atomic.AddInt64(&l.inFlight, -1)
		l.passStat.Add(1)
	}, nil
}
