package bootkit

import "time"

type bootkitOptions struct {
	startTimeout time.Duration
	stopTimeout  time.Duration
}

type bootkitApplyOptions struct {
	bootkit *bootkitOptions
}

type Option interface {
	apply(options *bootkitApplyOptions)
}

type startTimeoutOption time.Duration

func (t startTimeoutOption) apply(m *bootkitApplyOptions) {
	m.bootkit.startTimeout = time.Duration(t)
}

func StartTimeout(duration time.Duration) Option {
	return startTimeoutOption(duration)
}

type stopTimeoutOption time.Duration

func (t stopTimeoutOption) apply(m *bootkitApplyOptions) {
	m.bootkit.stopTimeout = time.Duration(t)
}

func StopTimeout(duration time.Duration) Option {
	return stopTimeoutOption(duration)
}
