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

func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}

	if reflect.ValueOf(i).IsNil() {
		return true
	}

	return false
}

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

func IsMatchPattern(pattern string, value string) bool {
	if "*" == pattern {
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
		return strings.HasSuffix(value, pattern[i+1:])
	default:
		// "*" is in the middle
		prefix := pattern[0:1]
		suffix := pattern[i+1:]
		return strings.HasPrefix(value, prefix) && strings.HasSuffix(value, suffix)
	}
}
