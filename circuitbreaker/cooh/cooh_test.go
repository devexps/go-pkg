package cooh

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getBreaker() *Breaker {
	return NewBreaker(
		WithWindow(time.Millisecond*1000),
		WithBucket(10),
		WithRequest(100),
		WithSuccess(0.5),
	).(*Breaker)
	//counterOpts := window.RollingCounterOpts{
	//	Size:           10,
	//	BucketDuration: time.Millisecond * 100,
	//}
	//stat := window.NewRollingCounter(counterOpts)
	//return &Breaker{
	//	stat: stat,
	//	r:    rand.New(rand.NewSource(time.Now().UnixNano())),
	//
	//	request: 100,
	//	k:       2,
	//	state:   StateClosed,
	//}
}

func markSuccessWithDuration(b *Breaker, count int, sleep time.Duration) {
	for i := 0; i < count; i++ {
		b.MarkSuccess()
		time.Sleep(sleep)
	}
}

func markFailedWithDuration(b *Breaker, count int, sleep time.Duration) {
	for i := 0; i < count; i++ {
		b.MarkFailed()
		time.Sleep(sleep)
	}
}

func markSuccess(b *Breaker, count int) {
	for i := 0; i < count; i++ {
		b.MarkSuccess()
	}
}

func markFailed(b *Breaker, count int) {
	for i := 0; i < count; i++ {
		b.MarkFailed()
	}
}

func testClose(t *testing.T, b *Breaker) {
	markSuccess(b, 80)
	assert.Equal(t, b.Allow(), nil)
	markSuccess(b, 120)
	assert.Equal(t, b.Allow(), nil)
}

func testOpen(t *testing.T, b *Breaker) {
	markSuccess(b, 100)
	assert.Equal(t, b.Allow(), nil)
	markFailed(b, 10000000)
	assert.NotEqual(t, b.Allow(), nil)
}

func testHalfOpen(t *testing.T, b *Breaker) {
	// failback
	assert.Equal(t, b.Allow(), nil)
	t.Run("allow single failed", func(t *testing.T) {
		markFailed(b, 10000000)
		assert.NotEqual(t, b.Allow(), nil)
	})
	time.Sleep(2 * time.Second)
	t.Run("allow single succeed", func(t *testing.T) {
		assert.Equal(t, b.Allow(), nil)
		markSuccess(b, 10000000)
		assert.Equal(t, b.Allow(), nil)
	})
}

func Test(t *testing.T) {
	b := getBreaker()
	testClose(t, b)

	b = getBreaker()
	testOpen(t, b)

	b = getBreaker()
	testHalfOpen(t, b)
}

func TestSelfProtection(t *testing.T) {
	t.Run("total request < 100", func(t *testing.T) {
		b := getBreaker()
		markFailed(b, 99)
		assert.Equal(t, b.Allow(), nil)
	})
	t.Run("total request > 100, total < 2 * success", func(t *testing.T) {
		b := getBreaker()
		size := rand.Intn(10000000)
		succ := size + 1
		markSuccess(b, succ)
		markFailed(b, size-succ)
		assert.Equal(t, b.Allow(), nil)
	})
}

func TestSummary(t *testing.T) {
	var (
		b           *Breaker
		succ, total int64
	)

	sleep := 50 * time.Millisecond
	t.Run("succ == total", func(t *testing.T) {
		b = getBreaker()
		markSuccessWithDuration(b, 10, sleep)
		succ, total = b.summary()
		assert.Equal(t, succ, int64(10))
		assert.Equal(t, total, int64(10))
	})

	t.Run("fail == total", func(t *testing.T) {
		b = getBreaker()
		markFailedWithDuration(b, 10, sleep)
		succ, total = b.summary()
		assert.Equal(t, succ, int64(0))
		assert.Equal(t, total, int64(10))
	})

	t.Run("succ = 1/2 * total, fail = 1/2 * total", func(t *testing.T) {
		b = getBreaker()
		markFailedWithDuration(b, 5, sleep)
		markSuccessWithDuration(b, 5, sleep)
		succ, total = b.summary()
		assert.Equal(t, succ, int64(5))
		assert.Equal(t, total, int64(10))
	})

	t.Run("auto reset rolling counter", func(t *testing.T) {
		time.Sleep(time.Second)
		succ, total = b.summary()
		assert.Equal(t, succ, int64(0))
		assert.Equal(t, total, int64(0))
	})
}

func TestTrueOnProba(t *testing.T) {
	const proba = math.Pi / 10
	const total = 100000
	const epsilon = 0.05
	var count int
	b := getBreaker()
	for i := 0; i < total; i++ {
		if b.trueOnProba(proba) {
			count++
		}
	}

	ratio := float64(count) / float64(total)
	assert.InEpsilon(t, proba, ratio, epsilon)
}

func BenchmarkBreakerAllow(b *testing.B) {
	breaker := getBreaker()
	b.ResetTimer()
	for i := 0; i <= b.N; i++ {
		_ = breaker.Allow()
		if i%2 == 0 {
			breaker.MarkSuccess()
		} else {
			breaker.MarkFailed()
		}
	}
}
