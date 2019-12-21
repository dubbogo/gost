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

package gxbytes

import (
	"errors"
	"sync"
)

var ErrSizeTooLarge = errors.New(`acquired size is too large`)

var (
	// bufPoolSize declare the cap of each pool
	bufPoolSize = []int{16, 1 << 10, 2 << 10, 4 << 10, 8 << 10, 16 << 10, 32 << 10, 64 << 10}

	// bufPools store pools
	bufPools []sync.Pool

	// bufPoolLen calc length of pools at init
	bufPoolLen = len(bufPoolSize)
)

func init() {
	InitPool(bufPoolSize)
}

// InitPool must called once
func InitPool(poolSize []int) {
	bufPoolSize = poolSize
	bufPoolLen = len(bufPoolSize)

	bufPools = make([]sync.Pool, len(bufPoolSize))
	for i, size := range bufPoolSize {
		size := size
		bufPools[i] = sync.Pool{New: func() interface{} {
			return make([]byte, 0, size)
		}}
	}
}

func findIndex(size int) int {
	for i := 0; i < bufPoolLen; i++ {
		if bufPoolSize[i] >= size {
			return i
		}
	}
	return bufPoolLen
}

func AcquireBytes(size int) ([]byte, error) {
	idx := findIndex(size)
	if idx >= bufPoolLen {
		return make([]byte, 0, size), ErrSizeTooLarge
	}

	return bufPools[idx].Get().([]byte)[:size], nil
}

func ReleaseBytes(buf []byte) error {
	idx := findIndex(cap(buf))
	if idx >= bufPoolLen {
		return ErrSizeTooLarge
	}
	bufPools[idx].Put(buf)
	return nil
}
