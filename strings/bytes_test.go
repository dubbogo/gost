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

package gxstrings

import "testing"

/*
type sbytes struct {
	string
	Cap int
}

// StringToBytes converts a string into bytes without mem-allocs.
func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&sbytes{s, len(s)}))
}

if the StringToBytes is implemented as above, the test result is as follows.

// const str = "hello, world" // 11Bytes
goos: darwin
goarch: amd64
pkg: github.com/dubbogo/gost/strings
BenchmarkStringToBytes
BenchmarkStringToBytes-12    	1000000000	         0.279 ns/op

// const str = "hello, world" // 11Bytes
goos: darwin
goarch: amd64
pkg: github.com/dubbogo/gost/strings
BenchmarkStringToBytesRaw
BenchmarkStringToBytesRaw-12    	1000000000	         0.253 ns/op

// const str = "hello, world, hello, world, helL" // 24 Bytes
goos: darwin
goarch: amd64
pkg: github.com/dubbogo/gost/strings
BenchmarkStringToBytes
BenchmarkStringToBytes-12    	1000000000	         0.256 ns/op

// const str = "hello, world, hello, world, helL" // 24 Bytes
goos: darwin
goarch: amd64
pkg: github.com/dubbogo/gost/strings
BenchmarkStringToBytesRaw
BenchmarkStringToBytesRaw-12    	1000000000	         0.253 ns/op

const str = "hello, world, hello, world, hello, world" // 40 Bytes
goos: darwin
goarch: amd64
pkg: github.com/dubbogo/gost/strings
BenchmarkStringToBytes
BenchmarkStringToBytes-12    	1000000000	         0.254 ns/op

const str = "hello, world, hello, world, hello, world" // 40 Bytes
goos: darwin
goarch: amd64
pkg: github.com/dubbogo/gost/strings
BenchmarkStringToBytesRaw
BenchmarkStringToBytesRaw-12    	1000000000	         0.269 ns/op

// const str = "dfjaslfjaldsfjalsdfjdl;fjadlfjal;sdfjl;adsfjladsfkadssdfadsfadsf" // 64bytes
goos: darwin
goarch: amd64
pkg: github.com/dubbogo/gost/strings
BenchmarkStringToBytes
BenchmarkStringToBytes-12    	1000000000	         0.280 ns/op

// const str = "dfjaslfjaldsfjalsdfjdl;fjadlfjal;sdfjl;adsfjladsfkadssdfadsfadsf" // 64bytes
goos: darwin
goarch: amd64
pkg: github.com/dubbogo/gost/strings
BenchmarkStringToBytesRaw
BenchmarkStringToBytesRaw-12    	39803451	        29.9 ns/op

the summary is:

if the string length is less than 24 bytes, the shorter the string, the better the performance of `[]byte(string)`.
if the string length is greater than 24 bytes, the longer the string, the better the performance of `StringToBytes`.
*/

const str = "dfjaslfjaldsfjalsdfjdl;fjadlfjal;sdfjl;adsfjladsfkadssdfadsfadsf" // 64bytes

func BenchmarkStringToBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		StringToBytes(str)
	}
}

func BenchmarkStringToBytesRaw(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = []byte(str)
	}
}
