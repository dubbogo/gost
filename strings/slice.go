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

import (
	"unsafe"
)

// Slice converts a string to a byte slice without memory allocation.
// It uses unsafe operations to directly reference the underlying string data,
// avoiding the overhead of copying bytes.
//
// WARNING: This function uses unsafe operations. The returned byte slice
// shares the same underlying memory as the string. Modifying the returned
// slice will cause undefined behavior since strings are immutable in Go.
//
// Parameters:
//   - s: The string to convert
//
// Returns a byte slice that references the string's underlying data.
//
// Use this function only when you need read-only access to string bytes
// with zero allocation, and you understand the safety implications.
func Slice(s string) []byte {
	ptr := unsafe.StringData(s)
	return unsafe.Slice((*byte)(ptr), len(s))
}
