// +build go1.9

package gxsync

// ref: github.com/sasha-s/go-deadlock

import "sync"

// Map is sync.Map wrapper
type Map struct {
	sync.Map
}
