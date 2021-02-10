package gxetcd

import (
	"time"
)

const (
	// ConnDelay connection delay
	ConnDelay = 3
	// MaxFailTimes max failure times
	MaxFailTimes = 15
	// RegistryETCDV3Client client name
	RegistryETCDV3Client = "etcd registry"
	// MetadataETCDV3Client client name
	MetadataETCDV3Client = "etcd metadata"
)

// Options client configuration
type Options struct {
	name      string
	endpoints []string
	client    *Client
	timeout   time.Duration
	heartbeat int // heartbeat second
}

// Option will define a function of handling Options
type Option func(*Options)

// WithEndpoints sets etcd client endpoints
func WithEndpoints(endpoints ...string) Option {
	return func(opt *Options) {
		opt.endpoints = endpoints
	}
}

// WithName sets etcd client name
func WithName(name string) Option {
	return func(opt *Options) {
		opt.name = name
	}
}

// WithTimeout sets etcd client timeout
func WithTimeout(timeout time.Duration) Option {
	return func(opt *Options) {
		opt.timeout = timeout
	}
}

// WithHeartbeat sets etcd client heartbeat
func WithHeartbeat(heartbeat int) Option {
	return func(opt *Options) {
		opt.heartbeat = heartbeat
	}
}
