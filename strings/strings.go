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
	"reflect"
	"regexp"
	"strings"
)

// IsNil checks whether the given interface value is nil.
// It performs a deep nil check by examining both the interface itself
// and the underlying value using reflection.
// Returns true if the value is nil or if the underlying value is nil.
// For non-nilable types (e.g., int, string, struct), it returns false.
func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}

	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return v.IsNil()
	default:
		return false
	}
}

// RegSplit splits a string by a regular expression pattern.
// It uses the provided regex pattern to find all split positions in the text
// and returns a slice of substrings between those positions.
//
// Parameters:
//   - text: The string to be split
//   - regexSplit: The regular expression pattern used as delimiter
//
// Returns a slice of strings containing the split results.
//
// Example:
//
//	RegSplit("a;b;c", "\\s*[;]+\\s*") returns ["a", "b", "c"]
func RegSplit(text string, regexSplit string) []string {
	reg := regexp.MustCompile(regexSplit)
	indexes := reg.FindAllStringIndex(text, -1)
	lastStart := 0
	result := make([]string, len(indexes)+1)
	for i, element := range indexes {
		result[i] = text[lastStart:element[0]]
		lastStart = element[1]
	}
	result[len(indexes)] = text[lastStart:]
	return result
}

// IsMatchPattern is used to determine whether pattern
// and value match with wildcards currently supported *
func IsMatchPattern(pattern string, value string) bool {
	if pattern == "*" {
		return true
	}
	if len(pattern) == 0 && len(value) == 0 {
		return true
	}
	if len(pattern) == 0 || len(value) == 0 {
		return false
	}
	i := strings.LastIndex(pattern, "*")
	switch i {
	case -1:
		// doesn't find "*"
		return value == pattern
	case len(pattern) - 1:
		// "*" is at the end
		return strings.HasPrefix(value, pattern[0:i])
	case 0:
		// "*" is at the beginning
		return strings.HasSuffix(value, pattern[1:])
	default:
		// "*" is in the middle
		prefix := pattern[0:i]
		suffix := pattern[i+1:]
		return strings.HasPrefix(value, prefix) && strings.HasSuffix(value, suffix)
	}
}
