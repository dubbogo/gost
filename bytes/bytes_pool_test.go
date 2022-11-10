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
	"fmt"
	"testing"
)
import (
	"github.com/stretchr/testify/assert"
)

func Test_findIndex(t *testing.T) {
	bp := NewBytesPool([]int{16, 4 << 10, 16 << 10, 32 << 10, 64 << 10})

	type args struct {
		size int
	}
	tests := []struct {
		args args
		want int
	}{
		{args{1}, 0},
		{args{15}, 0},
		{args{16}, 0},
		{args{17}, 1},
		{args{4095}, 1},
		{args{4096}, 1},
		{args{4097}, 2},
		{args{16 << 10}, 2},
		{args{64 << 10}, 4},
		{args{64 << 11}, 5},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.args.size), func(t *testing.T) {
			if got := bp.findIndex(tt.args.size); got != tt.want {
				t.Errorf("[%v] findIndex() = %v, want %v", tt.args.size, got, tt.want)
			}
		})
	}
}

func BenchmarkAcquireBytesSize32(b *testing.B)  { benchmarkAcquireBytes(b, 32) }
func BenchmarkAcquireBytesSize10k(b *testing.B) { benchmarkAcquireBytes(b, 10000) }
func BenchmarkAcquireBytesSize60k(b *testing.B) { benchmarkAcquireBytes(b, 60000) }
func BenchmarkAcquireBytesSize70k(b *testing.B) { benchmarkAcquireBytes(b, 70000) }

func benchmarkAcquireBytes(b *testing.B, size int) {
	for i := 0; i < b.N; i++ {
		bytes := AcquireBytes(size)
		ReleaseBytes(bytes)
	}
}

func BenchmarkFindIndexSize8(b *testing.B)   { benchmarkfindIndex(b, 8) }
func BenchmarkFindIndexSize60k(b *testing.B) { benchmarkfindIndex(b, 60000) }

func benchmarkfindIndex(b *testing.B, size int) {
	for i := 0; i < b.N; i++ {
		defaultBytesPool.findIndex(size)
	}
}

func TestAcquireBytes(t *testing.T) {
	bytes := AcquireBytes(10)
	assert.Equal(t, 10, len(*bytes))
	assert.Equal(t, 512, cap(*bytes))

	bytes3 := AcquireBytes(1000000)
	assert.Equal(t, 1000000, cap(*bytes3))
	assert.Equal(t, 1000000, cap(*bytes3))
}
