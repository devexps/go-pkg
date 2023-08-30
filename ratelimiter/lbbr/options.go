package lbbr

import "time"

// Option function for L-BBR limiter
type Option func(*options)

// options of bbr limiter.
type options struct {
	// WindowSize defines time duration per window
	Window time.Duration
	// BucketNum defines bucket number for each window
	Bucket int
	// CPUThreshold
	CPUThreshold int64
	// CPUQuota
	CPUQuota float64
}

// WithWindow with window size.
func WithWindow(d time.Duration) Option {
	return func(o *options) {
		o.Window = d
	}
}

// WithBucket with bucket ize.
func WithBucket(b int) Option {
	return func(o *options) {
		o.Bucket = b
	}
}

// WithCPUThreshold with cpu threshold;
func WithCPUThreshold(threshold int64) Option {
	return func(o *options) {
		o.CPUThreshold = threshold
	}
}

// WithCPUQuota with real cpu quota(if it can not collect from process correct);
func WithCPUQuota(quota float64) Option {
	return func(o *options) {
		o.CPUQuota = quota
	}
}
