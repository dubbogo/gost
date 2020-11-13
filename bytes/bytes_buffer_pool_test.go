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
	"testing"
)

func TestBytesBufferPool(t *testing.T) {
	buf := GetBytesBuffer()
	bytes := []byte{0x00, 0x01, 0x02, 0x03, 0x04}
	buf.Write(bytes)
	if buf.Len() != len(bytes) {
		t.Error("iobuffer len not match write bytes' size")
	}
	PutBytesBuffer(buf)
	//buf2 := GetBytesBuffer()
	// https://go-review.googlesource.com/c/go/+/162919/
	// before go 1.13, sync.Pool just reserves some objs before every gc and will be cleanup by gc.
	// after Go 1.13, maybe there are many reserved objs after gc.
	//if buf != buf2 {
	//	t.Errorf("buf pointer %p != buf2 pointer %p", buf, buf2)
	//}
}
