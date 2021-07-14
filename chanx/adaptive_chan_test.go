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

package chanx

import (
	"sync"
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestAdaptiveChan(t *testing.T) {
	ch := NewAdaptiveChan(100, 100, 100)

	var count int

	for i := 1; i < 200; i++ {
		ch.In() <- i
	}

	for i := 1; i < 60; i++ {
		v, _ := <-ch.Out()
		count += v.(int)
	}

	assert.Equal(t, 100, ch.buffer.Cap())

	for i := 200; i <= 1200; i++ {
		ch.In() <- i
	}
	assert.Equal(t, 1600, ch.buffer.Cap())

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for v := range ch.Out() {
			count += v.(int)
		}
	}()

	close(ch.In())

	wg.Wait()

	assert.Equal(t, 720600, count)

}
