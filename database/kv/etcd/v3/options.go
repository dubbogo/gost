/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gxetcd

import (
	"time"
)

const (
	// ConnDelay connection delay
	ConnDelay = 3
	// MaxFailTimes max failure times
	MaxFailTimes = 15
	// RegistryETCDV3Client Client Name
	RegistryETCDV3Client = "etcd registry"
	// MetadataETCDV3Client Client Name
	MetadataETCDV3Client = "etcd metadata"
)

// Options Client configuration
type Options struct {
	//Name etcd server Name
	Name string
	//Endpoints etcd Endpoints
	Endpoints []string
	//Client etcd Client
	Client *Client
	//Timeout Timeout
	Timeout time.Duration
	//Heartbeat
	Heartbeat int // Heartbeat second
}

// Option will define a function of handling Options
type Option func(*Options)

// WithEndpoints sets etcd Client Endpoints
func WithEndpoints(endpoints ...string) Option {
	return func(opt *Options) {
		opt.Endpoints = endpoints
	}
}

// WithName sets etcd Client Name
func WithName(name string) Option {
	return func(opt *Options) {
		opt.Name = name
	}
}

// WithTimeout sets etcd Client Timeout
func WithTimeout(timeout time.Duration) Option {
	return func(opt *Options) {
		opt.Timeout = timeout
	}
}

// WithHeartbeat sets etcd Client Heartbeat
func WithHeartbeat(heartbeat int) Option {
	return func(opt *Options) {
		opt.Heartbeat = heartbeat
	}
}
