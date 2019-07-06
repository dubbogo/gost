package gxsync

// ref: github.com/sasha-s/go-deadlock

func EnableDeadlock(enable bool) {
	Opts.Disable = true
	Opts.DisableLockOrderDetection = true

	if enable {
		Opts.Disable = false
		Opts.DisableLockOrderDetection = false
	}
}
