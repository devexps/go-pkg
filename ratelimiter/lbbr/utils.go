package lbbr

import (
	"sync/atomic"
	"time"

	"github.com/devexps/go-pkg/v2/cpu"
)

var (
	gCPU  int64
	decay = 0.95
)

func init() {
	go cpuproc()
}

// cpu = cpuᵗ⁻¹ * decay + cpuᵗ * (1 - decay)
func cpuproc() {
	ticker := time.NewTicker(time.Millisecond * 500) // same to cpu sample rate
	defer func() {
		ticker.Stop()
		if err := recover(); err != nil {
			go cpuproc()
		}
	}()

	for range ticker.C {
		stat := &cpu.Stat{}
		cpu.ReadStat(stat)
		stat.Usage = min(stat.Usage, 1000)
		prevCPU := atomic.LoadInt64(&gCPU)
		curCPU := int64(float64(prevCPU)*decay + float64(stat.Usage)*(1.0-decay))
		atomic.StoreInt64(&gCPU, curCPU)
	}
}

func min(l, r uint64) uint64 {
	if l < r {
		return l
	}
	return r
}

// Stat contains the metrics snapshot of L-BBR.
type Stat struct {
	CPU         int64
	InFlight    int64
	MaxInFlight int64
	MinRt       int64
	MaxPass     int64
}

// counterCache is used to cache maxPASS and minRt result.
// Value of current bucket is not counted in real time.
// Cache time is equal to a bucket duration.
type counterCache struct {
	val  int64
	time time.Time
}
