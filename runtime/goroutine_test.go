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

package gxruntime

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestGoSafe(t *testing.T) {
	times := 1

	wg := sync.WaitGroup{}
	GoSafely(&wg,
		func() {
			panic("hello")
		},
		func(r interface{}) {
			times++
		},
	)

	wg.Wait()
	assert.True(t, times == 2)

	GoSafely(nil,
		func() {
			panic("hello")
		},
		func(r interface{}) {
			times++
		},
	)
	time.Sleep(1e9)
	assert.True(t, times == 3)
}

func TestGoUnterminal(t *testing.T) {
	times := uint64(1)
	wg := sync.WaitGroup{}
	GoUnterminal(
		&wg,
		func(){
			if atomic.AddUint64(&times, 1) == 2 {
				panic("hello")
			}
		},
	)
	wg.Wait()
	assert.True(t, times == 3)

	GoUnterminal(nil, func() {
		times++
	})
	time.Sleep(1e9)
	assert.True(t, times == 4)
}