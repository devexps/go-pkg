package lbbr

import (
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/devexps/go-pkg/v2/ratelimiter"
	"github.com/devexps/go-pkg/v2/window"

	"github.com/stretchr/testify/assert"
)

var (
	windowSizeTest   = time.Second
	bucketNumTest    = 10
	cpuThresholdTest = int64(800)

	optsForTest = []Option{
		WithWindow(windowSizeTest),
		WithBucket(bucketNumTest),
		WithCPUThreshold(cpuThresholdTest),
		WithCPUQuota(0),
	}
)

func TestAllowRate(t *testing.T) {
	limiter := NewLimiter(optsForTest...)
	var wg sync.WaitGroup
	var ok, drop int64
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				done, err := limiter.Allow()
				if err != nil {
					atomic.AddInt64(&drop, 1)
				} else {
					atomic.AddInt64(&ok, 1)
					count := rand.Intn(100)
					time.Sleep(time.Millisecond * time.Duration(count))
					done(ratelimiter.DoneInfo{})
				}
			}
		}()
	}
	wg.Wait()
	t.Log("drop: ", drop)
	t.Log("ok: ", ok)
}

func TestMaxPass(t *testing.T) {
	bucketDuration := windowSizeTest / time.Duration(bucketNumTest)
	limiter := NewLimiter(optsForTest...)
	for i := 1; i <= 10; i++ {
		limiter.passStat.Add(int64(i * 100))
		time.Sleep(bucketDuration)
	}
	assert.Equal(t, int64(1000), limiter.maxPASS())

	// default max pass is equal to 1.
	limiter = NewLimiter(optsForTest...)
	assert.Equal(t, int64(1), limiter.maxPASS())
}

func TestMaxPassWithCache(t *testing.T) {
	bucketDuration := windowSizeTest / time.Duration(bucketNumTest)
	limiter := NewLimiter(optsForTest...)
	// witch cache, value of the latest bucket is not counted instantly.
	// after a bucket duration time, this bucket will be fully counted.
	limiter.passStat.Add(int64(50))
	time.Sleep(bucketDuration / 2)
	assert.Equal(t, int64(1), limiter.maxPASS())

	limiter.passStat.Add(int64(50))
	time.Sleep(bucketDuration / 2)
	assert.Equal(t, int64(1), limiter.maxPASS())

	limiter.passStat.Add(int64(1))
	time.Sleep(bucketDuration)
	assert.Equal(t, int64(100), limiter.maxPASS())
}

func TestMinRt(t *testing.T) {
	bucketDuration := windowSizeTest / time.Duration(bucketNumTest)
	limiter := NewLimiter(optsForTest...)
	for i := 0; i < 10; i++ {
		for j := i*10 + 1; j <= i*10+10; j++ {
			limiter.rtStat.Add(int64(j))
		}
		if i != 9 {
			time.Sleep(bucketDuration)
		}
	}
	assert.Equal(t, int64(6), limiter.minRT())

	// default max min rt is equal to maxFloat64.
	limiter = NewLimiter(optsForTest...)
	limiter.rtStat = window.NewRollingCounter(window.RollingCounterOpts{Size: 10, BucketDuration: bucketDuration})
	assert.Equal(t, int64(math.Ceil(math.MaxFloat64)), limiter.minRT())
}

func TestMinRtWithCache(t *testing.T) {
	bucketDuration := windowSizeTest / time.Duration(bucketNumTest)
	limiter := NewLimiter(optsForTest...)
	for i := 0; i < 10; i++ {
		for j := i*10 + 1; j <= i*10+5; j++ {
			limiter.rtStat.Add(int64(j))
		}
		if i != 9 {
			time.Sleep(bucketDuration / 2)
		}
		_ = limiter.minRT()
		for j := i*10 + 6; j <= i*10+10; j++ {
			limiter.rtStat.Add(int64(j))
		}
		if i != 9 {
			time.Sleep(bucketDuration / 2)
		}
	}
	assert.Equal(t, int64(6), limiter.minRT())
}

func TestMaxQps(t *testing.T) {
	limiter := NewLimiter(optsForTest...)
	bucketDuration := windowSizeTest / time.Duration(bucketNumTest)
	passStat := window.NewRollingCounter(window.RollingCounterOpts{Size: 10, BucketDuration: bucketDuration})
	rtStat := window.NewRollingCounter(window.RollingCounterOpts{Size: 10, BucketDuration: bucketDuration})
	for i := 0; i < 10; i++ {
		passStat.Add(int64((i + 2) * 100))
		for j := i*10 + 1; j <= i*10+10; j++ {
			rtStat.Add(int64(j))
		}
		if i != 9 {
			time.Sleep(bucketDuration)
		}
	}
	limiter.passStat = passStat
	limiter.rtStat = rtStat
	assert.Equal(t, int64(60), limiter.maxInFlight())
}

func TestShouldDrop(t *testing.T) {
	var cpu int64
	limiter := NewLimiter(optsForTest...)
	limiter.cpu = func() int64 {
		return cpu
	}
	bucketDuration := windowSizeTest / time.Duration(bucketNumTest)
	passStat := window.NewRollingCounter(window.RollingCounterOpts{Size: 10, BucketDuration: bucketDuration})
	rtStat := window.NewRollingCounter(window.RollingCounterOpts{Size: 10, BucketDuration: bucketDuration})
	for i := 0; i < 10; i++ {
		passStat.Add(int64((i + 1) * 100))
		for j := i*10 + 1; j <= i*10+10; j++ {
			rtStat.Add(int64(j))
		}
		if i != 9 {
			time.Sleep(bucketDuration)
		}
	}
	limiter.passStat = passStat
	limiter.rtStat = rtStat

	// cpu >=  800, inflight < maxQps
	cpu = 800
	limiter.inFlight = 50
	assert.Equal(t, false, limiter.shouldDrop())

	// cpu >=  800, inflight > maxQps
	cpu = 800
	limiter.inFlight = 80
	assert.Equal(t, true, limiter.shouldDrop())

	// cpu < 800, inflight > maxQps, cold duration
	cpu = 700
	limiter.inFlight = 80
	assert.Equal(t, true, limiter.shouldDrop())

	// cpu < 800, inflight > maxQps
	time.Sleep(2 * time.Second)
	cpu = 700
	limiter.inFlight = 80
	assert.Equal(t, false, limiter.shouldDrop())
}

func BenchmarkAllowUnderLowLoad(b *testing.B) {
	limiter := NewLimiter(optsForTest...)
	limiter.cpu = func() int64 {
		return 500
	}
	b.ResetTimer()
	for i := 0; i <= b.N; i++ {
		done, err := limiter.Allow()
		if err == nil {
			done(ratelimiter.DoneInfo{})
		}
	}
}
