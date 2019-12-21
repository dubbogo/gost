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

var ErrSizeTooLarge = errors.New(`size is too large`)

type BytesPool struct {
	sizes  []int // sizes declare the cap of each slot
	slots  []sync.Pool
	length int
}

var defaultBytesPool = NewBytesPool([]int{16, 1 << 10, 2 << 10, 4 << 10, 8 << 10, 16 << 10, 32 << 10, 64 << 10})

// NewBytesPool
func NewBytesPool(slotSize []int) *BytesPool {
	bp := &BytesPool{}
	bp.sizes = slotSize
	bp.length = len(bp.sizes)

	bp.slots = make([]sync.Pool, bp.length)
	for i, size := range bp.sizes {
		size := size
		bp.slots[i] = sync.Pool{New: func() interface{} {
			return make([]byte, 0, size)
		}}
	}
	return bp
}

func SetDefaultBytesPool(bp *BytesPool) {
	defaultBytesPool = bp
}

func (bp *BytesPool) findIndex(size int) int {
	for i := 0; i < bp.length; i++ {
		if bp.sizes[i] >= size {
			return i
		}
	}
	return bp.length
}

func (bp *BytesPool) AcquireBytes(size int) ([]byte, error) {
	idx := bp.findIndex(size)
	if idx >= bp.length {
		return make([]byte, 0, size), ErrSizeTooLarge
	}

	return bp.slots[idx].Get().([]byte)[:size], nil
}

func (bp *BytesPool) ReleaseBytes(buf []byte) error {
	idx := bp.findIndex(cap(buf))
	if idx >= bp.length {
		return ErrSizeTooLarge
	}

	bp.slots[idx].Put(buf)
	return nil
}

func AcquireBytes(size int) ([]byte, error) { return defaultBytesPool.AcquireBytes(size) }

func ReleaseBytes(buf []byte) error { return defaultBytesPool.ReleaseBytes(buf) }
